package patcharchive

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/wavefnd/wave-platform/internal/account"
	"github.com/wavefnd/wave-platform/internal/audit"
	"github.com/wavefnd/wave-platform/internal/identifier"
	maildomain "github.com/wavefnd/wave-platform/internal/mail"
	"github.com/wavefnd/wave-platform/internal/mailbox"
	"github.com/wavefnd/wave-platform/internal/permission"
	"github.com/wavefnd/wave-platform/internal/storage"
)

func TestArchiveExposesOnlyGitPatchMessages(t *testing.T) {
	database, err := storage.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	mailboxes, messages := mailbox.NewRepository(database), maildomain.NewRepository(database)
	box := mailbox.Mailbox{ID: "patch-box", AccountID: "service/patchs", Address: "patchs@wave-lang.dev", CreatedAt: time.Now().UTC()}
	if err := mailboxes.UpsertMailbox(box); err != nil {
		t.Fatal(err)
	}

	put := func(subject, body string) string {
		id, _ := identifier.New("message")
		message := maildomain.Message{ID: id, MessageID: "<" + id + "@example.net>", ThreadID: id, From: "Contributor <dev@example.net>",
			To: []string{box.Address}, Subject: subject, ReceivedAt: time.Now().UTC(), CreatedAt: time.Now().UTC()}
		raw := "From: " + message.From + "\r\nTo: " + box.Address + "\r\nSubject: " + subject + "\r\nContent-Type: text/plain; charset=utf-8\r\n\r\n" + body
		if err := messages.UpsertMessage(message, []byte(raw)); err != nil {
			t.Fatal(err)
		}
		entryID, _ := identifier.New("mailbox-entry")
		if err := mailboxes.AddEntry(mailbox.Entry{ID: entryID, MailboxID: box.ID, MessageID: id, Folder: "Inbox", CreatedAt: message.CreatedAt}); err != nil {
			t.Fatal(err)
		}
		return id
	}
	patchID := put("[PATCH v2 1/2] parser: keep source ranges", "Explain the parser change.\n\nSigned-off-by: Dev <dev@example.net>\n---\n file | 1 +\n\ndiff --git a/src/parser.go b/src/parser.go\n--- a/src/parser.go\n+++ b/src/parser.go\n@@ -1 +1 @@\n-old\n+new\n")
	put("[PATCH] Please read this", "This message only looks like a patch.")
	put("General discussion", "diff --git a/a b/a\n")

	service := NewService(database, "service/patchs", box.Address)
	items, err := service.List("parser", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0].ID != patchID || items[0].Version != 2 || items[0].Part != 1 || items[0].Total != 2 || len(items[0].Files) != 1 {
		t.Fatalf("items = %#v", items)
	}
	item, err := service.Get(patchID)
	if err != nil || !strings.Contains(item.Body, "+new") || item.AuthorEmail != "dev@example.net" {
		t.Fatalf("item=%#v err=%v", item, err)
	}
}

