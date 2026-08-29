package notification

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/wavefnd/wave-platform/internal/storage"
)

func TestNotificationLifecycleAndAccountIsolation(t *testing.T) {
	database, err := storage.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = database.Close() }()
	service := NewService(database)
	service.now = func() time.Time { return time.Date(2026, time.August, 29, 5, 0, 0, 0, time.UTC) }
	created, err := service.Notify(Input{RecipientAccountID: "recipient", ActorAccountID: "actor", ActorName: "Member",
		Kind: "question.answer", Subject: "How does Wave compile?", URL: "/questions/question-1"})
	if err != nil {
		t.Fatal(err)
	}
	items, unread, err := service.List("recipient", 20)
	if err != nil || len(items) != 1 || unread != 1 || items[0].ID != created.ID {
		t.Fatalf("items=%#v unread=%d err=%v", items, unread, err)
	}
	if _, err := service.MarkRead("another-account", created.ID); !errors.Is(err, storage.ErrNotFound) {
		t.Fatalf("cross-account read error=%v", err)
	}
	if _, err := service.MarkRead("recipient", created.ID); err != nil {
		t.Fatal(err)
	}
	_, unread, err = service.List("recipient", 20)
	if err != nil || unread != 0 {
		t.Fatalf("unread=%d err=%v", unread, err)
	}
}

func TestNotificationSkipsSelfAndRetainsNewestItems(t *testing.T) {
	database, err := storage.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = database.Close() }()
	service := NewService(database)
	if item, err := service.Notify(Input{RecipientAccountID: "same", ActorAccountID: "same", Kind: "blog.comment",
		Subject: "Post", URL: "/blog/post"}); err != nil || item.ID != "" {
		t.Fatalf("self notification=%#v err=%v", item, err)
	}
	base := time.Date(2026, time.August, 29, 0, 0, 0, 0, time.UTC)
	for index := 0; index < MaxStoredPerAccount+3; index++ {
		current := index
		service.now = func() time.Time { return base.Add(time.Duration(current) * time.Second) }
		if _, err := service.Notify(Input{RecipientAccountID: "recipient", Kind: "rfc.status",
			Subject: fmt.Sprintf("RFC %d", index), Detail: "discussion", URL: fmt.Sprintf("/rfcs/%d", index)}); err != nil {
			t.Fatal(err)
		}
	}
	items, unread, err := service.List("recipient", 50)
	if err != nil || unread != MaxStoredPerAccount || len(items) != 50 || items[0].Subject != "RFC 202" {
		t.Fatalf("items=%d unread=%d first=%#v err=%v", len(items), unread, items[0], err)
	}
}
