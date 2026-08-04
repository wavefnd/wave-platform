package identity

import (
	"bytes"
	"errors"
	"fmt"
	"mime"
	stdmail "net/mail"
	"net/textproto"
	"strings"
	"time"

	"github.com/wavefnd/wave-platform/internal/account"
	"github.com/wavefnd/wave-platform/internal/audit"
	"github.com/wavefnd/wave-platform/internal/auth"
	"github.com/wavefnd/wave-platform/internal/identifier"
	maildomain "github.com/wavefnd/wave-platform/internal/mail"
	"github.com/wavefnd/wave-platform/internal/mailbox"
	"github.com/wavefnd/wave-platform/internal/permission"
	"github.com/wavefnd/wave-platform/internal/session"
	"github.com/wavefnd/wave-platform/internal/storage"
)

var (
	ErrRegistrationClosed  = errors.New("registration is closed")
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrAccountInactive     = errors.New("account is not active")
	ErrInvalidRegistration = errors.New("invalid registration")
	ErrInvalidMail         = errors.New("invalid mail message")
	ErrRelayDenied         = errors.New("mail relay denied")
)

var reservedUsernames = map[string]bool{
	"abuse": true, "admin": true, "administrator": true, "hostmaster": true,
	"mailer-daemon": true, "postmaster": true, "root": true, "security": true,
}

type Registration struct {
	DisplayName   string
	Username      string
	TOTPSecret    string
	RecoveryEmail string
	Admin         bool
}

type MailboxItem struct {
	Entry          mailbox.Entry
	Message        maildomain.Message
	Body           string
	DeliveryStatus string
}

type OutgoingMail struct {
	To      string
	Subject string
	Body    string
}

type Service struct {
	accounts         *account.Repository
	totp             *auth.TOTPService
	sessions         *session.Service
	mailboxes        *mailbox.Repository
	mail             *maildomain.Repository
	permissions      *permission.Repository
	audit            *audit.Repository
	mailDomain       string
	registrationOpen bool
	publicURL        string
	now              func() time.Time
}

func NewServiceWithTOTP(database *storage.Database, mailDomain string, registrationOpen bool, sessionDuration time.Duration,
	encryptionKey, issuer, publicURL string) (*Service, error) {
	totpService, err := auth.NewTOTPService(database, encryptionKey, issuer)
	if err != nil {
		return nil, err
	}
	return &Service{
		accounts: account.NewRepository(database), totp: totpService,
		sessions:  session.NewService(session.NewRepository(database), sessionDuration),
		mailboxes: mailbox.NewRepository(database), mail: maildomain.NewRepository(database), permissions: permission.NewRepository(database),
		audit:      audit.NewRepository(database),
		mailDomain: strings.ToLower(mailDomain), registrationOpen: registrationOpen, publicURL: strings.TrimRight(publicURL, "/"), now: time.Now,
	}, nil
}

