package document

import (
	"strings"
	"testing"

	"github.com/wavefnd/wave-platform/internal/storage"
)

func TestSeedOfficialPublishesEnglishAndKoreanReference(t *testing.T) {
	database, err := storage.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = database.Close() })

	count, err := SeedOfficial(database)
	if err != nil || count != 48 {
		t.Fatalf("seed count=%d err=%v", count, err)
	}
	repository := NewRepository(database)
	for _, locale := range []string{"en", "ko"} {
		items, err := repository.Summaries(locale)
		if err != nil || len(items) != 24 {
			t.Fatalf("%s summaries=%d err=%v", locale, len(items), err)
		}
		view, err := repository.Published(locale, "language/explicit-memory-type-model")
		if err != nil {
			t.Fatal(err)
		}
		if view.Title != "Wave Explicit Memory Type Model" || view.Version != "0.2.0-pre-beta" || view.SourceRevision != "bd5549b" {
			t.Fatalf("unexpected memory document: %#v", view)
		}
		if !strings.Contains(view.Markdown, "ptr<T>") || !strings.Contains(view.Markdown, "```wave") {
			t.Fatal("published revision did not preserve the Markdown authoring source")
		}
	}
	install, err := repository.Published("en", "getting-started/install")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(install.Markdown, "curl -fsSL https://wave-lang.dev/install.sh | bash -s -- latest") {
		t.Fatal("official installer command is missing")
	}
	count, err = SeedOfficial(database)
	if err != nil || count != 0 {
		t.Fatalf("second seed count=%d err=%v", count, err)
	}
}

func TestOfficialInstallDocumentsIncludeWindowsInstaller(t *testing.T) {
	database, err := storage.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = database.Close() })
	if _, err := SeedOfficial(database); err != nil {
		t.Fatal(err)
	}
	repository := NewRepository(database)
	for _, locale := range []string{"en", "ko"} {
		install, err := repository.Published(locale, "getting-started/install")
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(install.Markdown, "https://wave-lang.dev/install.ps1") || !strings.Contains(install.Markdown, "-Latest") {
			t.Fatalf("%s official Windows installer command is missing", locale)
		}
	}
}
