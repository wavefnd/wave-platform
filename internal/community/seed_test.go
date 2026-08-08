package community

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	releasedomain "github.com/wavefnd/wave-platform/internal/release"
	"github.com/wavefnd/wave-platform/internal/storage"
)

func TestSeedLanguageReleasesImportsOnlyVersionReleases(t *testing.T) {
	database, err := storage.Open(filepath.Join(t.TempDir(), "data"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = database.Close() })

	count, err := SeedLanguageReleases(database)
	if err != nil {
		t.Fatal(err)
	}
	if count != 15 {
		t.Fatalf("count = %d, want 15", count)
	}

	repository := releasedomain.NewRepository(database)
	items, err := repository.Releases(0)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 15 {
		t.Fatalf("releases = %d, want 15", len(items))
	}
	latestOnly, err := repository.Releases(1)
	if err != nil || len(latestOnly) != 1 || latestOnly[0].Slug != items[0].Slug {
		t.Fatalf("latest release = %#v, err = %v", latestOnly, err)
	}
	if items[0].Slug != "wave-v0-2-0-pre-beta" {
		t.Fatalf("latest slug = %q", items[0].Slug)
	}
	latest, err := repository.Release(items[0].Slug)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(latest.Content, "## What's changed") {
		t.Fatal("release MIME body was not resolved")
	}
	stored, err := database.Get(storage.Key("release", "object", latest.Slug))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(stored), "What's changed") {
		t.Fatal("release object must not duplicate the mail body")
	}
	if _, err := database.Get(storage.Key("community", "announcement", "object", latest.Slug)); !errors.Is(err, storage.ErrNotFound) {
		t.Fatalf("release leaked into community namespace: %v", err)
	}

	count, err = SeedLanguageReleases(database)
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
}