func (service *Service) Register(registration Registration) (account.Account, error) {
	if !registration.Admin && !service.registrationOpen {
		return account.Account{}, ErrRegistrationClosed
	}
	displayName := strings.TrimSpace(registration.DisplayName)
	if displayName == "" || len([]rune(displayName)) > 80 {
		return account.Account{}, fmt.Errorf("%w: display name must contain between 1 and 80 characters", ErrInvalidRegistration)
	}
	username := strings.TrimSpace(strings.ToLower(registration.Username))
	if username == "" {
		var err error
		username, err = account.LocalPart(displayName)
		if err != nil {
			return account.Account{}, fmt.Errorf("%w: %v", ErrInvalidRegistration, err)
		}
	}
	base, err := account.LocalPart(username)
	if err != nil {
		return account.Account{}, fmt.Errorf("%w: %v", ErrInvalidRegistration, err)
	}
	username = base
	if !registration.Admin && reservedUsernames[username] {
		username = "user-" + username
		base = username
	}
	if registration.Admin {
		if _, err := service.accounts.ByUsername(username); err == nil {
			return account.Account{}, account.ErrConflict
		}
	} else {
		username = service.availableUsername(base)
	}
	address, err := account.Address(username, service.mailDomain)
	if err != nil {
		return account.Account{}, err
	}
	id, err := identifier.New("account")
	if err != nil {
		return account.Account{}, err
	}
	now := service.now().UTC()
	item := account.Account{ID: id, Username: username, DisplayName: displayName, Email: address,
		Status: "active", CreatedAt: now, UpdatedAt: now}
	if err := service.accounts.Create(item); err != nil {
		return account.Account{}, err
	}
	rollbackAccount := true
	defer func() {
		if rollbackAccount {
			_ = service.accounts.Delete(item)
		}
	}()
	if registration.TOTPSecret == "" || registration.RecoveryEmail == "" {
		return account.Account{}, fmt.Errorf("%w: authenticator and recovery email are required", ErrInvalidRegistration)
	}
	if err := service.totp.PutFactor(item.ID, registration.TOTPSecret, registration.RecoveryEmail, false); err != nil {
		return account.Account{}, err
	}
	rollbackCredential := true
	defer func() {
		if rollbackCredential {
			_ = service.totp.DeleteFactor(item.ID)
		}
	}()
	mailboxID, err := identifier.New("mailbox")
	if err != nil {
		return account.Account{}, err
	}
	box := mailbox.Mailbox{ID: mailboxID, AccountID: item.ID, Address: item.Email, CreatedAt: now}
	if err := service.mailboxes.UpsertMailbox(box); err != nil {
		return account.Account{}, fmt.Errorf("create mailbox: %w", err)
	}
	rollbackMailbox := true
	defer func() {
		if rollbackMailbox {
			_ = service.mailboxes.DeleteMailbox(box)
		}
	}()
	if registration.Admin {
		if err := service.ensureManagementRoles(); err != nil {
			return account.Account{}, err
		}
		if err := service.permissions.Assign(permission.Assignment{AccountID: item.ID, RoleID: "platform-owner", Scope: "platform"}); err != nil {
			return account.Account{}, err
		}
		rollbackAssignment := true
		defer func() {
			if rollbackAssignment {
				_ = service.permissions.Unassign(item.ID, "platform-owner")
			}
		}()
		eventID, err := identifier.New("audit")
		if err != nil {
			return account.Account{}, err
		}
		if err := service.audit.Append(audit.Event{ID: eventID, ActorID: "system/bootstrap", ResourceID: "account/" + item.ID,
			Action: "account.bootstrap", Result: "success", OccurredAt: now}); err != nil {
			return account.Account{}, err
		}
		rollbackAssignment = false
	}
	rollbackCredential = false
	rollbackAccount = false
	rollbackMailbox = false
	return item, nil
}

func (service *Service) BootstrapTOTPAdmin(displayName, username, recoveryEmail, secret string) (account.Account, bool, error) {
	displayName, username = strings.TrimSpace(displayName), strings.TrimSpace(username)
	if displayName == "" && username == "" && recoveryEmail == "" && secret == "" {
		return account.Account{}, false, nil
	}
	if displayName == "" || recoveryEmail == "" || secret == "" {
		return account.Account{}, false, errors.New("admin display name, recovery email, and TOTP secret must be configured together")
	}
	if err := service.ensureManagementRoles(); err != nil {
		return account.Account{}, false, err
	}
	lookup := username
	if lookup == "" {
		var err error
		lookup, err = account.LocalPart(displayName)
		if err != nil {
			return account.Account{}, false, err
		}
	}
	existing, err := service.accounts.ByUsername(lookup)
	if err == nil {
		if !service.IsOwner(existing.ID) {
			return account.Account{}, false, fmt.Errorf("admin username %q belongs to a non-owner account", lookup)
		}
		if _, factorErr := service.totp.Factor(existing.ID); errors.Is(factorErr, storage.ErrNotFound) {
			if err := service.totp.PutFactor(existing.ID, secret, recoveryEmail, false); err != nil {
				return account.Account{}, false, err
			}
			_ = service.sendRecoveryVerification(existing.ID, recoveryEmail)
		}
		return existing, false, nil
	}
	if !errors.Is(err, storage.ErrNotFound) {
		return account.Account{}, false, err
	}
	item, err := service.Register(Registration{DisplayName: displayName, Username: username, TOTPSecret: secret,
		RecoveryEmail: recoveryEmail, Admin: true})
	if err == nil {
		_ = service.sendRecoveryVerification(item.ID, recoveryEmail)
	}
	return item, err == nil, err
}

