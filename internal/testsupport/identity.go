package testsupport

import (
	"time"

	"github.com/pquerna/otp/totp"
	"github.com/wavefnd/wave-platform/internal/account"
	"github.com/wavefnd/wave-platform/internal/identity"
	"github.com/wavefnd/wave-platform/internal/session"
	"github.com/wavefnd/wave-platform/internal/storage"
)

const (
	TOTPSecret        = "JBSWY3DPEHPK3PXPJBSWY3DPEHPK3PXP"
	authEncryptionKey = "MDEyMzQ1Njc4OWFiY2RlZjAxMjM0NTY3ODlhYmNkZWY="
)

func NewIdentity(database *storage.Database) (*identity.Service, error) {
	return identity.NewServiceWithTOTP(database, "wave-lang.dev", true, time.Hour,
		authEncryptionKey, "Wave Test", "http://localhost")
}

func Register(service *identity.Service, displayName string) (account.Account, error) {
	return service.Register(identity.Registration{DisplayName: displayName, TOTPSecret: TOTPSecret,
		RecoveryEmail: "recovery@example.net"})
}

func Authenticate(service *identity.Service, identifier string) (account.Account, string, session.Session, error) {
	code, err := totp.GenerateCode(TOTPSecret, time.Now())
	if err != nil {
		return account.Account{}, "", session.Session{}, err
	}
	return service.AuthenticateTOTP(identifier, code, "test")
}
