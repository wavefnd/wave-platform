package platformstats

import (
	"context"
	"testing"
	"time"

	maildomain "github.com/wavefnd/wave-platform/internal/mail"
	"github.com/wavefnd/wave-platform/internal/storage"
)

func TestSnapshotCountsStoredPlatformActivity(t *testing.T) {
	database, err := storage.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()

	if err := database.Set(storage.Key("account", "object", "one"), []byte("<account/>")); err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 8, 5, 12, 0, 0, 0, time.Local)
	mail := maildomain.NewRepository(database)
	for _, message := range []maildomain.Message{
		{ID: "today", MessageID: "today@wave.local", CreatedAt: now.Add(-time.Hour)},
		{ID: "yesterday", MessageID: "yesterday@wave.local", CreatedAt: now.Add(-24 * time.Hour)},
	} {
		if err := mail.UpsertMessage(message, []byte("Subject: test\r\n\r\nbody")); err != nil {
			t.Fatal(err)
		}
	}

	service := NewService(database, nil)
	service.now = func() time.Time { return now }
	snapshot, err := service.Snapshot(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if snapshot.Accounts != 1 || snapshot.MessagesToday != 1 || snapshot.GitMirrors != 0 {
		t.Fatalf("snapshot = %#v", snapshot)
	}
}