func (service *Service) TOTPConfigured() bool { return service.totp.Configured() }

func (service *Service) BeginTOTPRegistration(displayName, recoveryEmail string) (auth.EnrollmentResult, error) {
	if !service.registrationOpen {
		return auth.EnrollmentResult{}, ErrRegistrationClosed
	}
	displayName = strings.TrimSpace(displayName)
	if displayName == "" || len([]rune(displayName)) > 80 {
		return auth.EnrollmentResult{}, fmt.Errorf("%w: display name must contain between 1 and 80 characters", ErrInvalidRegistration)
	}
	username, err := account.LocalPart(displayName)
	if err != nil {
		return auth.EnrollmentResult{}, fmt.Errorf("%w: %v", ErrInvalidRegistration, err)
	}
	address, err := account.Address(username, service.mailDomain)
	if err != nil {
		return auth.EnrollmentResult{}, err
	}
	return service.totp.BeginEnrollment("registration", "", displayName, "", recoveryEmail, address)
}

func (service *Service) CompleteTOTPRegistration(token, code, userAgent string) (account.Account, string, session.Session, error) {
	enrollment, secret, err := service.totp.Enrollment(token, code)
	if err != nil || enrollment.Kind != "registration" {
		return account.Account{}, "", session.Session{}, ErrInvalidCredentials
	}
	item, err := service.Register(Registration{DisplayName: enrollment.DisplayName, TOTPSecret: secret,
		RecoveryEmail: enrollment.RecoveryEmail})
	if err != nil {
		return account.Account{}, "", session.Session{}, err
	}
	if err := service.totp.PutFactorFromEnrollment(item.ID, secret, enrollment.RecoveryEmail, false, code); err != nil {
		return account.Account{}, "", session.Session{}, err
	}
	_ = service.totp.DeleteEnrollment(token)
	_ = service.sendRecoveryVerification(item.ID, enrollment.RecoveryEmail)
	sessionToken, currentSession, err := service.sessions.Create(item.ID, userAgent)
	return item, sessionToken, currentSession, err
}

func (service *Service) AuthenticateTOTP(identifier, code, userAgent string) (account.Account, string, session.Session, error) {
	item, err := service.accountByIdentifier(identifier)
	if err != nil {
		return account.Account{}, "", session.Session{}, ErrInvalidCredentials
	}
	if item.Status != "active" {
		return account.Account{}, "", session.Session{}, ErrAccountInactive
	}
	if err := service.totp.Verify(item.ID, code); err != nil {
		return account.Account{}, "", session.Session{}, ErrInvalidCredentials
	}
	token, currentSession, err := service.sessions.Create(item.ID, userAgent)
	return item, token, currentSession, err
}

func (service *Service) BeginTOTPRotation(accountID, currentCode string) (auth.EnrollmentResult, error) {
	if err := service.totp.Verify(accountID, currentCode); err != nil {
		return auth.EnrollmentResult{}, ErrInvalidCredentials
	}
	item, err := service.accounts.Account(accountID)
	if err != nil {
		return auth.EnrollmentResult{}, err
	}
	factor, err := service.totp.Factor(accountID)
	if err != nil {
		return auth.EnrollmentResult{}, err
	}
	return service.totp.BeginEnrollment("rotation", accountID, "", "", factor.RecoveryEmail, item.Email)
}

func (service *Service) CompleteTOTPRotation(accountID, token, code string) error {
	enrollment, secret, err := service.totp.Enrollment(token, code)
	if err != nil || enrollment.Kind != "rotation" || enrollment.AccountID != accountID {
		return ErrInvalidCredentials
	}
	factor, err := service.totp.Factor(accountID)
	if err != nil {
		return err
	}
	if err := service.totp.PutFactorFromEnrollment(accountID, secret, factor.RecoveryEmail, factor.RecoveryVerified, code); err != nil {
		return err
	}
	_ = service.totp.DeleteEnrollment(token)
	return service.sessions.RevokeAccount(accountID)
}

