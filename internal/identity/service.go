package identity

import (
	"bytes"
	"errors"
	"fmt"
	"html"
	"mime"
	"mime/multipart"
	"mime/quotedprintable"
	stdmail "net/mail"
	"net/textproto"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/wavefnd/wave-platform/internal/account"
	"github.com/wavefnd/wave-platform/internal/audit"
	"github.com/wavefnd/wave-platform/internal/auth"
	"github.com/wavefnd/wave-platform/internal/identifier"
	maildomain "github.com/wavefnd/wave-platform/internal/mail"
	"github.com/wavefnd/wave-platform/internal/mailbox"
	patchdomain "github.com/wavefnd/wave-platform/internal/patcharchive"
	"github.com/wavefnd/wave-platform/internal/permission"
	"github.com/wavefnd/wave-platform/internal/session"
	"github.com/wavefnd/wave-platform/internal/storage"
	webhookdomain "github.com/wavefnd/wave-platform/internal/webhook"
)

var (
	ErrRegistrationClosed    = errors.New("registration is closed")
	ErrInvalidCredentials    = errors.New("invalid credentials")
	ErrAccountInactive       = errors.New("account is not active")
	ErrInvalidRegistration   = errors.New("invalid registration")
	ErrInvalidMail           = errors.New("invalid mail message")
	ErrRelayDenied           = errors.New("mail relay denied")
	ErrInvalidProfile        = errors.New("invalid profile")
	ErrAddressChoiceRequired = errors.New("a different Wave mail address is required for this duplicate display name")
)

const ManagementMailboxAccountID = "role/platform-management"
const PatchMailboxAccountID = "service/patchs"

var managementLocalParts = []string{"admin", "help", "info", "support"}

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

type systemMailContent struct {
	Subject     string
	Heading     string
	Intro       string
	ActionLabel string
	ActionURL   string
	Expiry      string
	Ignore      string
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
	webhooks         *webhookdomain.Service
	publicURL        string
	now              func() time.Time
}

