package blog

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/wavefnd/wave-platform/internal/account"
	"github.com/wavefnd/wave-platform/internal/audit"
	"github.com/wavefnd/wave-platform/internal/identifier"
	notificationdomain "github.com/wavefnd/wave-platform/internal/notification"
	"github.com/wavefnd/wave-platform/internal/storage"
	webhookdomain "github.com/wavefnd/wave-platform/internal/webhook"
)

var (
	ErrInvalidPost        = errors.New("invalid blog post")
	ErrInvalidComment     = errors.New("invalid blog comment")
	ErrCommentsClosed     = errors.New("blog comments are not open")
	ErrCommentRateLimited = errors.New("wait before posting another blog comment")
	slugPattern           = regexp.MustCompile(`^[a-z0-9]+(?:[.-][a-z0-9]+)*$`)
	markdownImagePattern  = regexp.MustCompile(`(?i)<\s*img\b|!\[[^\]]*\]\s*(?:\([^)]*\)|\[[^\]]*\])`)
)

type Service struct {
	repository    *Repository
	accounts      *account.Repository
	audit         *audit.Repository
	webhooks      *webhookdomain.Service
	notifications *notificationdomain.Service
	now           func() time.Time
}

func NewService(database *storage.Database) *Service {
	return &Service{repository: NewRepository(database), accounts: account.NewRepository(database),
		audit: audit.NewRepository(database), now: time.Now}
}

func (service *Service) Repository() *Repository { return service.repository }

func (service *Service) SetWebhookService(webhooks *webhookdomain.Service) {
	service.webhooks = webhooks
}

func (service *Service) SetNotificationService(notifications *notificationdomain.Service) {
	service.notifications = notifications
}

func (service *Service) Save(actorID string, input Input) (Post, error) {
	input.Category = NormalizeCategory(input.Category)
	input.RoadmapStatus = strings.ToLower(strings.TrimSpace(input.RoadmapStatus))
	input.TargetDate = strings.TrimSpace(input.TargetDate)
	input.Title = strings.TrimSpace(input.Title)
	input.Summary = strings.TrimSpace(input.Summary)
	input.Content = strings.TrimSpace(strings.ReplaceAll(input.Content, "\r\n", "\n"))
	input.Status = strings.ToLower(strings.TrimSpace(input.Status))
	input.CommentPolicy = strings.ToLower(strings.TrimSpace(input.CommentPolicy))
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
	if input.Category != "roadmap" && (len([]rune(input.Summary)) < 1 || len([]rune(input.Summary)) > 500) {
		return Post{}, fmt.Errorf("%w: summary must contain between 1 and 500 characters", ErrInvalidPost)
	}
	if len([]rune(input.Content)) < 1 || len([]rune(input.Content)) > 200000 {
		return Post{}, fmt.Errorf("%w: content must contain between 1 and 200000 characters", ErrInvalidPost)
	}
	if input.Status != "draft" && input.Status != "published" {
		return Post{}, fmt.Errorf("%w: status must be draft or published", ErrInvalidPost)
	}
	if input.CommentPolicy == "" {
		input.CommentPolicy = NormalizeCommentPolicy(input.Category, "")
	}
	if input.Category != "article" {
		input.CommentPolicy = "disabled"
	} else if input.CommentPolicy != "open" && input.CommentPolicy != "locked" && input.CommentPolicy != "disabled" {
		return Post{}, fmt.Errorf("%w: comment policy must be open, locked, or disabled", ErrInvalidPost)
	}
	author, err := service.accounts.Account(actorID)
	if err != nil {
		return Post{}, err
	}
	now := service.now().UTC()
	item, err := service.repository.Post(input.Slug, true)
	previouslyPublished := false
	if errors.Is(err, storage.ErrNotFound) {
		item = Post{Slug: input.Slug, CreatedAt: now}
	} else if err != nil {
		return Post{}, err
	} else {
		previouslyPublished = item.PublishedAt != ""
	}
	item.Category, item.Title, item.Content, item.Status = input.Category, input.Title, input.Content, input.Status
	item.CommentPolicy = input.CommentPolicy
	item.Summary = input.Summary
	if item.Category == "roadmap" {
		item.Summary = SummaryFromContent(input.Content)
	}
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
	if service.webhooks != nil && item.Status == "published" && !previouslyPublished {
		eventType, path := webhookdomain.EventBlogPublished, "/blog/"+item.Slug
		if item.Category == "release" {
			eventType, path = webhookdomain.EventReleasePublished, "/releases/"+item.Slug
		}
		_ = service.webhooks.Publish(webhookdomain.Event{Type: eventType, Title: item.Title, Summary: item.Summary,
			AuthorName: item.AuthorName, ResourceID: "blog/" + item.Slug, URL: path, OccurredAt: now})
	}
	return item, nil
}

