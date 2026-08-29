package notification

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/wavefnd/wave-platform/internal/identifier"
	"github.com/wavefnd/wave-platform/internal/storage"
)

var ErrInvalidNotification = errors.New("invalid notification")

var validKinds = map[string]bool{
	"blog.comment": true, "community.reply": true, "question.answer": true,
	"question.accepted": true, "rfc.comment": true, "rfc.status": true,
}

type Service struct {
	repository *Repository
	now        func() time.Time
}

func NewService(database *storage.Database) *Service {
	return &Service{repository: NewRepository(database), now: time.Now}
}

func (service *Service) Repository() *Repository { return service.repository }

func (service *Service) Notify(input Input) (Item, error) {
	input.RecipientAccountID = strings.TrimSpace(input.RecipientAccountID)
	input.ActorAccountID = strings.TrimSpace(input.ActorAccountID)
	input.ActorName = strings.TrimSpace(input.ActorName)
	input.Kind = strings.ToLower(strings.TrimSpace(input.Kind))
	input.Subject = strings.TrimSpace(input.Subject)
	input.Detail = strings.TrimSpace(input.Detail)
	input.URL = strings.TrimSpace(input.URL)
	if input.RecipientAccountID == "" || !validKinds[input.Kind] || input.Subject == "" || len([]rune(input.Subject)) > 180 ||
		!strings.HasPrefix(input.URL, "/") || strings.HasPrefix(input.URL, "//") || strings.ContainsAny(input.URL, "\r\n") {
		return Item{}, ErrInvalidNotification
	}
	if len([]rune(input.ActorName)) > 100 || len([]rune(input.Detail)) > 120 {
		return Item{}, ErrInvalidNotification
	}
	if input.ActorAccountID != "" && input.ActorAccountID == input.RecipientAccountID {
		return Item{}, nil
	}
	id, err := identifier.New("notification")
	if err != nil {
		return Item{}, err
	}
	item := Item{ID: id, RecipientAccountID: input.RecipientAccountID, ActorAccountID: input.ActorAccountID,
		ActorName: input.ActorName, Kind: input.Kind, Subject: input.Subject, Detail: input.Detail,
		URL: input.URL, CreatedAt: service.now().UTC()}
	if err := service.repository.Upsert(item); err != nil {
		return Item{}, err
	}
	if err := service.prune(input.RecipientAccountID); err != nil {
		return Item{}, fmt.Errorf("prune notifications: %w", err)
	}
	return item, nil
}

func (service *Service) List(accountID string, limit int) ([]Item, int, error) {
	if limit < 1 || limit > 50 {
		limit = 20
	}
	all, err := service.repository.List(strings.TrimSpace(accountID), 0)
	if err != nil {
		return nil, 0, err
	}
	unread := 0
	for _, item := range all {
		if item.ReadAt == nil {
			unread++
		}
	}
	if len(all) > limit {
		all = all[:limit]
	}
	return all, unread, nil
}

func (service *Service) MarkRead(accountID, notificationID string) (Item, error) {
	item, err := service.repository.Item(strings.TrimSpace(accountID), strings.TrimSpace(notificationID))
	if err != nil {
		return Item{}, err
	}
	if item.ReadAt == nil {
		now := service.now().UTC()
		item.ReadAt = &now
		if err := service.repository.Upsert(item); err != nil {
			return Item{}, err
		}
	}
	return item, nil
}

func (service *Service) MarkAllRead(accountID string) error {
	items, err := service.repository.List(strings.TrimSpace(accountID), 0)
	if err != nil {
		return err
	}
	now := service.now().UTC()
	for _, item := range items {
		if item.ReadAt != nil {
			continue
		}
		item.ReadAt = &now
		if err := service.repository.Upsert(item); err != nil {
			return err
		}
	}
	return nil
}

func (service *Service) prune(accountID string) error {
	items, err := service.repository.List(accountID, 0)
	if err != nil {
		return err
	}
	for _, item := range items[minimum(len(items), MaxStoredPerAccount):] {
		if err := service.repository.Delete(accountID, item.ID); err != nil {
			return err
		}
	}
	return nil
}

func minimum(left, right int) int {
	if left < right {
		return left
	}
	return right
}
