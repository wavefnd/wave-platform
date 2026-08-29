package rfc

import (
	"errors"
	"testing"
	"time"

	"github.com/wavefnd/wave-platform/internal/account"
	"github.com/wavefnd/wave-platform/internal/audit"
	notificationdomain "github.com/wavefnd/wave-platform/internal/notification"
	"github.com/wavefnd/wave-platform/internal/permission"
	"github.com/wavefnd/wave-platform/internal/storage"
)

func TestRFCProposalDiscussionAndIndependentMaintainerRole(t *testing.T) {
	database, err := storage.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = database.Close() })
	now := time.Date(2026, 8, 21, 14, 0, 0, 0, time.UTC)
	accounts, permissions := account.NewRepository(database), permission.NewRepository(database)
	for _, id := range []string{"owner", "admin", "author", "maintainer", "reader"} {
		if err := accounts.Create(account.Account{ID: id, Username: id, DisplayName: id, Email: id + "@wave.test", Status: "active", CreatedAt: now, UpdatedAt: now}); err != nil {
			t.Fatal(err)
		}
	}
	_ = permissions.Assign(permission.Assignment{AccountID: "owner", RoleID: "platform-owner", Scope: "platform"})
	_ = permissions.Assign(permission.Assignment{AccountID: "admin", RoleID: "platform-admin", Scope: "platform"})
	_ = permissions.Assign(permission.Assignment{AccountID: "maintainer", RoleID: "rfc-maintainer", Scope: "rfc"})
	service := NewService(database)
	notifications := notificationdomain.NewService(database)
	service.SetNotificationService(notifications)
	service.now = func() time.Time { return now }

	created, err := service.Create("author", ProposalInput{Title: "WebAssembly target support", Content: "## Motivation\n\nDefine an official WebAssembly target and stable host ABI."})
	if err != nil || created.Number != 1 || created.Status != "draft" {
		t.Fatalf("created=%#v err=%v", created, err)
	}
	if _, err := service.Update("reader", created.Number, ProposalInput{Title: created.Title, Content: created.Content}); !errors.Is(err, ErrForbidden) {
		t.Fatalf("reader update error=%v", err)
	}
	if _, err := service.UpdateStatus("admin", created.Number, "discussion"); !errors.Is(err, ErrForbidden) {
		t.Fatalf("administrator status error=%v", err)
	}
	item, err := service.UpdateStatus("maintainer", created.Number, "discussion")
	if err != nil || item.Status != "discussion" {
		t.Fatalf("discussion=%#v err=%v", item, err)
	}
	if _, err := service.Update("author", created.Number, ProposalInput{Title: created.Title, Content: created.Content + "\nChanged."}); !errors.Is(err, ErrForbidden) {
		t.Fatalf("non-draft author update error=%v", err)
	}
	comment, err := service.AddComment("reader", created.Number, CommentInput{Body: "Please specify browser and WASI compatibility."})
	if err != nil || comment.AuthorName != "reader" {
		t.Fatalf("comment=%#v err=%v", comment, err)
	}
	detail, err := service.Repository().Proposal(created.Number)
	if err != nil || detail.CommentCount != 1 || len(detail.Comments) != 1 {
		t.Fatalf("detail=%#v err=%v", detail, err)
	}
	events, err := audit.NewRepository(database).Events(10)
	if err != nil || len(events) != 3 || events[0].Action != "rfc.comment" || events[1].Action != "rfc.status.discussion" || events[2].Action != "rfc.create" {
		t.Fatalf("events=%#v err=%v", events, err)
	}
	items, unread, err := notifications.List("author", 20)
	kinds := map[string]bool{}
	for _, item := range items {
		kinds[item.Kind] = true
	}
	if err != nil || unread != 2 || len(items) != 2 || !kinds["rfc.comment"] || !kinds["rfc.status"] {
		t.Fatalf("notifications=%#v unread=%d err=%v", items, unread, err)
	}
}
