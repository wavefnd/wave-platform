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
		{Slug: "wave-next", Category: "roadmap", RoadmapStatus: "planned", RoadmapOrder: 2, TargetDate: "2026-09-01", Title: "Wave next", Content: "- [ ] Parser", Status: "published"},
		{Slug: "wave-later", Category: "roadmap", RoadmapStatus: "in-progress", RoadmapOrder: 1, TargetDate: "2026-12-01", Title: "Wave later", Content: "- [x] Design", Status: "published"},
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

func TestRoadmapDerivesVersionSlugAndContentSummary(t *testing.T) {
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
	content := "12345678901234567890123456789012345678901234567890 more"
	item, err := service.Save("owner", Input{Category: "roadmap", RoadmapStatus: "planned", RoadmapOrder: 1,
		TargetDate: "2026-09-01", Title: "v0.3.0-pre-beta", Content: content, Status: "published"})
	if err != nil {
		t.Fatal(err)
	}
	if item.Slug != "v0.3.0-pre-beta" {
		t.Fatalf("slug=%q", item.Slug)
	}
	if got := []rune(item.Summary); len(got) != 48 || string(got) != string([]rune(content)[:48]) {
		t.Fatalf("summary=%q runes=%d", item.Summary, len(got))
	}
}

func TestArticleCommentsArePublicModeratableAndRateLimited(t *testing.T) {
	database, err := storage.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = database.Close() })
	now := time.Date(2026, 8, 26, 3, 0, 0, 0, time.UTC)
	accounts := account.NewRepository(database)
	for _, item := range []account.Account{
		{ID: "owner", Username: "owner", DisplayName: "Wave Owner", Email: "owner@wave.test", Status: "active", CreatedAt: now, UpdatedAt: now},
		{ID: "member", Username: "member", DisplayName: "Wave Member", Email: "member@wave.test", Status: "active", CreatedAt: now, UpdatedAt: now},
	} {
		if err := accounts.Create(item); err != nil {
			t.Fatal(err)
		}
	}
	service := NewService(database)
	service.now = func() time.Time { return now }
	post, err := service.Save("owner", Input{Slug: "compiler-notes", Category: "article", Title: "Compiler notes",
		Summary: "An engineering update", Content: "## Parser", Status: "published", CommentPolicy: "open"})
	if err != nil {
		t.Fatal(err)
	}
	if post.CommentPolicy != "open" {
		t.Fatalf("comment policy=%q", post.CommentPolicy)
	}
	comment, err := service.AddComment("member", post.Slug, CommentInput{Body: "This makes the parser easier to follow."})
	if err != nil {
		t.Fatal(err)
	}
	if comment.AuthorName != "Wave Member" || comment.Status != "visible" {
		t.Fatalf("comment=%#v", comment)
	}
	if _, err := service.AddComment("member", post.Slug, CommentInput{Body: "A second comment."}); !errors.Is(err, ErrCommentRateLimited) {
		t.Fatalf("rate-limit error=%v", err)
	}
	if _, err := service.AddComment("owner", post.Slug, CommentInput{Body: "![image](https://example.test/image.webp)"}); !errors.Is(err, ErrInvalidComment) {
		t.Fatalf("image error=%v", err)
	}
	if _, err := service.AddComment("owner", post.Slug, CommentInput{Body: "좋은 글입니다."}); !errors.Is(err, ErrInvalidComment) {
		t.Fatalf("language error=%v", err)
	}
	hidden, err := service.SetCommentStatus("owner", post.Slug, comment.ID, "hidden")
	if err != nil || hidden.Status != "hidden" {
		t.Fatalf("hidden=%#v err=%v", hidden, err)
	}
	visible, err := service.Comments(post.Slug, false)
	if err != nil || len(visible) != 0 {
		t.Fatalf("visible=%#v err=%v", visible, err)
	}
	all, err := service.Comments(post.Slug, true)
	if err != nil || len(all) != 1 {
		t.Fatalf("all=%#v err=%v", all, err)
	}
}

func TestReleaseCommentsRemainDisabled(t *testing.T) {
	database, err := storage.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = database.Close() })
	now := time.Date(2026, 8, 26, 3, 0, 0, 0, time.UTC)
	if err := account.NewRepository(database).Create(account.Account{ID: "owner", Username: "owner", DisplayName: "Wave Owner",
		Email: "owner@wave.test", Status: "active", CreatedAt: now, UpdatedAt: now}); err != nil {
		t.Fatal(err)
	}
	service := NewService(database)
	service.now = func() time.Time { return now }
	post, err := service.Save("owner", Input{Slug: "v0.3.0", Category: "release", Title: "Wave v0.3.0",
		Summary: "Release", Content: "Details", Status: "published", CommentPolicy: "open"})
	if err != nil {
		t.Fatal(err)
	}
	if post.CommentPolicy != "disabled" {
		t.Fatalf("release policy=%q", post.CommentPolicy)
	}
	if _, err := service.Comments(post.Slug, false); !errors.Is(err, ErrCommentsClosed) {
		t.Fatalf("comments error=%v", err)
	}
}
