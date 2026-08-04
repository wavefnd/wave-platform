package auth

import (
	"encoding/base64"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/pquerna/otp/totp"
	"github.com/wavefnd/wave-platform/internal/storage"
)

func TestTOTPFactorEncryptsSecretAndRejectsReplay(t *testing.T) {
	database, err := storage.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	service, err := NewTOTPService(database, base64.StdEncoding.EncodeToString([]byte("0123456789abcdef0123456789abcdef")), "Wave Test")
	if err != nil {
		t.Fatal(err)
	}
	now := time.Unix(1_800_000_000, 0).UTC()
	service.now = func() time.Time { return now }
	result, err := service.BeginEnrollment("registration", "account-1", "", "", "recover@example.com", "user@wave.test")
	if err != nil {
		t.Fatal(err)
	}
	code, err := totp.GenerateCode(result.Secret, now)
	if err != nil {
		t.Fatal(err)
	}
	item, secret, err := service.Enrollment(result.Token, code)
	if err != nil {
		t.Fatal(err)
	}
	if err := service.PutFactorFromEnrollment("account-1", secret, item.RecoveryEmail, false, code); err != nil {
		t.Fatal(err)
	}
	factor, err := service.Factor("account-1")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(factor.EncryptedSecret, result.Secret) || factor.EncryptedSecret == result.Secret {
		t.Fatal("TOTP secret was stored in plaintext")
	}
	if err := service.Verify("account-1", code); !errors.Is(err, ErrTOTPReplay) {
		t.Fatalf("replay error = %v", err)
	}
	now = now.Add(30 * time.Second)
	nextCode, _ := totp.GenerateCode(result.Secret, now)
	if err := service.Verify("account-1", nextCode); err != nil {
		t.Fatalf("next code: %v", err)
	}
}

func TestNormalizeRecoveryEmailRejectsDisplayName(t *testing.T) {
	if _, err := NormalizeRecoveryEmail("User <user@example.com>"); err == nil {
		t.Fatal("display-name address should be rejected")
	}
}
