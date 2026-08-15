package blog

import (
	"errors"
	"testing"
	"time"

	"github.com/wavefnd/wave-platform/internal/account"
	"github.com/wavefnd/wave-platform/internal/audit"
	"github.com/wavefnd/wave-platform/internal/storage"
)

func TestSavePublishesAndUpdatesBlogPost(t *testing.T) {
	database, err := storage.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = database.Close() })
	now := time.Date(2026, 8, 14, 10, 0, 0, 0, time.UTC)
	if err := account.NewRepository(database).Create(account.Account{ID: "owner", Username: "owner", DisplayName: "Wave Owner",
		Email: "owner@wave.test", Status: "active", CreatedAt: now, UpdatedAt: now}); err != nil {
		t.Fatal(err)
	}
	service := NewService(database)
	service.now = func() time.Time { return now }
	item, err := service.Save("owner", Input{Slug: "wave-020", Locale: "en", Title: "Wave 0.2.0",
		Summary: "Release summary", Content: "## Release\n\nDetails.", Status: "draft"})
	if err != nil || item.PublishedAt != "" {
		t.Fatalf("draft=%#v err=%v", item, err)
	}
	if _, err := service.Repository().Post(item.Slug, false); !errors.Is(err, storage.ErrNotFound) {
		t.Fatalf("public draft error=%v", err)
	}
	service.now = func() time.Time { return now.Add(time.Hour) }
	item, err = service.Save("owner", Input{Slug: item.Slug, Locale: "en", Title: item.Title,
		Summary: item.Summary, Content: item.Content, Status: "published"})
	if err != nil || item.PublishedAt == "" {
		t.Fatalf("published=%#v err=%v", item, err)
	}
	posts, err := service.Repository().Posts("en", false, 10)
	if err != nil || len(posts) != 1 || posts[0].AuthorName != "Wave Owner" {
		t.Fatalf("posts=%#v err=%v", posts, err)
	}
	events, err := audit.NewRepository(database).Events(10)
	if err != nil || len(events) != 2 || events[0].ResourceID != "blog/wave-020" {
		t.Fatalf("events=%#v err=%v", events, err)
	}
}

func TestSaveRejectsInvalidBlogPost(t *testing.T) {
	database, err := storage.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = database.Close() })
	service := NewService(database)
	if _, err := service.Save("missing", Input{Slug: "Not Valid", Locale: "en", Title: "Title", Summary: "Summary", Content: "Body", Status: "published"}); !errors.Is(err, ErrInvalidPost) {
		t.Fatalf("invalid slug error=%v", err)
	}
}
