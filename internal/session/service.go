package session

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/wavefnd/wave-platform/internal/storage"
)

type Service struct {
	repository *Repository
	duration   time.Duration
	now        func() time.Time
}

func NewService(repository *Repository, duration time.Duration) *Service {
	if duration <= 0 {
		duration = 30 * 24 * time.Hour
	}
	return &Service{repository: repository, duration: duration, now: time.Now}
}

func (service *Service) Create(accountID, userAgent string) (string, Session, error) {
	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		return "", Session{}, fmt.Errorf("generate session token: %w", err)
	}
	token := base64.RawURLEncoding.EncodeToString(secret)
	now := service.now().UTC()
	userAgentRunes := []rune(userAgent)
	if len(userAgentRunes) > 256 {
		userAgent = string(userAgentRunes[:256])
	}
	item := Session{ID: TokenID(token), AccountID: accountID, CreatedAt: now, LastSeenAt: now,
		ExpiresAt: now.Add(service.duration), UserAgent: userAgent}
	if err := service.repository.Put(item); err != nil {
		return "", Session{}, err
	}
	return token, item, nil
}

func (service *Service) Verify(token string) (Session, error) {
	if len(token) < 32 || len(token) > 256 {
		return Session{}, storage.ErrNotFound
	}
	item, err := service.repository.Session(TokenID(token))
	if err != nil {
		return Session{}, err
	}
	if !item.ExpiresAt.After(service.now().UTC()) {
		_ = service.repository.Delete(item)
		return Session{}, storage.ErrNotFound
	}
	return item, nil
}

func (service *Service) Revoke(token string) error {
	item, err := service.Verify(token)
	if errors.Is(err, storage.ErrNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	return service.repository.Delete(item)
}

func (service *Service) RevokeAccount(accountID string) error {
	return service.repository.DeleteByAccount(accountID)
}

func TokenID(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}
