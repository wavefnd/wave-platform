package blog

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/wavefnd/wave-platform/internal/account"
	"github.com/wavefnd/wave-platform/internal/audit"
	"github.com/wavefnd/wave-platform/internal/identifier"
	"github.com/wavefnd/wave-platform/internal/storage"
)

var (
	ErrInvalidPost = errors.New("invalid blog post")
	slugPattern    = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
)

type Service struct {
	repository *Repository
	accounts   *account.Repository
	audit      *audit.Repository
	now        func() time.Time
}

func NewService(database *storage.Database) *Service {
	return &Service{repository: NewRepository(database), accounts: account.NewRepository(database),
		audit: audit.NewRepository(database), now: time.Now}
}

func (service *Service) Repository() *Repository { return service.repository }

func (service *Service) Save(actorID string, input Input) (Post, error) {
	input.Slug = strings.ToLower(strings.TrimSpace(input.Slug))
	input.Locale = strings.ToLower(strings.TrimSpace(input.Locale))
	input.Title = strings.TrimSpace(input.Title)
	input.Summary = strings.TrimSpace(input.Summary)
	input.Content = strings.TrimSpace(strings.ReplaceAll(input.Content, "\r\n", "\n"))
	input.Status = strings.ToLower(strings.TrimSpace(input.Status))
	if !slugPattern.MatchString(input.Slug) || len(input.Slug) > 120 {
		return Post{}, fmt.Errorf("%w: slug must contain lowercase letters, numbers, and single hyphens", ErrInvalidPost)
	}
	if input.Locale != "en" && input.Locale != "ko" {
		return Post{}, fmt.Errorf("%w: locale must be en or ko", ErrInvalidPost)
	}
	if len([]rune(input.Title)) < 1 || len([]rune(input.Title)) > 160 {
		return Post{}, fmt.Errorf("%w: title must contain between 1 and 160 characters", ErrInvalidPost)
	}
	if len([]rune(input.Summary)) < 1 || len([]rune(input.Summary)) > 500 {
		return Post{}, fmt.Errorf("%w: summary must contain between 1 and 500 characters", ErrInvalidPost)
	}
	if len([]rune(input.Content)) < 1 || len([]rune(input.Content)) > 200000 {
		return Post{}, fmt.Errorf("%w: content must contain between 1 and 200000 characters", ErrInvalidPost)
	}
	if input.Status != "draft" && input.Status != "published" {
		return Post{}, fmt.Errorf("%w: status must be draft or published", ErrInvalidPost)
	}
	author, err := service.accounts.Account(actorID)
	if err != nil {
		return Post{}, err
	}
	now := service.now().UTC()
	item, err := service.repository.Post(input.Slug, true)
	if errors.Is(err, storage.ErrNotFound) {
		item = Post{Slug: input.Slug, CreatedAt: now}
	} else if err != nil {
		return Post{}, err
	}
	item.Locale, item.Title, item.Summary, item.Content, item.Status = input.Locale, input.Title, input.Summary, input.Content, input.Status
	item.AuthorAccountID, item.AuthorName, item.UpdatedAt = author.ID, author.DisplayName, now
	if item.Status == "published" && item.PublishedAt == "" {
		item.PublishedAt = now.Format(timeLayout)
	}
	if err := service.repository.Upsert(item); err != nil {
		return Post{}, err
	}
	if err := service.appendAudit(actorID, item.Slug, "admin.blog.save"); err != nil {
		return Post{}, err
	}
	return item, nil
}

func (service *Service) appendAudit(actorID, slug, action string) error {
	id, err := identifier.New("audit")
	if err != nil {
		return err
	}
	return service.audit.Append(audit.Event{ID: id, ActorID: "account/" + actorID,
		ResourceID: "blog/" + slug, Action: action, Result: "success", OccurredAt: service.now().UTC()})
}
