package storage

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestDatabaseLifecycleAndDirectories(t *testing.T) {
	root := filepath.Join(t.TempDir(), "platform-data")
	database, err := Open(root)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	t.Cleanup(func() { _ = database.Close() })

	for _, name := range []string{"badger", "blobs", "mail", "git", "search", "temporary", "backups"} {
		info, err := os.Stat(filepath.Join(root, name))
		if err != nil {
			t.Fatalf("stat %s: %v", name, err)
		}
		if !info.IsDir() {
			t.Fatalf("%s is not a directory", name)
		}
	}
	for _, name := range []string{filepath.Join("git", "mirrors"), filepath.Join("git", "temporary")} {
		info, err := os.Stat(filepath.Join(root, name))
		if err != nil || !info.IsDir() {
			t.Fatalf("git storage directory %s: %v", name, err)
		}
	}

	key := Key("meta", "object", "platform")
	if err := database.Set(key, []byte("ready")); err != nil {
		t.Fatalf("Set() error = %v", err)
	}
	value, err := database.Get(key)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if string(value) != "ready" {
		t.Fatalf("value = %q", value)
	}
	if err := database.Health(); err != nil {
		t.Fatalf("Health() error = %v", err)
	}
	if err := database.Delete(key); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if _, err := database.Get(key); !errors.Is(err, ErrNotFound) {
		t.Fatalf("Get() after delete error = %v", err)
	}
}

func TestKeyNormalizesSegments(t *testing.T) {
	if got := string(Key("/document/", "object", "", "/document-id/")); got != "document/object/document-id" {
		t.Fatalf("Key() = %q", got)
	}
	if got := string(Prefix("document", "object")); got != "document/object/" {
		t.Fatalf("Prefix() = %q", got)
	}
}
