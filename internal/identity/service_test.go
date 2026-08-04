package identity

import (
	"testing"
	"time"

	"github.com/pquerna/otp/totp"
	maildomain "github.com/wavefnd/wave-platform/internal/mail"
	"github.com/wavefnd/wave-platform/internal/permission"
	"github.com/wavefnd/wave-platform/internal/storage"
)

const testTOTPSecret = "JBSWY3DPEHPK3PXPJBSWY3DPEHPK3PXP"

func newTestIdentity(t *testing.T, database *storage.Database) *Service {
	t.Helper()
	service, err := NewServiceWithTOTP(database, "wave-lang.dev", true, time.Hour,
		"MDEyMzQ1Njc4OWFiY2RlZjAxMjM0NTY3ODlhYmNkZWY=", "Wave Test", "http://localhost")
	if err != nil {
		t.Fatal(err)
	}
	return service
}

func testRegistration(displayName string) Registration {
	return Registration{DisplayName: displayName, TOTPSecret: testTOTPSecret, RecoveryEmail: "recovery@example.net"}
}

func TestRegisterCreatesAddressAuthenticatorMailboxAndSession(t *testing.T) {
	database, err := storage.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	service := newTestIdentity(t, database)

	item, err := service.Register(testRegistration("John Mark"))
	if err != nil {
		t.Fatal(err)
	}
	if item.Username != "john-mark" || item.Email != "john-mark@wave-lang.dev" {
		t.Fatalf("account = %#v", item)
	}
	box, err := service.Mailbox(item.ID)
	if err != nil {
		t.Fatal(err)
	}
	if box.Address != item.Email || box.AccountID != item.ID {
		t.Fatalf("mailbox = %#v", box)
	}

	factor, err := service.SecurityStatus(item.ID)
	if err != nil || factor.RecoveryEmail != "recovery@example.net" {
		t.Fatal(err)
	}
	code, err := totp.GenerateCode(testTOTPSecret, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	authenticated, token, currentSession, err := service.AuthenticateTOTP(item.Email, code, "test")
	if err != nil {
		t.Fatal(err)
	}
	if authenticated.ID != item.ID || token == "" || currentSession.AccountID != item.ID {
		t.Fatal("session was not created")
	}
}

func TestExternalMailIsQueuedWithoutCreatingARecipientMailboxEntry(t *testing.T) {
	database, err := storage.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	service := newTestIdentity(t, database)
	sender, err := service.Register(testRegistration("Ada Lovelace"))
	if err != nil {
		t.Fatal(err)
	}

	sent, err := service.SendMail(sender, OutgoingMail{To: "person@example.net", Subject: "External mail", Body: "Queued body."})
	if err != nil {
		t.Fatal(err)
	}
	if sent.DeliveryStatus != "queued" {
		t.Fatalf("delivery status = %q", sent.DeliveryStatus)
	}
	deliveries, err := maildomain.NewRepository(database).DeliveriesByMessage(sent.Message.ID)
	if err != nil || len(deliveries) != 1 {
		t.Fatalf("deliveries=%d err=%v", len(deliveries), err)
	}
	if deliveries[0].Destination != "remote-smtp" || deliveries[0].Recipient != "person@example.net" {
		t.Fatalf("delivery = %#v", deliveries[0])
	}
	items, err := service.MailboxItems(sender.ID, "Sent")
	if err != nil || len(items) != 1 {
		t.Fatalf("sent items=%d err=%v", len(items), err)
	}
}

func TestDuplicateDisplayNameGetsUniqueAddress(t *testing.T) {
	database, err := storage.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	service := newTestIdentity(t, database)
	first, err := service.Register(testRegistration("John Mark"))
	if err != nil {
		t.Fatal(err)
	}
	second, err := service.Register(testRegistration("John Mark"))
	if err != nil {
		t.Fatal(err)
	}
	if first.Email != "john-mark@wave-lang.dev" || second.Email != "john-mark-2@wave-lang.dev" {
		t.Fatalf("addresses = %q, %q", first.Email, second.Email)
	}
}

func TestBootstrapTOTPAdminIsIdempotent(t *testing.T) {
	database, err := storage.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	service := newTestIdentity(t, database)
	first, created, err := service.BootstrapTOTPAdmin("Wave Administrator", "wave-admin", "owner@example.net", testTOTPSecret)
	if err != nil || !created {
		t.Fatalf("first bootstrap: created=%v err=%v", created, err)
	}
	second, created, err := service.BootstrapTOTPAdmin("Wave Administrator", "wave-admin", "owner@example.net", testTOTPSecret)
	if err != nil || created || first.ID != second.ID {
		t.Fatalf("second bootstrap: created=%v err=%v", created, err)
	}
	if !service.IsOwner(first.ID) || !service.IsAdministrator(first.ID) {
		t.Fatal("owner role was not assigned")
	}
}

func TestPlatformAdministratorCanManageWithoutBecomingOwner(t *testing.T) {
	database, err := storage.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	service := newTestIdentity(t, database)
	item, err := service.Register(testRegistration("Operations Admin"))
	if err != nil {
		t.Fatal(err)
	}
	permissions := permission.NewRepository(database)
	if err := permissions.PutRole(permission.Role{ID: "platform-admin", Name: "Platform administrator", Permissions: []string{"platform.admin.*"}}); err != nil {
		t.Fatal(err)
	}
	if err := permissions.Assign(permission.Assignment{AccountID: item.ID, RoleID: "platform-admin", Scope: "platform"}); err != nil {
		t.Fatal(err)
	}
	if !service.IsAdministrator(item.ID) || service.IsOwner(item.ID) {
		t.Fatal("administrator and owner roles were not kept separate")
	}
}

func TestSendMailUsesOneMessageForSenderAndRecipient(t *testing.T) {
	database, err := storage.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	service := newTestIdentity(t, database)
	sender, err := service.Register(testRegistration("Ada Lovelace"))
	if err != nil {
		t.Fatal(err)
	}
	recipient, err := service.Register(testRegistration("Grace Hopper"))
	if err != nil {
		t.Fatal(err)
	}

	sent, err := service.SendMail(sender, OutgoingMail{To: recipient.Email, Subject: "Wave mail test", Body: "The message body."})
	if err != nil {
		t.Fatal(err)
	}
	senderItems, err := service.MailboxItems(sender.ID, "Sent")
	if err != nil {
		t.Fatal(err)
	}
	recipientItems, err := service.MailboxItems(recipient.ID, "Inbox")
	if err != nil {
		t.Fatal(err)
	}
	if len(senderItems) != 1 || len(recipientItems) != 1 {
		t.Fatalf("sent=%d inbox=%d", len(senderItems), len(recipientItems))
	}
	if senderItems[0].Message.ID != sent.Message.ID || recipientItems[0].Message.ID != sent.Message.ID {
		t.Fatal("sender and recipient entries do not reference the same mail message")
	}
	if recipientItems[0].Body != "The message body." || !containsFlag(recipientItems[0].Entry.Flags, "unread") {
		t.Fatalf("recipient item = %#v", recipientItems[0])
	}

	updated, err := service.UpdateMailboxEntry(recipient.ID, recipientItems[0].Entry.ID, "archive")
	if err != nil {
		t.Fatal(err)
	}
	if updated.Entry.Folder != "Archive" {
		t.Fatalf("folder = %q", updated.Entry.Folder)
	}
	archived, err := service.MailboxItems(recipient.ID, "Archive")
	if err != nil || len(archived) != 1 {
		t.Fatalf("archived=%d err=%v", len(archived), err)
	}
}

func containsFlag(flags []string, target string) bool {
	for _, flag := range flags {
		if flag == target {
			return true
		}
	}
	return false
}