func TestMaintainerReviewsAndDownloadsThreadedPatchSeriesAsMbox(t *testing.T) {
	database, err := storage.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	now := time.Date(2026, 8, 21, 12, 0, 0, 0, time.UTC)
	mailboxes, messages := mailbox.NewRepository(database), maildomain.NewRepository(database)
	box := mailbox.Mailbox{ID: "patch-box", AccountID: "service/patchs", Address: "patchs@wave-lang.dev", CreatedAt: now}
	if err := mailboxes.UpsertMailbox(box); err != nil {
		t.Fatal(err)
	}
	put := func(messageID, subject, parent, body string, received time.Time) string {
		id, _ := identifier.New("message")
		message := maildomain.Message{ID: id, MessageID: messageID, ThreadID: id, From: "Dev <dev@example.net>",
			To: []string{box.Address}, Subject: subject, ReceivedAt: received, CreatedAt: received}
		raw := "From: " + message.From + "\r\nTo: " + box.Address + "\r\nSubject: " + subject + "\r\nMessage-ID: " + messageID + "\r\n"
		if parent != "" {
			raw += "In-Reply-To: " + parent + "\r\n"
		}
		raw += "Content-Type: text/plain; charset=utf-8\r\n\r\n" + body
		if err := messages.UpsertMessage(message, []byte(raw)); err != nil {
			t.Fatal(err)
		}
		entryID, _ := identifier.New("mailbox-entry")
		if err := mailboxes.AddEntry(mailbox.Entry{ID: entryID, MailboxID: box.ID, MessageID: id, Folder: "Inbox", CreatedAt: received}); err != nil {
			t.Fatal(err)
		}
		return id
	}
	coverHeaderID := "<series-cover@example.net>"
	put(coverHeaderID, "[PATCH v3 0/2] compiler cleanup", "", "Two compiler cleanup patches.", now)
	patchOneID := put("<series-1@example.net>", "[PATCH v3 1/2] parser: simplify ranges", coverHeaderID,
		"First patch.\n\ndiff --git a/parser.go b/parser.go\n--- a/parser.go\n+++ b/parser.go\n", now.Add(time.Minute))
	put("<series-2@example.net>", "[PATCH v3 2/2] codegen: simplify ranges", coverHeaderID,
		"Second patch.\n\ndiff --git a/codegen.go b/codegen.go\n--- a/codegen.go\n+++ b/codegen.go\n", now.Add(2*time.Minute))

	accounts := account.NewRepository(database)
	permissions := permission.NewRepository(database)
	for _, id := range []string{"owner", "admin", "maintainer"} {
		if err := accounts.Create(account.Account{ID: id, Username: id, DisplayName: strings.ToUpper(id), Email: id + "@wave-lang.dev", Status: "active", CreatedAt: now, UpdatedAt: now}); err != nil {
			t.Fatal(err)
		}
	}
	_ = permissions.Assign(permission.Assignment{AccountID: "owner", RoleID: "platform-owner", Scope: "platform"})
	_ = permissions.Assign(permission.Assignment{AccountID: "admin", RoleID: "platform-admin", Scope: "platform"})
	_ = permissions.Assign(permission.Assignment{AccountID: "maintainer", RoleID: "source-maintainer", Scope: "source"})
	service := NewService(database, "service/patchs", box.Address)
	service.now = func() time.Time { return now.Add(time.Hour) }
	if _, err := service.UpdateReview("admin", patchOneID, ReviewInput{Status: "accepted", TargetRepository: "wavefnd/Wave"}); !errors.Is(err, ErrForbidden) {
		t.Fatalf("administrator review error = %v", err)
	}
	if _, err := service.UpdateReview("maintainer", patchOneID, ReviewInput{Status: "accepted", TargetRepository: "invalid"}); !errors.Is(err, ErrInvalidReview) {
		t.Fatalf("invalid review error = %v", err)
	}
	item, err := service.UpdateReview("maintainer", patchOneID, ReviewInput{Status: "accepted", TargetRepository: "wavefnd/Wave"})
	if err != nil || item.ReviewStatus != "accepted" || item.TargetRepository != "wavefnd/Wave" || item.AssigneeName != "MAINTAINER" || item.SeriesCount != 3 || len(item.SHA256) != 64 {
		t.Fatalf("reviewed item=%#v err=%v", item, err)
	}
	content, filename, err := service.DownloadMbox("maintainer", patchOneID, true)
	if err != nil || !strings.HasPrefix(string(content), "From dev@example.net ") || strings.Count(string(content), "\nFrom dev@example.net ") != 2 ||
		!strings.Contains(string(content), "[PATCH v3 0/2]") || !strings.Contains(string(content), "[PATCH v3 2/2]") || !strings.HasPrefix(filename, "patch-series-") {
		t.Fatalf("filename=%q mbox=%q err=%v", filename, content, err)
	}
	if _, _, err := service.DownloadMbox("admin", patchOneID, false); !errors.Is(err, ErrForbidden) {
		t.Fatalf("administrator download error = %v", err)
	}
	if _, err := service.AddReviewComment("admin", patchOneID, ReviewCommentInput{Body: "Looks good."}); !errors.Is(err, ErrForbidden) {
		t.Fatalf("administrator comment error = %v", err)
	}
	if _, err := service.AddReviewComment("maintainer", patchOneID, ReviewCommentInput{Line: 1, Body: "This is not a diff line."}); !errors.Is(err, ErrInvalidComment) {
		t.Fatalf("invalid inline comment error = %v", err)
	}
	general, err := service.AddReviewComment("maintainer", patchOneID, ReviewCommentInput{Body: "Please add a test for this series."})
	if err != nil || general.Line != 0 || general.AuthorName != "MAINTAINER" {
		t.Fatalf("general comment=%#v err=%v", general, err)
	}
	inline, err := service.AddReviewComment("maintainer", patchOneID, ReviewCommentInput{Line: 3, Body: "Please keep this path stable."})
	if err != nil || inline.Path != "parser.go" || inline.LineText != "diff --git a/parser.go b/parser.go" {
		t.Fatalf("inline comment=%#v err=%v", inline, err)
	}
	resolved, err := service.ResolveReviewComment("owner", patchOneID, inline.ID, true)
	if err != nil || !resolved.Resolved {
		t.Fatalf("resolved comment=%#v err=%v", resolved, err)
	}
	withComments, err := service.Get(patchOneID)
	if err != nil || withComments.ReviewCommentCount != 2 || len(withComments.ReviewComments) != 2 || !withComments.ReviewComments[1].Resolved {
		t.Fatalf("patch comments=%#v err=%v", withComments.ReviewComments, err)
	}
	events, err := audit.NewRepository(database).Events(10)
	if err != nil || len(events) != 5 || events[0].Action != "patch.review-comment.resolve" ||
		events[1].Action != "patch.review-comment.create" || events[2].Action != "patch.review-comment.create" ||
		events[3].Action != "patch.series.download" || events[4].Action != "patch.review.accepted" {
		t.Fatalf("audit events=%#v err=%v", events, err)
	}
}