func (service *Service) SetWebhookService(webhooks *webhookdomain.Service) {
	service.webhooks = webhooks
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
	requestedUsername := strings.TrimSpace(strings.ToLower(registration.Username))
	username := requestedUsername
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
	if registration.Admin {
		if _, err := service.accounts.ByUsername(username); err == nil {
			return account.Account{}, account.ErrConflict
		}
	} else if requestedUsername == "" {
		if account.IsReservedLocalPart(base) {
			base = "user-" + base
		}
		username = service.availableUsername(base)
	} else {
		username, err = account.MailLocalPart(requestedUsername)
		if err != nil {
			return account.Account{}, fmt.Errorf("%w: %v", ErrInvalidRegistration, err)
		}
		if _, lookupErr := service.accounts.ByUsername(username); lookupErr == nil {
			return account.Account{}, account.ErrConflict
		} else if !errors.Is(lookupErr, storage.ErrNotFound) {
			return account.Account{}, lookupErr
		}
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
	item := account.Account{ID: id, Username: username, DisplayName: displayName, Email: address, TimeZone: "UTC",
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

func (service *Service) BeginTOTPRegistration(displayName, username, recoveryEmail string) (auth.EnrollmentResult, error) {
	if !service.registrationOpen {
		return auth.EnrollmentResult{}, ErrRegistrationClosed
	}
	displayName = strings.TrimSpace(displayName)
	if displayName == "" || len([]rune(displayName)) > 80 {
		return auth.EnrollmentResult{}, fmt.Errorf("%w: display name must contain between 1 and 80 characters", ErrInvalidRegistration)
	}
	generated, choiceRequired, err := service.RegistrationAddress(displayName)
	if err != nil {
		return auth.EnrollmentResult{}, err
	}
	username = strings.TrimSpace(username)
	if !choiceRequired {
		if username != "" && !strings.EqualFold(username, generated) {
			return auth.EnrollmentResult{}, fmt.Errorf("%w: a custom mail address is available only when the generated address is already in use", ErrInvalidRegistration)
		}
		username = generated
	} else {
		if username == "" {
			return auth.EnrollmentResult{}, ErrAddressChoiceRequired
		}
		var err error
		username, err = account.MailLocalPart(username)
		if err != nil {
			return auth.EnrollmentResult{}, fmt.Errorf("%w: %v", ErrInvalidRegistration, err)
		}
		if _, lookupErr := service.accounts.ByUsername(username); lookupErr == nil {
			return auth.EnrollmentResult{}, account.ErrConflict
		} else if !errors.Is(lookupErr, storage.ErrNotFound) {
			return auth.EnrollmentResult{}, lookupErr
		}
		if _, lookupErr := service.accounts.ByProfileAlias(username); lookupErr == nil {
			return auth.EnrollmentResult{}, account.ErrConflict
		} else if !errors.Is(lookupErr, storage.ErrNotFound) {
			return auth.EnrollmentResult{}, lookupErr
		}
	}
	address, err := account.Address(username, service.mailDomain)
	if err != nil {
		return auth.EnrollmentResult{}, err
	}
	return service.totp.BeginEnrollment("registration", "", displayName, username, recoveryEmail, address)
}

func (service *Service) RegistrationAddress(displayName string) (string, bool, error) {
	displayName = strings.TrimSpace(displayName)
	if displayName == "" || len([]rune(displayName)) > 80 {
		return "", false, fmt.Errorf("%w: display name must contain between 1 and 80 characters", ErrInvalidRegistration)
	}
	generated, err := account.LocalPart(displayName)
	if err != nil {
		return "", false, fmt.Errorf("%w: %v", ErrInvalidRegistration, err)
	}
	if account.IsReservedLocalPart(generated) {
		return "", false, fmt.Errorf("%w: this display name is reserved", ErrInvalidRegistration)
	}
	_, err = service.accounts.ByUsername(generated)
	if err == nil {
		return generated, true, nil
	}
	if !errors.Is(err, storage.ErrNotFound) {
		return "", false, err
	}
	_, err = service.accounts.ByProfileAlias(generated)
	if err == nil {
		return generated, true, nil
	}
	if !errors.Is(err, storage.ErrNotFound) {
		return "", false, err
	}
	return generated, false, nil
}

func (service *Service) CompleteTOTPRegistration(token, code, userAgent string) (account.Account, string, session.Session, error) {
	enrollment, secret, err := service.totp.Enrollment(token, code)
	if err != nil || enrollment.Kind != "registration" {
		return account.Account{}, "", session.Session{}, ErrInvalidCredentials
	}
	item, err := service.Register(Registration{DisplayName: enrollment.DisplayName, Username: enrollment.Username, TOTPSecret: secret,
		RecoveryEmail: enrollment.RecoveryEmail})
	if err != nil {
		return account.Account{}, "", session.Session{}, err
	}
	if err := service.totp.PutFactorFromEnrollment(item.ID, secret, enrollment.RecoveryEmail, false, code); err != nil {
		return account.Account{}, "", session.Session{}, err
	}
	_ = service.totp.DeleteEnrollment(token)
	_ = service.sendRecoveryVerification(item.ID, enrollment.RecoveryEmail)
	_ = service.sendWelcomeMail(item)
	sessionToken, currentSession, err := service.sessions.Create(item.ID, userAgent)
	return item, sessionToken, currentSession, err
}

func (service *Service) sendWelcomeMail(recipient account.Account) error {
	founder, err := service.founderAccount()
	if err != nil {
		return err
	}
	body := strings.Join([]string{
		"Hi " + recipient.DisplayName + ",",
		"",
		"Welcome to the Wave programming language community.",
		"",
		"Wave is an open-source systems programming language project built to explore safe, expressive, and practical software from applications down to the system layer. This platform is the shared home for its language, documentation, source, mail, questions, and community.",
		"",
		"You can exchange mail with other Wave members, meet people across the community, share what you are building, ask for help, and take part in the project. Contributions such as code, documentation, bug reports, design ideas, testing, and helping other members all support Wave and its maintainers.",
		"",
		"If you would also like to support ongoing development financially, you can sponsor Wave through Open Collective:",
		"https://opencollective.com/wave-lang",
		"",
		"I am glad you are here.",
		"",
		"— " + founder.DisplayName,
		"Founder of Wave",
	}, "\n")
	_, err = service.SendMail(founder, OutgoingMail{To: recipient.Email, Subject: "Welcome to the Wave community", Body: body})
	return err
}

func (service *Service) founderAccount() (account.Account, error) {
	items, err := service.accounts.Accounts()
	if err != nil {
		return account.Account{}, err
	}
	var founder account.Account
	for _, item := range items {
		owner, roleErr := service.permissions.HasRole(item.ID, "platform-owner")
		if roleErr != nil {
			return account.Account{}, roleErr
		}
		if owner && (founder.ID == "" || item.CreatedAt.Before(founder.CreatedAt)) {
			founder = item
		}
	}
	if founder.ID == "" {
		return account.Account{}, storage.ErrNotFound
	}
	return founder, nil
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
	_, err = service.sendSystemMail(email, systemMailContent{
		Subject:     "Reset your Wave authenticator",
		Heading:     "Reset your authenticator",
		Intro:       "Use the button below to connect a new authenticator app to your Wave account.",
		ActionLabel: "Reset authenticator",
		ActionURL:   link,
		Expiry:      "This link expires in 30 minutes.",
		Ignore:      "If you did not request this change, you can safely ignore this email.",
	})
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

func (service *Service) UpdateProfile(accountID, displayName, bio, timeZone string) (account.Account, error) {
	displayName = strings.TrimSpace(displayName)
	bio = strings.TrimSpace(strings.ReplaceAll(bio, "\r\n", "\n"))
	if len([]rune(displayName)) < 1 || len([]rune(displayName)) > 80 {
		return account.Account{}, fmt.Errorf("%w: display name must contain between 1 and 80 characters", ErrInvalidProfile)
	}
	if len([]rune(bio)) > 500 {
		return account.Account{}, fmt.Errorf("%w: profile introduction must be 500 characters or shorter", ErrInvalidProfile)
	}
	timeZone = strings.TrimSpace(timeZone)
	if timeZone == "" {
		timeZone = "UTC"
	}
	if len(timeZone) > 64 {
		return account.Account{}, fmt.Errorf("%w: time zone is invalid", ErrInvalidProfile)
	}
	if _, err := time.LoadLocation(timeZone); err != nil {
		return account.Account{}, fmt.Errorf("%w: time zone is invalid", ErrInvalidProfile)
	}
	item, err := service.accounts.Account(accountID)
	if err != nil {
		return account.Account{}, err
	}
	item.DisplayName = displayName
	item.Bio = bio
	item.TimeZone = timeZone
	item.UpdatedAt = service.now().UTC()
	if err := service.accounts.Update(item); err != nil {
		return account.Account{}, err
	}
	_ = service.appendAudit(accountID, "account/"+accountID+"/profile", "account.profile.update")
	return item, nil
}

func (service *Service) ChangeWaveAddress(accountID, currentCode, localPart string) (account.Account, error) {
	if err := service.totp.Verify(accountID, strings.TrimSpace(currentCode)); err != nil {
		return account.Account{}, ErrInvalidCredentials
	}
	localPart, err := account.MailLocalPart(localPart)
	if err != nil {
		return account.Account{}, fmt.Errorf("%w: %v", ErrInvalidProfile, err)
	}
	address, err := account.Address(localPart, service.mailDomain)
	if err != nil {
		return account.Account{}, fmt.Errorf("%w: invalid Wave address", ErrInvalidProfile)
	}
	item, err := service.accounts.Account(accountID)
	if err != nil {
		return account.Account{}, err
	}
	if strings.EqualFold(item.Email, address) {
		return item, nil
	}
	if !service.AddressChoiceAllowed(item) {
		return account.Account{}, fmt.Errorf("%w: a custom mail address is available only for duplicate display names", ErrInvalidProfile)
	}
	if _, lookupErr := service.accounts.ByEmail(address); lookupErr == nil {
		return account.Account{}, account.ErrConflict
	} else if !errors.Is(lookupErr, storage.ErrNotFound) {
		return account.Account{}, lookupErr
	}
	if _, lookupErr := service.accounts.ByUsername(localPart); lookupErr == nil {
		return account.Account{}, account.ErrConflict
	} else if !errors.Is(lookupErr, storage.ErrNotFound) {
		return account.Account{}, lookupErr
	}
	box, err := service.mailboxes.MailboxByAccount(accountID)
	if err != nil {
		return account.Account{}, err
	}
	previous := item
	previousLocalPart := item.Username
	previousAddress := box.Address
	item.Username = localPart
	item.Email = address
	item.UpdatedAt = service.now().UTC()
	if err := service.accounts.Update(item); err != nil {
		return account.Account{}, err
	}
	box.Address = address
	if err := service.mailboxes.UpdateAddress(box, previousAddress); err != nil {
		_ = service.accounts.Update(previous)
		return account.Account{}, err
	}
	_ = service.accounts.AddProfileAlias(previousLocalPart, item.ID)
	_ = service.appendAudit(accountID, "account/"+accountID+"/address", "account.address.update")
	return item, nil
}

func (service *Service) AddressChoiceAllowed(item account.Account) bool {
	generated, err := account.LocalPart(item.DisplayName)
	if err != nil || account.IsReservedLocalPart(generated) {
		return false
	}
	existing, err := service.accounts.ByUsername(generated)
	return err == nil && existing.ID != item.ID
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
	_, err = service.sendSystemMail(email, systemMailContent{
		Subject:     "Confirm your Wave recovery email",
		Heading:     "Confirm your recovery email",
		Intro:       "Use the button below to confirm this email address for your Wave account.",
		ActionLabel: "Confirm recovery email",
		ActionURL:   link,
		Expiry:      "This link expires in 24 hours.",
		Ignore:      "If you did not create or modify a Wave account, you can safely ignore this email.",
	})
	return err
}

func (service *Service) accountByIdentifier(identifier string) (account.Account, error) {
	identifier = strings.ToLower(strings.TrimSpace(identifier))
	if strings.Contains(identifier, "@") {
		return service.accounts.ByEmail(identifier)
	}
	return service.accounts.ByUsername(identifier)
}

func (service *Service) PublicAccount(localPart string) (account.Account, error) {
	localPart = strings.ToLower(strings.TrimSpace(localPart))
	item, err := service.accounts.ByEmail(localPart + "@" + service.mailDomain)
	if errors.Is(err, storage.ErrNotFound) {
		item, err = service.accounts.ByUsername(localPart)
	}
	if errors.Is(err, storage.ErrNotFound) {
		return service.accounts.ByProfileAlias(localPart)
	}
	return item, err
}

func (service *Service) PublicAccountByID(accountID string) (account.Account, error) {
	return service.accounts.Account(strings.TrimSpace(accountID))
}

func (service *Service) PublicAccounts() ([]account.Account, error) {
	items, err := service.accounts.Accounts()
	if err != nil {
		return nil, err
	}
	active := make([]account.Account, 0, len(items))
	for _, item := range items {
		if item.Status == "active" {
			active = append(active, item)
		}
	}
	sort.SliceStable(active, func(left, right int) bool {
		return strings.ToLower(active[left].Email) < strings.ToLower(active[right].Email)
	})
	return active, nil
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
			if _, aliasErr := service.accounts.ByProfileAlias(candidate); errors.Is(aliasErr, storage.ErrNotFound) {
				return candidate
			}
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

func (service *Service) EnsureManagementMailbox() (mailbox.Mailbox, error) {
	addresses := service.ManagementAddresses()
	if len(addresses) == 0 {
		return mailbox.Mailbox{}, errors.New("management mailbox has no addresses")
	}
	box, err := service.mailboxes.MailboxByAccount(ManagementMailboxAccountID)
	if errors.Is(err, storage.ErrNotFound) {
		id, idErr := identifier.New("mailbox")
		if idErr != nil {
			return mailbox.Mailbox{}, idErr
		}
		box = mailbox.Mailbox{ID: id, AccountID: ManagementMailboxAccountID, Address: addresses[0], CreatedAt: service.now().UTC()}
		if err := service.mailboxes.UpsertMailbox(box); err != nil {
			return mailbox.Mailbox{}, err
		}
	} else if err != nil {
		return mailbox.Mailbox{}, err
	}
	for _, address := range addresses {
		if err := service.mailboxes.AddAddress(box.ID, address); err != nil {
			return mailbox.Mailbox{}, err
		}
	}
	return box, nil
}

func (service *Service) ManagementAddresses() []string {
	addresses := make([]string, 0, len(managementLocalParts))
	for _, local := range managementLocalParts {
		addresses = append(addresses, local+"@"+service.mailDomain)
	}
	return addresses
}

func (service *Service) PatchAddress() string { return "patchs@" + service.mailDomain }

func (service *Service) EnsurePatchMailbox() (mailbox.Mailbox, error) {
	address := service.PatchAddress()
	box, err := service.mailboxes.MailboxByAccount(PatchMailboxAccountID)
	if errors.Is(err, storage.ErrNotFound) {
		id, idErr := identifier.New("mailbox")
		if idErr != nil {
			return mailbox.Mailbox{}, idErr
		}
		box = mailbox.Mailbox{ID: id, AccountID: PatchMailboxAccountID, Address: address, CreatedAt: service.now().UTC()}
		if err := service.mailboxes.UpsertMailbox(box); err != nil {
			return mailbox.Mailbox{}, err
		}
	} else if err != nil {
		return mailbox.Mailbox{}, err
	}
	if err := service.mailboxes.AddAddress(box.ID, address); err != nil {
		return mailbox.Mailbox{}, err
	}
	return box, nil
}

func (service *Service) ManagementMailboxItems(folder string) ([]MailboxItem, error) {
	box, err := service.EnsureManagementMailbox()
	if err != nil {
		return nil, err
	}
	return service.mailboxItems(box, folder)
}

func (service *Service) ManagementMailboxItem(entryID string) (MailboxItem, error) {
	box, err := service.EnsureManagementMailbox()
	if err != nil {
		return MailboxItem{}, err
	}
	return service.mailboxItem(box, entryID)
}

func (service *Service) UpdateManagementMailboxEntry(actorID, entryID, action string) (MailboxItem, error) {
	if !service.IsAdministrator(actorID) {
		return MailboxItem{}, ErrRelayDenied
	}
	item, err := service.ManagementMailboxItem(entryID)
	if err != nil {
		return MailboxItem{}, err
	}
	if err := service.applyMailboxAction(&item, action); err != nil {
		return MailboxItem{}, err
	}
	_ = service.appendAudit(actorID, "mailbox/management/entry/"+entryID, "admin.mailbox."+action)
	return item, nil
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
	return service.mailboxItems(box, folder)
}

func (service *Service) mailboxItems(box mailbox.Mailbox, folder string) ([]MailboxItem, error) {
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
	return service.mailboxItem(box, entryID)
}

func (service *Service) mailboxItem(box mailbox.Mailbox, entryID string) (MailboxItem, error) {
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

func (service *Service) sendSystemMail(to string, content systemMailContent) (MailboxItem, error) {
	recipients, err := normalizeRecipients(to)
	if err != nil {
		return MailboxItem{}, fmt.Errorf("%w: %v", ErrInvalidMail, err)
	}
	content.Subject = strings.TrimSpace(content.Subject)
	if content.Subject == "" || strings.ContainsAny(content.Subject, "\r\n") {
		return MailboxItem{}, ErrInvalidMail
	}
	plainBody, htmlBody, err := renderSystemMail(content, service.baseURL())
	if err != nil {
		return MailboxItem{}, fmt.Errorf("%w: %v", ErrInvalidMail, err)
	}
	messageID, err := identifier.New("message")
	if err != nil {
		return MailboxItem{}, err
	}
	now := service.now().UTC()
	from := "security@" + service.mailDomain
	message := maildomain.Message{ID: messageID, MessageID: "<" + messageID + "@" + service.mailDomain + ">",
		ThreadID: messageID, AuthorAccountID: "system/security", From: (&stdmail.Address{Name: "Wave Security", Address: from}).String(),
		To: recipients, Subject: content.Subject, ReceivedAt: now, CreatedAt: now}
	raw, err := encodeSystemMail(message, plainBody, htmlBody)
	if err != nil {
		return MailboxItem{}, err
	}
	if err := service.mail.UpsertMessage(message, raw); err != nil {
		return MailboxItem{}, err
	}
	if err := service.routeMessage(from, message, recipients, true); err != nil {
		return MailboxItem{}, err
	}
	return MailboxItem{Message: message, Body: plainBody, DeliveryStatus: service.deliveryStatus(message.ID)}, nil
}

// AcceptSMTP stores an RFC 5322/MIME message received through server-to-server
// SMTP. Internet ingress is restricted to local Wave mailbox recipients.
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

	subject := decodeMIMEHeader(parsed.Header.Get("Subject"))
	from := decodeMIMEHeader(parsed.Header.Get("From"))

	if parsedFrom, err := stdmail.ParseAddress(from); err == nil {
		if parsedFrom.Name != "" {
			from = fmt.Sprintf("%s <%s>", parsedFrom.Name, parsedFrom.Address)
		} else {
			from = parsedFrom.Address
		}
	}

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
	message, err = service.mail.Message(message.ID)
	if err != nil {
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
	body, err := service.mail.Body(message)
	if err != nil {
		return MailboxItem{}, err
	}
	if actor == nil && service.webhooks != nil && containsAddress(normalized, service.PatchAddress()) && patchdomain.Valid(message.Subject, body) {
		authorName := message.From
		if parsedFrom, parseErr := stdmail.ParseAddress(message.From); parseErr == nil {
			authorName = parsedFrom.Name
			if authorName == "" {
				authorName = parsedFrom.Address
			}
		}
		_ = service.webhooks.Publish(webhookdomain.Event{Type: webhookdomain.EventPatchReceived, Title: message.Subject,
			Summary: body, AuthorName: authorName,
			ResourceID: "patch/" + message.ID, URL: "/patches/" + message.ID, OccurredAt: now})
	}
	return MailboxItem{Entry: sentEntry, Message: message, Body: body,
		DeliveryStatus: service.deliveryStatus(message.ID)}, nil
}

func containsAddress(values []string, target string) bool {
	for _, value := range values {
		if strings.EqualFold(value, target) {
			return true
		}
	}
	return false
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
	if err := service.applyMailboxAction(&item, action); err != nil {
		return MailboxItem{}, err
	}
	return item, nil
}

func (service *Service) applyMailboxAction(item *MailboxItem, action string) error {
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
		return fmt.Errorf("%w: unsupported mailbox action", ErrInvalidMail)
	}
	if err := service.mailboxes.UpdateEntry(item.Entry); err != nil {
		return err
	}
	return nil
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

func renderSystemMail(content systemMailContent, platformURL string) (string, string, error) {
	for _, value := range []string{content.Heading, content.Intro, content.ActionLabel, content.ActionURL, content.Expiry, content.Ignore} {
		if strings.TrimSpace(value) == "" {
			return "", "", errors.New("system mail content is incomplete")
		}
	}
	actionURL, err := url.Parse(content.ActionURL)
	if err != nil || (actionURL.Scheme != "http" && actionURL.Scheme != "https") || actionURL.Host == "" {
		return "", "", errors.New("system mail action URL is invalid")
	}
	platformURL = strings.TrimRight(platformURL, "/")
	plain := strings.Join([]string{
		"Wave",
		"",
		content.Heading,
		"",
		content.Intro,
		"",
		content.ActionLabel + ":",
		content.ActionURL,
		"",
		content.Expiry,
		"",
		content.Ignore,
		"",
		"Wave Platform",
		platformURL,
	}, "\n")
	escape := html.EscapeString
	htmlBody := `<!doctype html>
<html lang="en">
<head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1"><title>` + escape(content.Subject) + `</title></head>
<body style="margin:0;padding:0;background:#f3f4f6;color:#1f2328;font-family:Arial,sans-serif">
<table role="presentation" width="100%" cellspacing="0" cellpadding="0" style="background:#f3f4f6;padding:32px 16px"><tr><td align="center">
<table role="presentation" width="100%" cellspacing="0" cellpadding="0" style="max-width:560px;background:#ffffff;border:1px solid #d8dbe2"><tr><td style="padding:32px">
<div style="margin-bottom:28px;color:#6654f1;font-size:22px;font-weight:700">Wave</div>
<h1 style="margin:0 0 16px;color:#1f2328;font-size:24px;line-height:1.3">` + escape(content.Heading) + `</h1>
<p style="margin:0 0 24px;color:#5f6672;font-size:15px;line-height:1.6">` + escape(content.Intro) + `</p>
<table role="presentation" cellspacing="0" cellpadding="0"><tr><td style="background:#6654f1"><a href="` + escape(content.ActionURL) + `" style="display:inline-block;padding:12px 18px;color:#ffffff;font-size:14px;font-weight:700;text-decoration:none">` + escape(content.ActionLabel) + `</a></td></tr></table>
<p style="margin:24px 0 0;color:#5f6672;font-size:13px;line-height:1.6">` + escape(content.Expiry) + `</p>
<p style="margin:8px 0 0;color:#7b828e;font-size:13px;line-height:1.6">` + escape(content.Ignore) + `</p>
<div style="margin-top:28px;padding-top:20px;border-top:1px solid #d8dbe2;color:#7b828e;font-size:12px;line-height:1.6">Wave Platform<br><a href="` + escape(platformURL) + `" style="color:#5b47e8;text-decoration:none">` + escape(platformURL) + `</a></div>
</td></tr></table>
</td></tr></table>
</body>
</html>`
	return plain, htmlBody, nil
}

func encodeSystemMail(message maildomain.Message, plainBody, htmlBody string) ([]byte, error) {
	var content bytes.Buffer
	writer := multipart.NewWriter(&content)
	for _, part := range []struct {
		contentType string
		body        string
	}{
		{contentType: "text/plain; charset=utf-8", body: plainBody},
		{contentType: "text/html; charset=utf-8", body: htmlBody},
	} {
		header := textproto.MIMEHeader{}
		header.Set("Content-Type", part.contentType)
		header.Set("Content-Transfer-Encoding", "quoted-printable")
		bodyWriter, err := writer.CreatePart(header)
		if err != nil {
			return nil, err
		}
		encoded := quotedprintable.NewWriter(bodyWriter)
		if _, err := encoded.Write([]byte(strings.ReplaceAll(part.body, "\n", "\r\n"))); err != nil {
			return nil, err
		}
		if err := encoded.Close(); err != nil {
			return nil, err
		}
	}
	if err := writer.Close(); err != nil {
		return nil, err
	}
	headers := textproto.MIMEHeader{}
	headers.Set("From", message.From)
	headers.Set("To", strings.Join(message.To, ", "))
	headers.Set("Subject", mime.QEncoding.Encode("utf-8", message.Subject))
	headers.Set("Date", message.CreatedAt.Format(time.RFC1123Z))
	headers.Set("Message-ID", message.MessageID)
	headers.Set("MIME-Version", "1.0")
	headers.Set("Auto-Submitted", "auto-generated")
	headers.Set("X-Auto-Response-Suppress", "All")
	headers.Set("Content-Type", mime.FormatMediaType("multipart/alternative", map[string]string{"boundary": writer.Boundary()}))
	order := []string{"From", "To", "Subject", "Date", "Message-ID", "MIME-Version", "Auto-Submitted", "X-Auto-Response-Suppress", "Content-Type"}
	var raw strings.Builder
	for _, name := range order {
		raw.WriteString(name + ": " + headers.Get(name) + "\r\n")
	}
	raw.WriteString("\r\n")
	raw.Write(content.Bytes())
	return []byte(raw.String()), nil
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

func decodeMIMEHeader(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}

	decoded, err := (&mime.WordDecoder{}).DecodeHeader(value)
	if err != nil {
		return value
	}

	return decoded
}
