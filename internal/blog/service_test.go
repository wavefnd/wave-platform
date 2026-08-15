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
	item, err := service.Save("owner", Input{Slug: "wave-020", Category: "release", Title: "Wave 0.2.0",
		Summary: "Release summary", Content: "## Release\n\nDetails.", Status: "draft"})
	if err != nil || item.PublishedAt != "" {
		t.Fatalf("draft=%#v err=%v", item, err)
	}
	if _, err := service.Repository().Post(item.Slug, false); !errors.Is(err, storage.ErrNotFound) {
		t.Fatalf("public draft error=%v", err)
	}
	service.now = func() time.Time { return now.Add(time.Hour) }
	item, err = service.Save("owner", Input{Slug: item.Slug, Category: "release", Title: item.Title,
		Summary: item.Summary, Content: item.Content, Status: "published"})
	if err != nil || item.PublishedAt == "" {
		t.Fatalf("published=%#v err=%v", item, err)
	}
	posts, err := service.Repository().Posts(false, 10)
	if err != nil || len(posts) != 1 || posts[0].AuthorName != "Wave Owner" || posts[0].Category != "release" {
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
	if _, err := service.Save("missing", Input{Slug: "Not Valid", Title: "Title", Summary: "Summary", Content: "Body", Status: "published"}); !errors.Is(err, ErrInvalidPost) {
		t.Fatalf("invalid slug error=%v", err)
	}
}

func TestRoadmapPostsUseEditableDisplayOrderInsteadOfTargetDate(t *testing.T) {
	database, err := storage.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = database.Close() })
	now := time.Date(2026, 8, 15, 10, 0, 0, 0, time.UTC)
	if err := account.NewRepository(database).Create(account.Account{ID: "owner", Username: "owner", DisplayName: "Wave Owner",
		Email: "owner@wave.test", Status: "active", CreatedAt: now, UpdatedAt: now}); err != nil {
		t.Fatal(err)
	}
	service := NewService(database)
	service.now = func() time.Time { return now }
	inputs := []Input{
		{Slug: "wave-next", Category: "roadmap", RoadmapStatus: "planned", RoadmapOrder: 2, TargetDate: "2026-09-01", Cadence: "Every 8 weeks", Title: "Wave next", Summary: "Next version", Content: "- [ ] Parser", Status: "published"},
		{Slug: "wave-later", Category: "roadmap", RoadmapStatus: "in-progress", RoadmapOrder: 1, TargetDate: "2026-12-01", Cadence: "Quarterly", Title: "Wave later", Summary: "Later version", Content: "- [x] Design", Status: "published"},
	}
	for _, input := range inputs {
		if _, err := service.Save("owner", input); err != nil {
			t.Fatal(err)
		}
	}
	items, err := service.Repository().PostsByCategory("roadmap", false, 0)
	if err != nil || len(items) != 2 || items[0].Slug != "wave-later" || items[1].Slug != "wave-next" {
		t.Fatalf("roadmap order=%#v err=%v", items, err)
	}
	inputs[0].RoadmapOrder = 0
	if _, err := service.Save("owner", inputs[0]); !errors.Is(err, ErrInvalidPost) {
		t.Fatalf("invalid roadmap order error=%v", err)
	}
}