func (service *Service) RequestTOTPRecovery(identifier string) error {
	item, err := service.accountByIdentifier(identifier)
	if err != nil {
		return nil
	}
	email, _, result, err := service.totp.BeginRecovery(item.ID)
	if err != nil {
		return nil
	}
	link := service.baseURL() + "/account/recover?token=" + result.Token
	body := "A request was made to reset the authenticator for your Wave account.\n\n" + link +
		"\n\nThis link expires in 30 minutes. If you did not request this, ignore this message."
	_, err = service.SendSystemMail(email, "Reset your Wave authenticator", body)
	return err
}

func (service *Service) TOTPRecovery(token string) (auth.EnrollmentResult, error) {
	_, result, err := service.totp.Recovery(token)
	return result, err
}

func (service *Service) CompleteTOTPRecovery(token, code string) error {
	accountID, err := service.totp.CompleteRecovery(token, code)
	if err != nil {
		return err
	}
	return service.sessions.RevokeAccount(accountID)
}

func (service *Service) SecurityStatus(accountID string) (auth.TOTPFactor, error) {
	return service.totp.Factor(accountID)
}

func (service *Service) ChangeRecoveryEmail(accountID, currentCode, email string) error {
	if err := service.totp.Verify(accountID, currentCode); err != nil {
		return ErrInvalidCredentials
	}
	return service.sendRecoveryVerification(accountID, email)
}

func (service *Service) VerifyRecoveryEmail(token string) error {
	_, err := service.totp.CompleteEmailVerification(token)
	return err
}

func (service *Service) sendRecoveryVerification(accountID, email string) error {
	token, err := service.totp.BeginEmailVerification(accountID, email)
	if err != nil {
		return err
	}
	link := service.baseURL() + "/account/verify-recovery?token=" + token
	body := "Confirm this address as the recovery email for your Wave account.\n\n" + link +
		"\n\nThis link expires in 24 hours. If you did not request this, ignore this message."
	_, err = service.SendSystemMail(email, "Confirm your Wave recovery email", body)
	return err
}

func (service *Service) accountByIdentifier(identifier string) (account.Account, error) {
	identifier = strings.ToLower(strings.TrimSpace(identifier))
	if strings.Contains(identifier, "@") {
		return service.accounts.ByEmail(identifier)
	}
	return service.accounts.ByUsername(identifier)
}

func (service *Service) baseURL() string {
	if service.publicURL != "" {
		return service.publicURL
	}
	return "http://localhost:8080"
}

func (service *Service) availableUsername(base string) string {
	for suffix := 1; suffix < 10000; suffix++ {
		candidate := base
		if suffix > 1 {
			candidate = fmt.Sprintf("%s-%d", base, suffix)
		}
		if _, err := service.accounts.ByUsername(candidate); errors.Is(err, storage.ErrNotFound) {
			return candidate
		}
	}
	return base + "-" + fmt.Sprint(service.now().UTC().Unix())
}

func (service *Service) AccountForToken(token string) (account.Account, session.Session, error) {
	currentSession, err := service.sessions.Verify(token)
	if err != nil {
		return account.Account{}, session.Session{}, err
	}
	item, err := service.accounts.Account(currentSession.AccountID)
	if err != nil {
		return account.Account{}, session.Session{}, err
	}
	if item.Status != "active" {
		return account.Account{}, session.Session{}, ErrAccountInactive
	}
	return item, currentSession, nil
}

func (service *Service) Revoke(token string) error { return service.sessions.Revoke(token) }

func (service *Service) IsAdministrator(accountID string) bool {
	if service.IsOwner(accountID) {
		return true
	}
	allowed, err := service.permissions.HasRole(accountID, "platform-admin")
	return err == nil && allowed
}

func (service *Service) IsOwner(accountID string) bool {
	allowed, err := service.permissions.HasRole(accountID, "platform-owner")
	return err == nil && allowed
}

