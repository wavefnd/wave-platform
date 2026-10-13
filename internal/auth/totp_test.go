package auth

import (
	"encoding/base64"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/dgraph-io/badger/v4"
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

func TestVerifySurfacesFailedAttemptPersistenceErrors(t *testing.T) {
	for _, failedCount := range []int{0, maxFailures - 1} {
		t.Run(fmt.Sprintf("failed-count-%d", failedCount), func(t *testing.T) {
			database, err := storage.Open(t.TempDir())
			if err != nil {
				t.Fatal(err)
			}
			defer database.Close()

			service, err := NewTOTPService(database, base64.StdEncoding.EncodeToString([]byte("0123456789abcdef0123456789abcdef")), "Wave Test")
			if err != nil {
				t.Fatal(err)
			}
			if err := service.PutFactor("account-1", "JBSWY3DPEHPK3PXPJBSWY3DPEHPK3PXP", "recover@example.com", false); err != nil {
				t.Fatal(err)
			}
			factor, err := service.Factor("account-1")
			if err != nil {
				t.Fatal(err)
			}
			factor.FailedCount = failedCount
			if err := service.put("factor", "account-1", factor); err != nil {
				t.Fatal(err)
			}

			var closeErr error
			service.now = func() time.Time {
				closeErr = database.Close()
				return time.Unix(1_800_000_000, 0).UTC()
			}

			err = service.Verify("account-1", "not-a-code")
			if closeErr != nil {
				t.Fatal(closeErr)
			}
			if err == nil || errors.Is(err, ErrInvalidTOTP) || !strings.Contains(err.Error(), "TOTP state persistence failed") {
				t.Fatalf("Verify() = %v; want a persistence error, not an ordinary invalid-code error", err)
			}
			writable, err := badger.Open(badger.DefaultOptions(filepath.Join(database.Root, "badger")).WithLogger(nil))
			if err != nil {
				t.Fatal(err)
			}
			database.DB = writable
			stored, err := service.Factor("account-1")
			if err != nil {
				t.Fatal(err)
			}
			if stored.FailedCount != failedCount {
				t.Fatalf("failed count = %d, want unchanged %d after failed write", stored.FailedCount, failedCount)
			}
		})
	}
}

func TestVerifyPersistsFailedAttemptsAndLockoutThreshold(t *testing.T) {
	for _, failedCount := range []int{0, maxFailures - 1} {
		t.Run(fmt.Sprintf("failed-count-%d", failedCount), func(t *testing.T) {
			database, err := storage.Open(t.TempDir())
			if err != nil {
				t.Fatal(err)
			}
			defer database.Close()

			service, err := NewTOTPService(database, base64.StdEncoding.EncodeToString([]byte("0123456789abcdef0123456789abcdef")), "Wave Test")
			if err != nil {
				t.Fatal(err)
			}
			if err := service.PutFactor("account-1", "JBSWY3DPEHPK3PXPJBSWY3DPEHPK3PXP", "recover@example.com", false); err != nil {
				t.Fatal(err)
			}
			factor, err := service.Factor("account-1")
			if err != nil {
				t.Fatal(err)
			}
			factor.FailedCount = failedCount
			if err := service.put("factor", "account-1", factor); err != nil {
				t.Fatal(err)
			}

			now := time.Unix(1_800_000_000, 0).UTC()
			service.now = func() time.Time { return now }
			if err := service.Verify("account-1", "not-a-code"); !errors.Is(err, ErrInvalidTOTP) {
				t.Fatalf("Verify() = %v; want ErrInvalidTOTP after a successful write", err)
			}
			stored, err := service.Factor("account-1")
			if err != nil {
				t.Fatal(err)
			}
			if failedCount < maxFailures-1 {
				if stored.FailedCount != failedCount+1 || !stored.LockedUntil.IsZero() {
					t.Fatalf("stored factor = %#v; want incremented count without lockout", stored)
				}
			} else if stored.FailedCount != 0 || !stored.LockedUntil.Equal(now.Add(15*time.Minute)) {
				t.Fatalf("stored factor = %#v; want reset count and a 15-minute lockout", stored)
			}
		})
	}
}

func TestNormalizeRecoveryEmailRejectsDisplayName(t *testing.T) {
	if _, err := NormalizeRecoveryEmail("User <user@example.com>"); err == nil {
		t.Fatal("display-name address should be rejected")
	}
}