func NormalizeCommentPolicy(category, value string) string {
	if NormalizeCategory(category) != "article" {
		return "disabled"
	}
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "locked" || value == "disabled" {
		return value
	}
	return "open"
}

func (service *Service) Comments(slug string, includeHidden bool) ([]Comment, error) {
	post, err := service.repository.Post(slug, false)
	if err != nil {
		return nil, err
	}
	if NormalizeCategory(post.Category) != "article" || NormalizeCommentPolicy(post.Category, post.CommentPolicy) == "disabled" {
		return nil, ErrCommentsClosed
	}
	return service.repository.Comments(post.Slug, includeHidden)
}

func (service *Service) EditorComments(slug string) ([]Comment, error) {
	post, err := service.repository.Post(slug, true)
	if err != nil {
		return nil, err
	}
	if NormalizeCategory(post.Category) != "article" {
		return nil, ErrCommentsClosed
	}
	return service.repository.Comments(post.Slug, true)
}

func (service *Service) AddComment(actorID, slug string, input CommentInput) (Comment, error) {
	post, err := service.repository.Post(slug, false)
	if err != nil {
		return Comment{}, err
	}
	if NormalizeCommentPolicy(post.Category, post.CommentPolicy) != "open" {
		return Comment{}, ErrCommentsClosed
	}
	body := strings.TrimSpace(strings.ReplaceAll(input.Body, "\r\n", "\n"))
	if len([]rune(body)) < 1 || len([]rune(body)) > 10000 {
		return Comment{}, fmt.Errorf("%w: body must contain between 1 and 10000 characters", ErrInvalidComment)
	}
	if markdownImagePattern.MatchString(body) {
		return Comment{}, fmt.Errorf("%w: images are not supported in blog comments", ErrInvalidComment)
	}
	if !englishProse(withoutMarkdownCode(body)) {
		return Comment{}, fmt.Errorf("%w: comments must be written in English", ErrInvalidComment)
	}
	author, err := service.accounts.Account(actorID)
	if err != nil {
		return Comment{}, err
	}
	existing, err := service.repository.Comments(post.Slug, true)
	if err != nil {
		return Comment{}, err
	}
	now := service.now().UTC()
	for index := len(existing) - 1; index >= 0; index-- {
		if existing[index].AuthorAccountID == actorID && now.Sub(existing[index].CreatedAt) < 30*time.Second {
			return Comment{}, ErrCommentRateLimited
		}
	}
	id, err := identifier.New("blog-comment")
	if err != nil {
		return Comment{}, err
	}
	item := Comment{ID: id, PostSlug: post.Slug, AuthorAccountID: author.ID, AuthorName: author.DisplayName,
		Body: body, Status: "visible", CreatedAt: now, UpdatedAt: now}
	if err := service.repository.AddComment(item); err != nil {
		return Comment{}, err
	}
	if err := service.appendAudit(actorID, post.Slug+"/comments/"+id, "blog.comment.create"); err != nil {
		return Comment{}, err
	}
	if service.notifications != nil {
		_, _ = service.notifications.Notify(notificationdomain.Input{RecipientAccountID: post.AuthorAccountID,
			ActorAccountID: author.ID, ActorName: author.DisplayName, Kind: "blog.comment", Subject: post.Title,
			URL: "/blog/" + post.Slug})
	}
	return item, nil
}

func (service *Service) SetCommentStatus(actorID, slug, commentID, status string) (Comment, error) {
	status = strings.ToLower(strings.TrimSpace(status))
	if status != "visible" && status != "hidden" {
		return Comment{}, fmt.Errorf("%w: status must be visible or hidden", ErrInvalidComment)
	}
	item, err := service.repository.Comment(slug, commentID)
	if err != nil {
		return Comment{}, err
	}
	item.Status, item.UpdatedAt = status, service.now().UTC()
	if err := service.repository.AddComment(item); err != nil {
		return Comment{}, err
	}
	if err := service.appendAudit(actorID, strings.ToLower(strings.TrimSpace(slug))+"/comments/"+commentID, "admin.blog.comment."+status); err != nil {
		return Comment{}, err
	}
	return item, nil
}

func englishProse(value string) bool {
	for _, character := range value {
		if unicode.IsLetter(character) && !unicode.In(character, unicode.Latin) {
			return false
		}
	}
	return true
}

func withoutMarkdownCode(value string) string {
	var result strings.Builder
	inFence := false
	for _, line := range strings.Split(value, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "```") {
			inFence = !inFence
			continue
		}
		if inFence {
			continue
		}
		inInlineCode := false
		for _, character := range line {
			if character == '`' {
				inInlineCode = !inInlineCode
				continue
			}
			if !inInlineCode {
				result.WriteRune(character)
			}
		}
		result.WriteByte('\n')
	}
	return result.String()
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