func (service *Service) ensureManagementRoles() error {
	for _, role := range []permission.Role{
		{ID: "platform-owner", Name: "Platform owner", Permissions: []string{"platform.*"}},
		{ID: "platform-admin", Name: "Platform administrator", Permissions: []string{"platform.admin.*"}},
	} {
		if err := service.permissions.PutRole(role); err != nil {
			return err
		}
	}
	return nil
}

func (service *Service) Mailbox(accountID string) (mailbox.Mailbox, error) {
	return service.mailboxes.MailboxByAccount(accountID)
}

func (service *Service) MailboxEntries(accountID, folder string) ([]mailbox.Entry, error) {
	box, err := service.Mailbox(accountID)
	if err != nil {
		return nil, err
	}
	return service.mailboxes.Entries(box.ID, folder)
}

func (service *Service) MailboxItems(accountID, folder string) ([]MailboxItem, error) {
	box, err := service.Mailbox(accountID)
	if err != nil {
		return nil, err
	}
	entries, err := service.mailboxes.Entries(box.ID, folder)
	if err != nil {
		return nil, err
	}
	items := make([]MailboxItem, 0, len(entries))
	for _, entry := range entries {
		message, err := service.mailboxes.Message(entry)
		if err != nil {
			return nil, err
		}
		body, err := service.mail.Body(message)
		if err != nil {
			return nil, err
		}
		items = append(items, MailboxItem{Entry: entry, Message: message, Body: body,
			DeliveryStatus: service.deliveryStatus(message.ID)})
	}
	return items, nil
}

func (service *Service) MailboxItem(accountID, entryID string) (MailboxItem, error) {
	box, err := service.Mailbox(accountID)
	if err != nil {
		return MailboxItem{}, err
	}
	entry, err := service.mailboxes.Entry(box.ID, entryID)
	if err != nil {
		return MailboxItem{}, err
	}
	message, err := service.mailboxes.Message(entry)
	if err != nil {
		return MailboxItem{}, err
	}
	body, err := service.mail.Body(message)
	if err != nil {
		return MailboxItem{}, err
	}
	return MailboxItem{Entry: entry, Message: message, Body: body,
		DeliveryStatus: service.deliveryStatus(message.ID)}, nil
}

func (service *Service) SendMail(actor account.Account, outgoing OutgoingMail) (MailboxItem, error) {
	recipients, err := normalizeRecipients(outgoing.To)
	if err != nil {
		return MailboxItem{}, fmt.Errorf("%w: %v", ErrInvalidMail, err)
	}
	if err := service.validateLocalRecipients(recipients); err != nil {
		return MailboxItem{}, err
	}
	subject := strings.TrimSpace(outgoing.Subject)
	body := strings.TrimSpace(strings.ReplaceAll(outgoing.Body, "\r\n", "\n"))
	if len([]rune(subject)) < 1 || len([]rune(subject)) > 180 || strings.ContainsAny(subject, "\r\n") {
		return MailboxItem{}, fmt.Errorf("%w: subject must contain between 1 and 180 characters", ErrInvalidMail)
	}
	if len([]rune(body)) < 1 || len([]rune(body)) > 50000 {
		return MailboxItem{}, fmt.Errorf("%w: body must contain between 1 and 50000 characters", ErrInvalidMail)
	}
	messageID, err := identifier.New("message")
	if err != nil {
		return MailboxItem{}, err
	}
	now := service.now().UTC()
	message := maildomain.Message{
		ID: messageID, MessageID: "<" + messageID + "@" + service.mailDomain + ">", ThreadID: messageID,
		AuthorAccountID: actor.ID, From: (&stdmail.Address{Name: actor.DisplayName, Address: actor.Email}).String(),
		To: recipients, Subject: subject, ReceivedAt: now, CreatedAt: now,
	}
	if err := service.mail.UpsertMessage(message, encodeRawMail(message, body)); err != nil {
		return MailboxItem{}, err
	}
	senderBox, err := service.Mailbox(actor.ID)
	if err != nil {
		return MailboxItem{}, err
	}
	sentEntry, err := service.addMailboxEntry(senderBox.ID, message.ID, "Sent", nil)
	if err != nil {
		return MailboxItem{}, err
	}
	if err := service.routeMessage(actor.Email, message, recipients, true); err != nil {
		return MailboxItem{}, err
	}
	_ = service.appendAudit(actor.ID, "mail/message/"+message.ID, "mail.send")
	return MailboxItem{Entry: sentEntry, Message: message, Body: body,
		DeliveryStatus: service.deliveryStatus(message.ID)}, nil
}

