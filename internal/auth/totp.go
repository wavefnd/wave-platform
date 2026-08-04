package auth

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base32"
	"encoding/base64"
	"encoding/hex"
	"encoding/xml"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"sync"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/hotp"
	"github.com/pquerna/otp/totp"
	"github.com/wavefnd/wave-platform/internal/storage"
)

var (
	ErrTOTPNotConfigured = errors.New("TOTP authentication is not configured")
	ErrInvalidTOTP       = errors.New("invalid TOTP code")
	ErrTOTPReplay        = errors.New("TOTP code was already used")
	ErrTemporarilyLocked = errors.New("authentication temporarily locked")
	ErrEnrollmentExpired = errors.New("TOTP enrollment expired")
	ErrRecoveryExpired   = errors.New("recovery request expired")
)

const maxFailures = 5

type TOTPFactor struct {
	XMLName          xml.Name  `xml:"https://wave-lang.dev/ns/platform/auth/v1 totp-factor"`
	AccountID        string    `xml:"account-id,attr"`
	EncryptedSecret  string    `xml:"encrypted-secret"`
	RecoveryEmail    string    `xml:"recovery-email"`
	RecoveryVerified bool      `xml:"recovery-verified"`
	FailedCount      int       `xml:"failed-count"`
	LockedUntil      time.Time `xml:"locked-until,omitempty"`
	LastUsedStep     uint64    `xml:"last-used-step,omitempty"`
	EnrolledAt       time.Time `xml:"enrolled-at"`
	UpdatedAt        time.Time `xml:"updated-at"`
}

type TOTPEnrollment struct {
	XMLName         xml.Name  `xml:"https://wave-lang.dev/ns/platform/auth/v1 totp-enrollment"`
	TokenHash       string    `xml:"token-hash,attr"`
	Kind            string    `xml:"kind"`
	AccountID       string    `xml:"account-id,omitempty"`
	DisplayName     string    `xml:"display-name,omitempty"`
	Username        string    `xml:"username,omitempty"`
	RecoveryEmail   string    `xml:"recovery-email"`
	EncryptedSecret string    `xml:"encrypted-secret"`
	CreatedAt       time.Time `xml:"created-at"`
	ExpiresAt       time.Time `xml:"expires-at"`
}

type RecoveryRequest struct {
	XMLName         xml.Name  `xml:"https://wave-lang.dev/ns/platform/auth/v1 recovery-request"`
	TokenHash       string    `xml:"token-hash,attr"`
	AccountID       string    `xml:"account-id"`
	EncryptedSecret string    `xml:"encrypted-secret"`
	CreatedAt       time.Time `xml:"created-at"`
	ExpiresAt       time.Time `xml:"expires-at"`
}

type EmailVerification struct {
	XMLName   xml.Name  `xml:"https://wave-lang.dev/ns/platform/auth/v1 email-verification"`
	TokenHash string    `xml:"token-hash,attr"`
	AccountID string    `xml:"account-id"`
	Email     string    `xml:"email"`
	CreatedAt time.Time `xml:"created-at"`
	ExpiresAt time.Time `xml:"expires-at"`
}

type EnrollmentResult struct {
	Token     string
	Secret    string
	URI       string
	ExpiresAt time.Time
}

type TOTPService struct {
	database *storage.Database
	aead     cipher.AEAD
	issuer   string
	now      func() time.Time
	mu       sync.Mutex
}

