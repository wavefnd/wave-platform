package patcharchive

import (
	"strings"
	"testing"
	"time"

	"github.com/wavefnd/wave-platform/internal/identifier"
	maildomain "github.com/wavefnd/wave-platform/internal/mail"
	"github.com/wavefnd/wave-platform/internal/mailbox"
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