func (service *Service) SendSystemMail(to, subject, body string) (MailboxItem, error) {
	recipients, err := normalizeRecipients(to)
	if err != nil {
		return MailboxItem{}, fmt.Errorf("%w: %v", ErrInvalidMail, err)
	}
	subject = strings.TrimSpace(subject)
	body = strings.TrimSpace(strings.ReplaceAll(body, "\r\n", "\n"))
	if subject == "" || body == "" || strings.ContainsAny(subject, "\r\n") {
		return MailboxItem{}, ErrInvalidMail
	}
	messageID, err := identifier.New("message")
	if err != nil {
		return MailboxItem{}, err
	}
	now := service.now().UTC()
	from := "security@" + service.mailDomain
	message := maildomain.Message{ID: messageID, MessageID: "<" + messageID + "@" + service.mailDomain + ">",
		ThreadID: messageID, AuthorAccountID: "system/security", From: (&stdmail.Address{Name: "Wave Security", Address: from}).String(),
		To: recipients, Subject: subject, ReceivedAt: now, CreatedAt: now}
	if err := service.mail.UpsertMessage(message, encodeRawMail(message, body)); err != nil {
		return MailboxItem{}, err
	}
	if err := service.routeMessage(from, message, recipients, true); err != nil {
		return MailboxItem{}, err
	}
	return MailboxItem{Message: message, Body: body, DeliveryStatus: service.deliveryStatus(message.ID)}, nil
}

// AcceptSMTP stores an RFC 5322/MIME message received through SMTP. A nil
// actor means Internet ingress and is therefore restricted to local mailbox
// recipients. A non-nil actor represents authenticated message submission.
func (service *Service) AcceptSMTP(actor *account.Account, envelopeFrom string, recipients []string, raw []byte) (MailboxItem, error) {
	if len(raw) == 0 {
		return MailboxItem{}, fmt.Errorf("%w: message data is empty", ErrInvalidMail)
	}
	normalized, err := normalizeRecipientSlice(recipients)
	if err != nil {
		return MailboxItem{}, fmt.Errorf("%w: %v", ErrInvalidMail, err)
	}
	if actor == nil {
		for _, recipient := range normalized {
			if !service.isLocalAddress(recipient) {
				return MailboxItem{}, ErrRelayDenied
			}
		}
	} else if !strings.EqualFold(strings.TrimSpace(envelopeFrom), actor.Email) {
		return MailboxItem{}, ErrRelayDenied
	}

	parsed, err := stdmail.ReadMessage(bytes.NewReader(raw))
	if err != nil {
		return MailboxItem{}, fmt.Errorf("%w: malformed RFC 5322 message", ErrInvalidMail)
	}
	messageID, err := identifier.New("message")
	if err != nil {
		return MailboxItem{}, err
	}
	now := service.now().UTC()
	subject, decodeErr := (&mime.WordDecoder{}).DecodeHeader(parsed.Header.Get("Subject"))
	if decodeErr != nil {
		subject = parsed.Header.Get("Subject")
	}
	from := strings.TrimSpace(parsed.Header.Get("From"))
	authorID := ""
	if actor != nil {
		headerFrom, parseErr := stdmail.ParseAddress(from)
		if parseErr != nil || !strings.EqualFold(headerFrom.Address, actor.Email) {
			return MailboxItem{}, fmt.Errorf("%w: From header must match authenticated account", ErrInvalidMail)
		}
		from = (&stdmail.Address{Name: actor.DisplayName, Address: actor.Email}).String()
		authorID = actor.ID
	} else if from == "" {
		from = strings.ToLower(strings.TrimSpace(envelopeFrom))
	}
	headerMessageID := strings.TrimSpace(parsed.Header.Get("Message-ID"))
	if headerMessageID == "" {
		headerMessageID = "<" + messageID + "@" + service.mailDomain + ">"
	}
	message := maildomain.Message{ID: messageID, MessageID: headerMessageID, ThreadID: messageID,
		AuthorAccountID: authorID, From: from, To: normalized, Subject: strings.TrimSpace(subject),
		ReceivedAt: now, CreatedAt: now}
	if err := service.mail.UpsertMessage(message, raw); err != nil {
		return MailboxItem{}, err
	}
	var sentEntry mailbox.Entry
	if actor != nil {
		box, err := service.Mailbox(actor.ID)
		if err != nil {
			return MailboxItem{}, err
		}
		sentEntry, err = service.addMailboxEntry(box.ID, message.ID, "Sent", nil)
		if err != nil {
			return MailboxItem{}, err
		}
	}
	if err := service.routeMessage(strings.ToLower(strings.TrimSpace(envelopeFrom)), message, normalized, actor != nil); err != nil {
		return MailboxItem{}, err
	}
	body, _ := service.mail.Body(message)
	return MailboxItem{Entry: sentEntry, Message: message, Body: body,
		DeliveryStatus: service.deliveryStatus(message.ID)}, nil
}