func NewTOTPService(database *storage.Database, encodedKey, issuer string) (*TOTPService, error) {
	service := &TOTPService{database: database, issuer: strings.TrimSpace(issuer), now: time.Now}
	if service.issuer == "" {
		service.issuer = "Wave Platform"
	}
	if strings.TrimSpace(encodedKey) == "" {
		return service, nil
	}
	key, err := base64.StdEncoding.DecodeString(strings.TrimSpace(encodedKey))
	if err != nil || len(key) != 32 {
		return nil, errors.New("WAVE_AUTH_ENCRYPTION_KEY must be a base64-encoded 32-byte key")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	service.aead, err = cipher.NewGCM(block)
	return service, err
}

func (service *TOTPService) Configured() bool { return service != nil && service.aead != nil }

func NormalizeRecoveryEmail(value string) (string, error) {
	value = strings.ToLower(strings.TrimSpace(value))
	parsed, err := mail.ParseAddress(value)
	if err != nil || parsed.Address != value || len(value) > 254 || !strings.Contains(value, "@") {
		return "", errors.New("a valid recovery email address is required")
	}
	return value, nil
}

func (service *TOTPService) BeginEnrollment(kind, accountID, displayName, username, recoveryEmail, accountName string) (EnrollmentResult, error) {
	if !service.Configured() {
		return EnrollmentResult{}, ErrTOTPNotConfigured
	}
	recoveryEmail, err := NormalizeRecoveryEmail(recoveryEmail)
	if err != nil {
		return EnrollmentResult{}, err
	}
	key, err := totp.Generate(totp.GenerateOpts{Issuer: service.issuer, AccountName: accountName, Period: 30,
		SecretSize: 32, Digits: otp.DigitsSix, Algorithm: otp.AlgorithmSHA1})
	if err != nil {
		return EnrollmentResult{}, err
	}
	encrypted, err := service.encrypt(key.Secret())
	if err != nil {
		return EnrollmentResult{}, err
	}
	token, tokenHash, err := randomToken()
	if err != nil {
		return EnrollmentResult{}, err
	}
	now := service.now().UTC()
	item := TOTPEnrollment{TokenHash: tokenHash, Kind: kind, AccountID: accountID, DisplayName: displayName,
		Username: username, RecoveryEmail: recoveryEmail, EncryptedSecret: encrypted, CreatedAt: now, ExpiresAt: now.Add(10 * time.Minute)}
	if err := service.putTTL("enrollment", tokenHash, item, 10*time.Minute); err != nil {
		return EnrollmentResult{}, err
	}
	return EnrollmentResult{Token: token, Secret: key.Secret(), URI: key.URL(), ExpiresAt: item.ExpiresAt}, nil
}

func (service *TOTPService) Enrollment(token, code string) (TOTPEnrollment, string, error) {
	var item TOTPEnrollment
	if err := service.get("enrollment", tokenDigest(token), &item); err != nil {
		return item, "", ErrEnrollmentExpired
	}
	if !item.ExpiresAt.After(service.now().UTC()) {
		return item, "", ErrEnrollmentExpired
	}
	secret, err := service.decrypt(item.EncryptedSecret)
	if err != nil {
		return item, "", err
	}
	if _, err := validStep(secret, code, service.now().UTC()); err != nil {
		return item, "", err
	}
	return item, secret, nil
}

func (service *TOTPService) DeleteEnrollment(token string) error {
	return service.database.Delete(storage.Key("auth", "totp", "enrollment", tokenDigest(token)))
}

func (service *TOTPService) PutFactor(accountID, secret, recoveryEmail string, verified bool) error {
	return service.putFactor(accountID, secret, recoveryEmail, verified, 0)
}

func (service *TOTPService) PutFactorFromEnrollment(accountID, secret, recoveryEmail string, verified bool, code string) error {
	step, err := validStep(secret, code, service.now().UTC())
	if err != nil {
		return err
	}
	return service.putFactor(accountID, secret, recoveryEmail, verified, step)
}

func (service *TOTPService) putFactor(accountID, secret, recoveryEmail string, verified bool, lastUsedStep uint64) error {
	if !service.Configured() {
		return ErrTOTPNotConfigured
	}
	secret, err := normalizeSecret(secret)
	if err != nil {
		return err
	}
	recoveryEmail, err = NormalizeRecoveryEmail(recoveryEmail)
	if err != nil {
		return err
	}
	encrypted, err := service.encrypt(secret)
	if err != nil {
		return err
	}
	now := service.now().UTC()
	previous, _ := service.Factor(accountID)
	enrolledAt := previous.EnrolledAt
	if enrolledAt.IsZero() {
		enrolledAt = now
	}
	return service.put("factor", accountID, TOTPFactor{AccountID: accountID, EncryptedSecret: encrypted,
		RecoveryEmail: recoveryEmail, RecoveryVerified: verified, LastUsedStep: lastUsedStep, EnrolledAt: enrolledAt, UpdatedAt: now})
}

func (service *TOTPService) Factor(accountID string) (TOTPFactor, error) {
	var item TOTPFactor
	err := service.get("factor", accountID, &item)
	return item, err
}

func (service *TOTPService) DeleteFactor(accountID string) error {
	return service.database.Delete(storage.Key("auth", "totp", "factor", accountID))
}

func (service *TOTPService) Verify(accountID, code string) error {
	if !service.Configured() {
		return ErrTOTPNotConfigured
	}
	service.mu.Lock()
	defer service.mu.Unlock()
	factor, err := service.Factor(accountID)
	if err != nil {
		return ErrInvalidTOTP
	}
	now := service.now().UTC()
	if factor.LockedUntil.After(now) {
		return ErrTemporarilyLocked
	}
	secret, err := service.decrypt(factor.EncryptedSecret)
	if err != nil {
		return err
	}
	step, err := validStep(secret, code, now)
	if err != nil {
		factor.FailedCount++
		if factor.FailedCount >= maxFailures {
			factor.FailedCount = 0
			factor.LockedUntil = now.Add(15 * time.Minute)
		}
		_ = service.put("factor", accountID, factor)
		return ErrInvalidTOTP
	}
	if step <= factor.LastUsedStep {
		return ErrTOTPReplay
	}
	factor.FailedCount = 0
	factor.LockedUntil = time.Time{}
	factor.LastUsedStep = step
	factor.UpdatedAt = now
	return service.put("factor", accountID, factor)
}

func (service *TOTPService) BeginRecovery(accountID string) (string, RecoveryRequest, EnrollmentResult, error) {
	factor, err := service.Factor(accountID)
	if err != nil || !factor.RecoveryVerified {
		return "", RecoveryRequest{}, EnrollmentResult{}, storage.ErrNotFound
	}
	key, err := totp.Generate(totp.GenerateOpts{Issuer: service.issuer, AccountName: factor.RecoveryEmail,
		Period: 30, SecretSize: 32, Digits: otp.DigitsSix, Algorithm: otp.AlgorithmSHA1})
	if err != nil {
		return "", RecoveryRequest{}, EnrollmentResult{}, err
	}
	encrypted, err := service.encrypt(key.Secret())
	if err != nil {
		return "", RecoveryRequest{}, EnrollmentResult{}, err
	}
	token, tokenHash, err := randomToken()
	if err != nil {
		return "", RecoveryRequest{}, EnrollmentResult{}, err
	}
	now := service.now().UTC()
	request := RecoveryRequest{TokenHash: tokenHash, AccountID: accountID, EncryptedSecret: encrypted,
		CreatedAt: now, ExpiresAt: now.Add(30 * time.Minute)}
	if err := service.putTTL("recovery", tokenHash, request, 30*time.Minute); err != nil {
		return "", RecoveryRequest{}, EnrollmentResult{}, err
	}
	return factor.RecoveryEmail, request, EnrollmentResult{Token: token, Secret: key.Secret(), URI: key.URL(), ExpiresAt: request.ExpiresAt}, nil
}

func (service *TOTPService) Recovery(token string) (RecoveryRequest, EnrollmentResult, error) {
	var item RecoveryRequest
	if err := service.get("recovery", tokenDigest(token), &item); err != nil || !item.ExpiresAt.After(service.now().UTC()) {
		return item, EnrollmentResult{}, ErrRecoveryExpired
	}
	secret, err := service.decrypt(item.EncryptedSecret)
	if err != nil {
		return item, EnrollmentResult{}, err
	}
	key, err := totp.Generate(totp.GenerateOpts{Issuer: service.issuer, AccountName: item.AccountID,
		Period: 30, Secret: mustDecodeSecret(secret), Digits: otp.DigitsSix, Algorithm: otp.AlgorithmSHA1})
	if err != nil {
		return item, EnrollmentResult{}, err
	}
	return item, EnrollmentResult{Token: token, Secret: secret, URI: key.URL(), ExpiresAt: item.ExpiresAt}, nil
}

func (service *TOTPService) CompleteRecovery(token, code string) (string, error) {
	request, result, err := service.Recovery(token)
	if err != nil {
		return "", err
	}
	if _, err := validStep(result.Secret, code, service.now().UTC()); err != nil {
		return "", ErrInvalidTOTP
	}
	factor, err := service.Factor(request.AccountID)
	if err != nil {
		return "", err
	}
	if err := service.PutFactorFromEnrollment(request.AccountID, result.Secret, factor.RecoveryEmail, factor.RecoveryVerified, code); err != nil {
		return "", err
	}
	_ = service.database.Delete(storage.Key("auth", "totp", "recovery", tokenDigest(token)))
	return request.AccountID, nil
}

func (service *TOTPService) BeginEmailVerification(accountID, email string) (string, error) {
	email, err := NormalizeRecoveryEmail(email)
	if err != nil {
		return "", err
	}
	token, tokenHash, err := randomToken()
	if err != nil {
		return "", err
	}
	now := service.now().UTC()
	item := EmailVerification{TokenHash: tokenHash, AccountID: accountID, Email: email,
		CreatedAt: now, ExpiresAt: now.Add(24 * time.Hour)}
	if err := service.putTTL("email-verification", tokenHash, item, 24*time.Hour); err != nil {
		return "", err
	}
	return token, nil
}

func (service *TOTPService) CompleteEmailVerification(token string) (EmailVerification, error) {
	var item EmailVerification
	if err := service.get("email-verification", tokenDigest(token), &item); err != nil || !item.ExpiresAt.After(service.now().UTC()) {
		return item, storage.ErrNotFound
	}
	factor, err := service.Factor(item.AccountID)
	if err != nil {
		return item, err
	}
	factor.RecoveryEmail = item.Email
	factor.RecoveryVerified = true
	factor.UpdatedAt = service.now().UTC()
	if err := service.put("factor", item.AccountID, factor); err != nil {
		return item, err
	}
	_ = service.database.Delete(storage.Key("auth", "totp", "email-verification", tokenDigest(token)))
	return item, nil
}

func (service *TOTPService) encrypt(value string) (string, error) {
	if !service.Configured() {
		return "", ErrTOTPNotConfigured
	}
	nonce := make([]byte, service.aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}
	sealed := service.aead.Seal(nonce, nonce, []byte(value), nil)
	return base64.RawURLEncoding.EncodeToString(sealed), nil
}

func (service *TOTPService) decrypt(value string) (string, error) {
	data, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil || len(data) < service.aead.NonceSize() {
		return "", errors.New("invalid encrypted TOTP secret")
	}
	plain, err := service.aead.Open(nil, data[:service.aead.NonceSize()], data[service.aead.NonceSize():], nil)
	return string(plain), err
}

func (service *TOTPService) put(kind, id string, value any) error {
	data, err := xml.Marshal(value)
	if err != nil {
		return err
	}
	return service.database.Set(storage.Key("auth", "totp", kind, id), data)
}

func (service *TOTPService) putTTL(kind, id string, value any, ttl time.Duration) error {
	data, err := xml.Marshal(value)
	if err != nil {
		return err
	}
	return service.database.DB.Update(func(transaction *badger.Txn) error {
		return transaction.SetEntry(badger.NewEntry(storage.Key("auth", "totp", kind, id), data).WithTTL(ttl))
	})
}

func (service *TOTPService) get(kind, id string, destination any) error {
	data, err := service.database.Get(storage.Key("auth", "totp", kind, id))
	if err != nil {
		return err
	}
	return xml.Unmarshal(data, destination)
}

func validStep(secret, code string, now time.Time) (uint64, error) {
	current := uint64(now.Unix() / 30)
	options := hotp.ValidateOpts{Digits: otp.DigitsSix, Algorithm: otp.AlgorithmSHA1}
	for _, step := range []uint64{current, current + 1, current - 1} {
		valid, err := hotp.ValidateCustom(strings.TrimSpace(code), step, secret, options)
		if err == nil && valid {
			return step, nil
		}
	}
	return 0, ErrInvalidTOTP
}

func randomToken() (string, string, error) {
	data := make([]byte, 32)
	if _, err := rand.Read(data); err != nil {
		return "", "", err
	}
	token := base64.RawURLEncoding.EncodeToString(data)
	return token, tokenDigest(token), nil
}

func tokenDigest(token string) string {
	digest := sha256.Sum256([]byte(strings.TrimSpace(token)))
	return hex.EncodeToString(digest[:])
}

func mustDecodeSecret(secret string) []byte {
	data, _ := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(strings.ToUpper(secret))
	return data
}

func normalizeSecret(secret string) (string, error) {
	secret = strings.ToUpper(strings.TrimSpace(strings.ReplaceAll(secret, " ", "")))
	secret = strings.TrimRight(secret, "=")
	data, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(secret)
	if err != nil || len(data) < 16 {
		return "", errors.New("TOTP secret must be valid Base32 with at least 128 bits")
	}
	return secret, nil
}

func MaskEmail(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) != 2 || len(parts[0]) < 2 {
		return "***"
	}
	return fmt.Sprintf("%c***%c@%s", parts[0][0], parts[0][len(parts[0])-1], parts[1])
}
