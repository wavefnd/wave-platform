package community_test

import (
	"encoding/base64"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/wavefnd/wave-platform/internal/community"
	"github.com/wavefnd/wave-platform/internal/mailbox"
	"github.com/wavefnd/wave-platform/internal/storage"
	"github.com/wavefnd/wave-platform/internal/testsupport"
	webhookdomain "github.com/wavefnd/wave-platform/internal/webhook"
)

func TestMailBackedPostReplyStayOutOfPersonalMailbox(t *testing.T) {
	database, err := storage.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = database.Close() })
	if err := community.SeedSpaces(database); err != nil {
		t.Fatal(err)
	}

	identities, err := testsupport.NewIdentity(database)
	if err != nil {
		t.Fatal(err)
	}
	author, err := testsupport.Register(identities, "John Mark")
	if err != nil {
		t.Fatal(err)
	}
	commenter, err := testsupport.Register(identities, "Wave User")
	if err != nil {
		t.Fatal(err)
	}

	service := community.NewService(database, "wave-lang.dev")
	post, err := service.CreatePost(author, community.CreatePostInput{
		SpaceID: "development", Title: "Wave compiler backend design", Body: "A mail-backed community post.", Tags: []string{"compiler", "design"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if post.Root.AuthorAccountID != author.ID || post.Root.Body != "A mail-backed community post." {
		t.Fatalf("post=%#v", post)
	}

	replied, err := service.AddReply(commenter, community.CreateReplyInput{ThreadID: post.Thread.ID, Body: "This reply is also a mail message."})
	if err != nil {
		t.Fatal(err)
	}
	if len(replied.Replies) != 1 || replied.Replies[0].AuthorAccountID != commenter.ID {
		t.Fatalf("replied=%#v", replied)
	}
	nested, err := service.AddReply(author, community.CreateReplyInput{
		ThreadID: post.Thread.ID, ParentMessageID: replied.Replies[0].ID, Body: "This reply belongs below the comment.",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(nested.Replies) != 2 || nested.Replies[1].ParentMessageID != replied.Replies[0].ID {
		t.Fatalf("nested reply parent was not preserved: %#v", nested.Replies)
	}

	score, err := service.Vote(commenter.ID, post.Thread.ID, "thread", post.Thread.ID, 1)
	if err != nil || score != 1 {
		t.Fatalf("vote score=%d err=%v", score, err)
	}
	score, err = service.Vote(commenter.ID, post.Thread.ID, "thread", post.Thread.ID, 0)
	if err != nil || score != 0 {
		t.Fatalf("removed vote score=%d err=%v", score, err)
	}

	box, err := identities.Mailbox(author.ID)
	if err != nil {
		t.Fatal(err)
	}
	mailboxRepository := mailbox.NewRepository(database)
	inbox, err := mailboxRepository.Entries(box.ID, "Inbox")
	if err != nil {
		t.Fatal(err)
	}
	sent, err := mailboxRepository.Entries(box.ID, "Sent")
	if err != nil {
		t.Fatal(err)
	}
	commenterBox, err := identities.Mailbox(commenter.ID)
	if err != nil {
		t.Fatal(err)
	}
	commenterSent, err := mailboxRepository.Entries(commenterBox.ID, "Sent")
	if err != nil {
		t.Fatal(err)
	}
	if len(inbox) != 0 || len(sent) != 0 || len(commenterSent) != 0 {
		t.Fatalf("community leaked into personal mailbox: inbox=%#v sent=%#v commenter-sent=%#v", inbox, sent, commenterSent)
	}

	legacyEntries := []mailbox.Entry{
		{ID: "legacy-community-sent", MailboxID: box.ID, MessageID: post.Root.ID, Folder: "Sent", CreatedAt: time.Now().UTC()},
		{ID: "legacy-community-inbox", MailboxID: box.ID, MessageID: replied.Replies[0].ID, Folder: "Inbox", CreatedAt: time.Now().UTC()},
	}
	for _, entry := range legacyEntries {
		if err := mailboxRepository.AddEntry(entry); err != nil {
			t.Fatal(err)
		}
	}
	removed, err := community.CleanupMailboxProjections(database)
	if err != nil || removed != len(legacyEntries) {
		t.Fatalf("cleanup removed=%d err=%v", removed, err)
	}
	remaining, err := mailboxRepository.Entries(box.ID, "")
	if err != nil || len(remaining) != 0 {
		t.Fatalf("remaining=%#v err=%v", remaining, err)
	}

	stored, err := database.Get(storage.Key("community", "thread", "object", post.Thread.ID))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(stored), "mail-backed community post") {
		t.Fatal("community metadata duplicated the mail body")
	}
}

func TestCommunityAndFounderPostsQueueDistinctWebhookEvents(t *testing.T) {
	database, err := storage.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = database.Close() })
	if err := community.SeedSpaces(database); err != nil {
		t.Fatal(err)
	}
	identities, err := testsupport.NewIdentity(database)
	if err != nil {
		t.Fatal(err)
	}
	owner, _, err := identities.BootstrapTOTPAdmin("Wave Owner", "wave-owner", "owner@example.net", testsupport.TOTPSecret)
	if err != nil {
		t.Fatal(err)
	}
	member, err := testsupport.Register(identities, "Community Writer")
	if err != nil {
		t.Fatal(err)
	}
	key := base64.StdEncoding.EncodeToString([]byte("0123456789abcdef0123456789abcdef"))
	webhooks, err := webhookdomain.NewService(database, key, "https://wave-lang.dev")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := webhooks.SaveEndpoint(owner.ID, webhookdomain.EndpointInput{Name: "Discord feed", Kind: "generic", URL: "https://hooks.example.test/wave",
		Events: []string{webhookdomain.EventCommunityPost, webhookdomain.EventFounderPost}, Enabled: true}); err != nil {
		t.Fatal(err)
	}
	service := community.NewService(database, "wave-lang.dev")
	service.SetWebhookService(webhooks)
	imagePath := "/media/lunastev/image-1787312400000-0123456789abcdef0123456789abcdef.webp"
	if _, err := service.CreatePost(member, community.CreatePostInput{SpaceID: "development", Title: "Community compiler update",
		Body: "A public community update.\n\n![Ignored community image](" + imagePath + ")"}); err != nil {
		t.Fatal(err)
	}
	if _, err := service.CreatePost(owner, community.CreatePostInput{SpaceID: "founder-notes", Title: "A note from the Wave founder",
		Body: "A public founder update.\n\n![External image](https://example.net/tracker.png)\n![Compiler graph](" + imagePath + ")"}); err != nil {
		t.Fatal(err)
	}
	deliveries, err := webhooks.Deliveries(10)
	if err != nil {
		t.Fatal(err)
	}
	if len(deliveries) != 2 || deliveries[0].EventType != webhookdomain.EventFounderPost || !strings.Contains(deliveries[0].Summary, "A public founder update.") || strings.Contains(deliveries[0].Summary, "![") ||
		deliveries[0].AuthorName != owner.DisplayName || deliveries[0].ImageURL != "https://wave-lang.dev"+imagePath ||
		deliveries[1].EventType != webhookdomain.EventCommunityPost || !strings.Contains(deliveries[1].Summary, "A public community update.") || deliveries[1].AuthorName != member.DisplayName || deliveries[1].ImageURL != "" {
		t.Fatalf("deliveries = %#v", deliveries)
	}
}

func TestOwnerSpacesRestrictPostsButAllowComments(t *testing.T) {
	database, err := storage.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = database.Close() })
	if err := community.SeedSpaces(database); err != nil {
		t.Fatal(err)
	}

	identities, err := testsupport.NewIdentity(database)
	if err != nil {
		t.Fatal(err)
	}
	owner, _, err := identities.BootstrapTOTPAdmin("Wave Owner", "wave-owner", "owner@example.net", testsupport.TOTPSecret)
	if err != nil {
		t.Fatal(err)
	}
	reader, err := testsupport.Register(identities, "Community Reader")
	if err != nil {
		t.Fatal(err)
	}
	service := community.NewService(database, "wave-lang.dev")

	if _, err := service.CreatePost(reader, community.CreatePostInput{SpaceID: "founder-notes", Title: "Reader cannot publish here", Body: "This must be rejected."}); !errors.Is(err, community.ErrPostingRestricted) {
		t.Fatalf("restricted post error=%v", err)
	}
	post, err := service.CreatePost(owner, community.CreatePostInput{SpaceID: "development-log", Title: "Compiler work completed today", Body: "Implemented and tested the parser changes."})
	if err != nil {
		t.Fatal(err)
	}
	view, err := service.AddReply(reader, community.CreateReplyInput{ThreadID: post.Thread.ID, Body: "Thanks for sharing the progress."})
	if err != nil || len(view.Replies) != 1 || view.Replies[0].AuthorAccountID != reader.ID {
		t.Fatalf("comment view=%#v err=%v", view, err)
	}
}