func (service *Service) HasLocalRecipient(address string) bool {
	address = strings.ToLower(strings.TrimSpace(address))
	if !service.isLocalAddress(address) {
		return false
	}
	_, err := service.mailboxes.MailboxByAddress(address)
	return err == nil
}

func (service *Service) routeMessage(sender string, message maildomain.Message, recipients []string, allowExternal bool) error {
	now := service.now().UTC()
	for _, recipient := range recipients {
		deliveryID, err := identifier.New("delivery")
		if err != nil {
			return err
		}
		delivery := maildomain.Delivery{ID: deliveryID, MessageID: message.ID, Sender: sender,
			Recipient: recipient, CreatedAt: now}
		if service.isLocalAddress(recipient) {
			box, err := service.mailboxes.MailboxByAddress(recipient)
			if err != nil {
				if errors.Is(err, storage.ErrNotFound) {
					return fmt.Errorf("%w: local recipient does not exist", ErrInvalidMail)
				}
				return err
			}
			if _, err := service.addMailboxEntry(box.ID, message.ID, "Inbox", []string{"unread"}); err != nil {
				return err
			}
			delivery.Destination = "local-mailbox"
			delivery.Status = "delivered"
			delivery.CompletedAt = now
		} else {
			if !allowExternal {
				return ErrRelayDenied
			}
			delivery.Destination = "remote-smtp"
			delivery.Status = "queued"
			delivery.NextAttemptAt = now
		}
		if err := service.mail.UpsertDelivery(delivery); err != nil {
			return err
		}
	}
	return nil
}

func (service *Service) isLocalAddress(address string) bool {
	parts := strings.Split(strings.ToLower(strings.TrimSpace(address)), "@")
	return len(parts) == 2 && parts[1] == service.mailDomain
}

func (service *Service) deliveryStatus(messageID string) string {
	deliveries, err := service.mail.DeliveriesByMessage(messageID)
	if err != nil || len(deliveries) == 0 {
		return ""
	}
	status := "delivered"
	rank := map[string]int{"delivered": 0, "queued": 1, "delivering": 2, "deferred": 3, "failed": 4}
	for _, delivery := range deliveries {
		if rank[delivery.Status] > rank[status] {
			status = delivery.Status
		}
	}
	return status
}

func (service *Service) validateLocalRecipients(recipients []string) error {
	for _, recipient := range recipients {
		if !service.isLocalAddress(recipient) {
			continue
		}
		if _, err := service.mailboxes.MailboxByAddress(recipient); err != nil {
			if errors.Is(err, storage.ErrNotFound) {
				return fmt.Errorf("%w: local recipient does not exist", ErrInvalidMail)
			}
			return err
		}
	}
	return nil
}

