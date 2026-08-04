package mailruntime

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"errors"
	"strings"
	"testing"

	"github.com/emersion/go-msgauth/dkim"
	"github.com/emersion/go-smtp"
	"github.com/wavefnd/wave-platform/internal/config"
	"github.com/wavefnd/wave-platform/internal/identity"
	maildomain "github.com/wavefnd/wave-platform/internal/mail"
	"github.com/wavefnd/wave-platform/internal/storage"
	"github.com/wavefnd/wave-platform/internal/testsupport"
)

type recordingSender struct {
	deliveries []maildomain.Delivery
	err        error
}

func TestDKIMSigningProducesVerifiableMessage(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	sender := &smtpSender{config: config.MailConfig{DKIMDomain: "wave-lang.dev", DKIMSelector: "wave-test"}, signer: privateKey}
	raw := []byte("From: Ada <ada@wave-lang.dev>\r\nTo: person@example.net\r\nSubject: Signed message\r\nDate: Tue, 05 Aug 2026 12:00:00 +0900\r\nMessage-ID: <signed@wave-lang.dev>\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=utf-8\r\nContent-Transfer-Encoding: 8bit\r\n\r\nSigned body.\r\n")
	signed, err := sender.sign(raw)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(string(signed), "DKIM-Signature:") {
		t.Fatalf("signed message has no DKIM header: %s", signed)
	}
	publicDER, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		t.Fatal(err)
	}
	verifications, err := dkim.VerifyWithOptions(strings.NewReader(string(signed)), &dkim.VerifyOptions{
		LookupTXT: func(domain string) ([]string, error) {
			if domain != "wave-test._domainkey.wave-lang.dev" {
				t.Fatalf("DKIM lookup domain = %q", domain)
			}
			return []string{"v=DKIM1; k=rsa; p=" + base64.StdEncoding.EncodeToString(publicDER)}, nil
		},
	})
	if err != nil || len(verifications) != 1 || verifications[0].Err != nil {
		t.Fatalf("verifications=%#v err=%v", verifications, err)
	}
}

func (sender *recordingSender) Send(delivery maildomain.Delivery, _ []byte) error {
	sender.deliveries = append(sender.deliveries, delivery)
	return sender.err
}

func TestSMTPIngressDeliversOnlyToLocalMailbox(t *testing.T) {
	database, err := storage.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	identityService, err := testsupport.NewIdentity(database)
	if err != nil {
		t.Fatal(err)
	}
	recipient, err := testsupport.Register(identityService, "Grace Hopper")
	if err != nil {
		t.Fatal(err)
	}
	session := &smtpSession{backend: &smtpBackend{identity: identityService, maxMessageBytes: 1024 * 1024}}
	if err := session.Mail("sender@example.net", nil); err != nil {
		t.Fatal(err)
	}
	if err := session.Rcpt("missing@wave-lang.dev", nil); err == nil {
		t.Fatal("missing local mailbox was accepted")
	}
	if err := session.Rcpt("relay@example.net", nil); err == nil {
		t.Fatal("Internet relay recipient was accepted")
	}
	if err := session.Rcpt(recipient.Email, nil); err != nil {
		t.Fatal(err)
	}
	raw := "From: Sender <sender@example.net>\r\nTo: " + recipient.Email + "\r\nSubject: Incoming message\r\nMessage-ID: <incoming@example.net>\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=utf-8\r\n\r\nHello from outside.\r\n"
	if err := session.Data(strings.NewReader(raw)); err != nil {
		t.Fatal(err)
	}
	items, err := identityService.MailboxItems(recipient.ID, "Inbox")
	if err != nil || len(items) != 1 {
		t.Fatalf("inbox=%d err=%v", len(items), err)
	}
	if items[0].Message.MessageID != "<incoming@example.net>" || !strings.Contains(items[0].Body, "Hello from outside") {
		t.Fatalf("message = %#v body=%q", items[0].Message, items[0].Body)
	}
}

func TestSMTPIngressDoesNotOfferClientAuthentication(t *testing.T) {
	session := &smtpSession{backend: &smtpBackend{maxMessageBytes: 1024 * 1024}}
	if mechanisms := session.AuthMechanisms(); len(mechanisms) != 0 {
		t.Fatalf("authentication mechanisms = %#v", mechanisms)
	}
	if _, err := session.Auth("PLAIN"); !errors.Is(err, smtp.ErrAuthUnsupported) {
		t.Fatalf("AUTH PLAIN error = %v", err)
	}
}

func TestDeliveryWorkerUpdatesPersistentStatus(t *testing.T) {
	database, err := storage.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	identityService, err := testsupport.NewIdentity(database)
	if err != nil {
		t.Fatal(err)
	}
	senderAccount, err := testsupport.Register(identityService, "Ada Lovelace")
	if err != nil {
		t.Fatal(err)
	}
	sent, err := identityService.SendMail(senderAccount, identity.OutgoingMail{To: "person@example.net", Subject: "Queue test", Body: "Message body"})
	if err != nil {
		t.Fatal(err)
	}
	transport := &recordingSender{}
	service, err := New(config.MailConfig{Hostname: "mail.wave-lang.dev"}, database, identityService)
	if err != nil {
		t.Fatal(err)
	}
	service.sender = transport
	service.processDeliveries()
	if len(transport.deliveries) != 1 {
		t.Fatalf("transport deliveries = %d", len(transport.deliveries))
	}
	deliveries, err := maildomain.NewRepository(database).DeliveriesByMessage(sent.Message.ID)
	if err != nil || len(deliveries) != 1 || deliveries[0].Status != "delivered" {
		t.Fatalf("deliveries=%#v err=%v", deliveries, err)
	}

	second, err := identityService.SendMail(senderAccount, identity.OutgoingMail{To: "other@example.net", Subject: "Retry test", Body: "Message body"})
	if err != nil {
		t.Fatal(err)
	}
	transport.err = errors.New("temporary remote failure")
	service.processDeliveries()
	failed, err := maildomain.NewRepository(database).DeliveriesByMessage(second.Message.ID)
	if err != nil || len(failed) != 1 || failed[0].Status != "deferred" || failed[0].Attempts != 1 {
		t.Fatalf("deferred deliveries=%#v err=%v", failed, err)
	}
}
