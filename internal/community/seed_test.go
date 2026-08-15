package community

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	blogdomain "github.com/wavefnd/wave-platform/internal/blog"
	"github.com/wavefnd/wave-platform/internal/storage"
)

func TestSeedReleaseBlogPostsImportsOnlyVersionReleases(t *testing.T) {
	database, err := storage.Open(filepath.Join(t.TempDir(), "data"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = database.Close() })

	count, err := SeedReleaseBlogPosts(database)
	if err != nil {
		t.Fatal(err)
	}
	if count != 15 {
		t.Fatalf("count = %d, want 15", count)
	}

	repository := blogdomain.NewRepository(database)
	items, err := repository.PostsByCategory("release", false, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 15 {
		t.Fatalf("releases = %d, want 15", len(items))
	}
	latestOnly, err := repository.PostsByCategory("release", false, 1)
	if err != nil || len(latestOnly) != 1 || latestOnly[0].Slug != items[0].Slug {
		t.Fatalf("latest release = %#v, err = %v", latestOnly, err)
	}
	if items[0].Slug != "wave-v0-2-0-pre-beta" {
		t.Fatalf("latest slug = %q", items[0].Slug)
	}
	latest, err := repository.Post(items[0].Slug, false)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(latest.Content, "## What's changed") {
		t.Fatal("release MIME body was not resolved")
	}
	stored, err := database.Get(storage.Key("blog", "post", latest.Slug))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(stored), "## What") {
		t.Fatal("blog post must store the release body")
	}
	if latest.Category != "release" || latest.Status != "published" || latest.AuthorName != "Wave Foundation" {
		t.Fatalf("release blog metadata = %#v", latest)
	}
	if _, err := database.Get(storage.Key("community", "announcement", "object", latest.Slug)); !errors.Is(err, storage.ErrNotFound) {
		t.Fatalf("release leaked into community namespace: %v", err)
	}

	count, err = SeedReleaseBlogPosts(database)
	if err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("second seed count = %d, want 0", count)
	}
}

func TestParseReleasePostRejectsPatch(t *testing.T) {
	if _, err := parseReleasePost("2026-01-01-patch-1.md", ""); err == nil {
		t.Fatal("expected patch rejection")
	}
}

func TestSeedSpacesIncludesPersonalWritingCategories(t *testing.T) {
	database, err := storage.Open(filepath.Join(t.TempDir(), "data"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = database.Close() })
	if err := SeedSpaces(database); err != nil {
		t.Fatal(err)
	}

	repository := NewRepository(database)
	for _, id := range []string{"founder-notes", "development-log"} {
		space, err := repository.Space(id)
		if err != nil {
			t.Fatalf("personal space %q: %v", id, err)
		}
		if space.PostingPolicy != "owner" || space.Visibility != "public" {
			t.Fatalf("personal space %q = %#v", id, space)
		}
	}
	for _, id := range []string{"general", "development", "operating-systems", "web", "compiler", "audio", "gui", "showcase", "help"} {
		space, err := repository.Space(id)
		if err != nil {
			t.Fatalf("member space %q: %v", id, err)
		}
		if space.PostingPolicy != "members" || space.Visibility != "public" {
			t.Fatalf("member space %q = %#v", id, space)
		}
	}
}