func normalizeRecipients(value string) ([]string, error) {
	addresses, err := stdmail.ParseAddressList(value)
	if err != nil {
		return nil, errors.New("recipient address is invalid")
	}
	if len(addresses) == 0 || len(addresses) > 50 {
		return nil, errors.New("between 1 and 50 recipients are required")
	}
	values := make([]string, 0, len(addresses))
	for _, address := range addresses {
		values = append(values, address.Address)
	}
	return normalizeRecipientSlice(values)
}

func normalizeRecipientSlice(values []string) ([]string, error) {
	if len(values) == 0 || len(values) > 50 {
		return nil, errors.New("between 1 and 50 recipients are required")
	}
	result := make([]string, 0, len(values))
	seen := map[string]bool{}
	for _, value := range values {
		address, err := stdmail.ParseAddress(strings.TrimSpace(value))
		if err != nil || address.Address == "" || strings.ContainsAny(address.Address, "\r\n") {
			return nil, errors.New("recipient address is invalid")
		}
		normalized := strings.ToLower(address.Address)
		if !seen[normalized] {
			seen[normalized] = true
			result = append(result, normalized)
		}
	}
	return result, nil
}

func (service *Service) UpdateMailboxEntry(accountID, entryID, action string) (MailboxItem, error) {
	item, err := service.MailboxItem(accountID, entryID)
	if err != nil {
		return MailboxItem{}, err
	}
	switch action {
	case "archive":
		item.Entry.Folder = "Archive"
	case "trash":
		item.Entry.Folder = "Trash"
	case "read":
		item.Entry.Flags = withoutFlag(item.Entry.Flags, "unread")
	case "unread":
		item.Entry.Flags = withFlag(item.Entry.Flags, "unread")
	default:
		return MailboxItem{}, fmt.Errorf("%w: unsupported mailbox action", ErrInvalidMail)
	}
	if err := service.mailboxes.UpdateEntry(item.Entry); err != nil {
		return MailboxItem{}, err
	}
	return item, nil
}

func (service *Service) addMailboxEntry(mailboxID, messageID, folder string, flags []string) (mailbox.Entry, error) {
	entryID, err := identifier.New("mailbox-entry")
	if err != nil {
		return mailbox.Entry{}, err
	}
	entry := mailbox.Entry{ID: entryID, MailboxID: mailboxID, MessageID: messageID, Folder: folder,
		Flags: flags, CreatedAt: service.now().UTC()}
	return entry, service.mailboxes.AddEntry(entry)
}

func (service *Service) appendAudit(actorID, resourceID, action string) error {
	id, err := identifier.New("audit")
	if err != nil {
		return err
	}
	return service.audit.Append(audit.Event{ID: id, ActorID: "account/" + actorID, ResourceID: resourceID,
		Action: action, Result: "success", OccurredAt: service.now().UTC()})
}

func encodeRawMail(message maildomain.Message, body string) []byte {
	headers := textproto.MIMEHeader{}
	headers.Set("From", message.From)
	headers.Set("To", strings.Join(message.To, ", "))
	headers.Set("Subject", mime.QEncoding.Encode("utf-8", message.Subject))
	headers.Set("Date", message.CreatedAt.Format(time.RFC1123Z))
	headers.Set("Message-ID", message.MessageID)
	headers.Set("MIME-Version", "1.0")
	headers.Set("Content-Type", "text/plain; charset=utf-8")
	headers.Set("Content-Transfer-Encoding", "8bit")
	order := []string{"From", "To", "Subject", "Date", "Message-ID", "MIME-Version", "Content-Type", "Content-Transfer-Encoding"}
	var raw strings.Builder
	for _, name := range order {
		raw.WriteString(name + ": " + headers.Get(name) + "\r\n")
	}
	raw.WriteString("\r\n")
	raw.WriteString(strings.ReplaceAll(body, "\n", "\r\n"))
	return []byte(raw.String())
}

func withoutFlag(flags []string, target string) []string {
	result := make([]string, 0, len(flags))
	for _, flag := range flags {
		if flag != target {
			result = append(result, flag)
		}
	}
	return result
}

func withFlag(flags []string, target string) []string {
	for _, flag := range flags {
		if flag == target {
			return flags
		}
	}
	return append(flags, target)
}
