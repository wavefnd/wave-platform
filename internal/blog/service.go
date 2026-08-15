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
	slugPattern    = regexp.MustCompile(`^[a-z0-9]+(?:[.-][a-z0-9]+)*$`)
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
	input.Category = NormalizeCategory(input.Category)
	input.RoadmapStatus = strings.ToLower(strings.TrimSpace(input.RoadmapStatus))
	input.TargetDate = strings.TrimSpace(input.TargetDate)
	input.Title = strings.TrimSpace(input.Title)
	input.Content = strings.TrimSpace(strings.ReplaceAll(input.Content, "\r\n", "\n"))
	input.Status = strings.ToLower(strings.TrimSpace(input.Status))
	input.Slug = strings.ToLower(strings.TrimSpace(input.Slug))
	if input.Category == "roadmap" && input.Slug == "" {
		input.Slug = SlugFromTitle(input.Title)
	}
	if !slugPattern.MatchString(input.Slug) || len(input.Slug) > 120 {
		return Post{}, fmt.Errorf("%w: slug must contain lowercase letters, numbers, dots, and single hyphens", ErrInvalidPost)
	}
	if input.Category != "article" && input.Category != "release" && input.Category != "roadmap" {
		return Post{}, fmt.Errorf("%w: category must be article, release, or roadmap", ErrInvalidPost)
	}
	if input.Category == "roadmap" {
		if input.RoadmapStatus != "planned" && input.RoadmapStatus != "in-progress" && input.RoadmapStatus != "released" {
			return Post{}, fmt.Errorf("%w: roadmap status must be planned, in-progress, or released", ErrInvalidPost)
		}
		if _, err := time.Parse("2006-01-02", input.TargetDate); err != nil {
			return Post{}, fmt.Errorf("%w: roadmap target release date must use YYYY-MM-DD", ErrInvalidPost)
		}
		if input.RoadmapOrder < 1 || input.RoadmapOrder > 1000000 {
			return Post{}, fmt.Errorf("%w: roadmap order must be between 1 and 1000000", ErrInvalidPost)
		}
	} else {
		input.RoadmapStatus, input.TargetDate = "", ""
		input.RoadmapOrder = 0
	}
	if len([]rune(input.Title)) < 1 || len([]rune(input.Title)) > 160 {
		return Post{}, fmt.Errorf("%w: title must contain between 1 and 160 characters", ErrInvalidPost)
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
	item.Category, item.Title, item.Content, item.Status = input.Category, input.Title, input.Content, input.Status
	item.Summary = SummaryFromContent(input.Content)
	item.RoadmapStatus, item.TargetDate = input.RoadmapStatus, input.TargetDate
	item.RoadmapOrder = input.RoadmapOrder
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

func SlugFromTitle(title string) string {
	var slug strings.Builder
	separator := false
	for _, character := range strings.ToLower(strings.TrimSpace(title)) {
		if character >= 'a' && character <= 'z' || character >= '0' && character <= '9' || character == '.' {
			if separator && slug.Len() > 0 {
				slug.WriteByte('-')
			}
			separator = false
			slug.WriteRune(character)
			continue
		}
		separator = true
	}
	return strings.Trim(slug.String(), ".-")
}

func SummaryFromContent(content string) string {
	characters := []rune(strings.Join(strings.Fields(content), " "))
	if len(characters) > 48 {
		characters = characters[:48]
	}
	return string(characters)
}

func NormalizeCategory(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" {
		return "article"
	}
	return value
}

func (service *Service) appendAudit(actorID, slug, action string) error {
	id, err := identifier.New("audit")
	if err != nil {
		return err
	}
	return service.audit.Append(audit.Event{ID: id, ActorID: "account/" + actorID,
		ResourceID: "blog/" + slug, Action: action, Result: "success", OccurredAt: service.now().UTC()})
}
