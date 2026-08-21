package identity

import (
	"bytes"
	"errors"
	"io"
	"mime"
	"mime/multipart"
	"mime/quotedprintable"
	stdmail "net/mail"
	"strings"
	"testing"
	"time"

	"github.com/pquerna/otp/totp"
	maildomain "github.com/wavefnd/wave-platform/internal/mail"
	patchdomain "github.com/wavefnd/wave-platform/internal/patcharchive"
	"github.com/wavefnd/wave-platform/internal/permission"
	"github.com/wavefnd/wave-platform/internal/storage"
	webhookdomain "github.com/wavefnd/wave-platform/internal/webhook"
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

func TestCompletedRegistrationReceivesFounderWelcomeMail(t *testing.T) {
	database, err := storage.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	service := newTestIdentity(t, database)
	founder, created, err := service.BootstrapTOTPAdmin("LunaStev", "lunastev", "founder@example.net", testTOTPSecret)
	if err != nil || !created {
		t.Fatalf("bootstrap founder: created=%v err=%v", created, err)
	}
	enrollment, err := service.BeginTOTPRegistration("New Member", "", "member@example.net")
	if err != nil {
		t.Fatal(err)
	}
	code, err := totp.GenerateCode(enrollment.Secret, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	member, _, _, err := service.CompleteTOTPRegistration(enrollment.Token, code, "registration-test")
	if err != nil {
		t.Fatal(err)
	}
	inbox, err := service.MailboxItems(member.ID, "Inbox")
	if err != nil || len(inbox) != 1 {
		t.Fatalf("welcome inbox=%#v err=%v", inbox, err)
	}
	welcome := inbox[0]
	if welcome.Message.Subject != "Welcome to the Wave community" || welcome.Message.AuthorAccountID != founder.ID ||
		!strings.Contains(welcome.Body, "https://opencollective.com/wave-lang") || !strings.Contains(welcome.Body, "Hi New Member,") {
		t.Fatalf("welcome message=%#v body=%q", welcome.Message, welcome.Body)
	}
	if welcome.DeliveryStatus != "delivered" || !containsFlag(welcome.Entry.Flags, "unread") {
		t.Fatalf("welcome entry=%#v delivery=%q", welcome.Entry, welcome.DeliveryStatus)
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

func TestRegistrationAddressChoiceIsLimitedToDuplicateNames(t *testing.T) {
	database, err := storage.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	service := newTestIdentity(t, database)

	local, required, err := service.RegistrationAddress("Same Name")
	if err != nil || local != "same-name" || required {
		t.Fatalf("initial suggestion local=%q required=%v err=%v", local, required, err)
	}
	if _, err := service.BeginTOTPRegistration("First Person", "custom-address", "first@example.net"); !errors.Is(err, ErrInvalidRegistration) {
		t.Fatalf("non-duplicate custom address error = %v", err)
	}
	if _, _, err := service.RegistrationAddress("Admin"); !errors.Is(err, ErrInvalidRegistration) {
		t.Fatalf("reserved display name error = %v", err)
	}
	first, err := service.Register(testRegistration("Same Name"))
	if err != nil {
		t.Fatal(err)
	}
	local, required, err = service.RegistrationAddress("Same Name")
	if err != nil || local != first.Username || !required {
		t.Fatalf("duplicate suggestion local=%q required=%v err=%v", local, required, err)
	}
	if _, err := service.BeginTOTPRegistration("Same Name", "", "second@example.net"); !errors.Is(err, ErrAddressChoiceRequired) {
		t.Fatalf("missing duplicate address error = %v", err)
	}
	if _, err := service.BeginTOTPRegistration("Same Name", "same-name-two", "second@example.net"); err != nil {
		t.Fatalf("duplicate custom address: %v", err)
	}
}

func TestDuplicateAccountCanChangeUniqueAddressAndPreservesProfileAlias(t *testing.T) {
	database, err := storage.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	service := newTestIdentity(t, database)
	if _, err := service.Register(testRegistration("Same Name")); err != nil {
		t.Fatal(err)
	}
	registration := testRegistration("Same Name")
	registration.Username = "same-name-two"
	second, err := service.Register(registration)
	if err != nil {
		t.Fatal(err)
	}
	code, err := totp.GenerateCode(testTOTPSecret, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	updated, err := service.ChangeWaveAddress(second.ID, code, "same-name-alt")
	if err != nil {
		t.Fatal(err)
	}
	if updated.Email != "same-name-alt@wave-lang.dev" || updated.Username != "same-name-alt" {
		t.Fatalf("updated account = %#v", updated)
	}
	box, err := service.Mailbox(second.ID)
	if err != nil || box.Address != updated.Email || !service.HasLocalRecipient(updated.Email) {
		t.Fatalf("mailbox=%#v err=%v", box, err)
	}
	aliased, err := service.PublicAccount("same-name-two")
	if err != nil || aliased.ID != second.ID {
		t.Fatalf("profile alias=%#v err=%v", aliased, err)
	}
}

func TestManagementAddressesShareOneMailbox(t *testing.T) {
	database, err := storage.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	service := newTestIdentity(t, database)
	box, err := service.EnsureManagementMailbox()
	if err != nil {
		t.Fatal(err)
	}
	for _, address := range service.ManagementAddresses() {
		if !service.HasLocalRecipient(address) {
			t.Fatalf("management recipient %q is unavailable", address)
		}
		resolved, err := service.mailboxes.MailboxByAddress(address)
		if err != nil || resolved.ID != box.ID {
			t.Fatalf("address %q mailbox=%#v err=%v", address, resolved, err)
		}
	}
	sender, err := service.Register(testRegistration("Mailbox Sender"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.SendMail(sender, OutgoingMail{To: "help@wave-lang.dev", Subject: "Need help", Body: "A management request."}); err != nil {
		t.Fatal(err)
	}
	items, err := service.ManagementMailboxItems("Inbox")
	if err != nil || len(items) != 1 || items[0].Message.Subject != "Need help" {
		t.Fatalf("management items=%#v err=%v", items, err)
	}
}

func TestPatchAddressIsReservedAndBackedByDedicatedMailbox(t *testing.T) {
	database, err := storage.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	service := newTestIdentity(t, database)
	box, err := service.EnsurePatchMailbox()
	if err != nil {
		t.Fatal(err)
	}
	if box.AccountID != PatchMailboxAccountID || service.PatchAddress() != "patchs@wave-lang.dev" || !service.HasLocalRecipient(service.PatchAddress()) {
		t.Fatalf("patch mailbox=%#v address=%q", box, service.PatchAddress())
	}
	if _, err := service.BeginTOTPRegistration("Patchs", "patchs", "patch@example.net"); !errors.Is(err, ErrInvalidRegistration) {
		t.Fatalf("reserved patch address error = %v", err)
	}
}

func TestIncomingPatchQueuesPatchWebhookEvent(t *testing.T) {
	database, err := storage.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	service := newTestIdentity(t, database)
	if _, err := service.EnsurePatchMailbox(); err != nil {
		t.Fatal(err)
	}
	webhooks, err := webhookdomain.NewService(database, "MDEyMzQ1Njc4OWFiY2RlZjAxMjM0NTY3ODlhYmNkZWY=", "https://wave-lang.dev")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := webhooks.SaveEndpoint("owner", webhookdomain.EndpointInput{Name: "Patch feed", Kind: "generic", URL: "https://hooks.example.test/patches", Events: []string{webhookdomain.EventPatchReceived}, Enabled: true}); err != nil {
		t.Fatal(err)
	}
	service.SetWebhookService(webhooks)
	raw := "From: Contributor <dev@example.net>\r\nTo: patchs@wave-lang.dev\r\nSubject: [PATCH] parser: keep ranges\r\nMessage-ID: <patch-1@example.net>\r\nContent-Type: text/plain; charset=utf-8\r\n\r\nCommit message.\n\ndiff --git a/parser.go b/parser.go\n--- a/parser.go\n+++ b/parser.go\n"
	item, err := service.AcceptSMTP(nil, "dev@example.net", []string{service.PatchAddress()}, []byte(raw))
	if err != nil {
		t.Fatal(err)
	}
	storedBody, err := service.mail.Body(item.Message)
	if err != nil || !patchdomain.Valid(item.Message.Subject, storedBody) {
		t.Fatalf("stored patch body was not recognized: subject=%q body=%q err=%v", item.Message.Subject, storedBody, err)
	}
	deliveries, err := webhooks.Deliveries(10)
	if err != nil || len(deliveries) != 1 || deliveries[0].EventType != webhookdomain.EventPatchReceived || !strings.Contains(deliveries[0].ResourceURL, item.Message.ID) {
		t.Fatalf("deliveries=%#v err=%v", deliveries, err)
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

func TestSystemMailIncludesPlainTextAndLinkedHTMLAlternatives(t *testing.T) {
	content := systemMailContent{
		Subject:     "Confirm your Wave recovery email",
		Heading:     "Confirm your recovery email",
		Intro:       "Use the button below to confirm this email address for your Wave account.",
		ActionLabel: "Confirm recovery email",
		ActionURL:   "https://wave-lang.dev/account/verify-recovery?token=abc&source=email",
		Expiry:      "This link expires in 24 hours.",
		Ignore:      "If you did not create or modify a Wave account, you can safely ignore this email.",
	}
	plainBody, htmlBody, err := renderSystemMail(content, "https://wave-lang.dev")
	if err != nil {
		t.Fatal(err)
	}
	message := maildomain.Message{
		ID: "message-test", MessageID: "<message-test@wave-lang.dev>", From: "Wave Security <security@wave-lang.dev>",
		To: []string{"user@example.net"}, Subject: content.Subject, CreatedAt: time.Date(2026, 8, 7, 12, 0, 0, 0, time.UTC),
	}
	raw, err := encodeSystemMail(message, plainBody, htmlBody)
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := stdmail.ReadMessage(bytes.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}
	mediaType, parameters, err := mime.ParseMediaType(parsed.Header.Get("Content-Type"))
	if err != nil || mediaType != "multipart/alternative" {
		t.Fatalf("content type=%q parameters=%#v err=%v", mediaType, parameters, err)
	}
	if parsed.Header.Get("Auto-Submitted") != "auto-generated" {
		t.Fatalf("Auto-Submitted=%q", parsed.Header.Get("Auto-Submitted"))
	}
	reader := multipart.NewReader(parsed.Body, parameters["boundary"])
	parts := map[string]string{}
	for {
		part, partErr := reader.NextRawPart()
		if partErr == io.EOF {
			break
		}
		if partErr != nil {
			t.Fatal(partErr)
		}
		partType, _, _ := mime.ParseMediaType(part.Header.Get("Content-Type"))
		data, readErr := io.ReadAll(quotedprintable.NewReader(part))
		if readErr != nil {
			t.Fatal(readErr)
		}
		parts[partType] = string(data)
	}
	if !strings.Contains(parts["text/plain"], content.ActionURL) {
		t.Fatalf("plain text does not contain action URL: %q", parts["text/plain"])
	}
	if !strings.Contains(parts["text/html"], `href="https://wave-lang.dev/account/verify-recovery?token=abc&amp;source=email"`) ||
		!strings.Contains(parts["text/html"], ">Confirm recovery email</a>") {
		t.Fatalf("HTML action link is missing: %q", parts["text/html"])
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
