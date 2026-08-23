package mailinglist

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/wavefnd/wave-platform/internal/account"
	"github.com/wavefnd/wave-platform/internal/audit"
	maildomain "github.com/wavefnd/wave-platform/internal/mail"
	"github.com/wavefnd/wave-platform/internal/mailbox"
	"github.com/wavefnd/wave-platform/internal/permission"
	"github.com/wavefnd/wave-platform/internal/storage"
	webhookdomain "github.com/wavefnd/wave-platform/internal/webhook"
)

type recordingWebhookPublisher struct {
	events []webhookdomain.Event
	err    error
}

func (publisher *recordingWebhookPublisher) Publish(event webhookdomain.Event) error {
	publisher.events = append(publisher.events, event)
	return publisher.err
}

func TestInternalMailingListSubscriptionThreadsAndPolicies(t *testing.T) {
	database, err := storage.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = database.Close() })
	now := time.Date(2026, 8, 23, 12, 0, 0, 0, time.UTC)
	accounts := account.NewRepository(database)
	for _, item := range []account.Account{
		{ID: "owner", Username: "owner", DisplayName: "Owner", Email: "owner@wave-lang.dev", Status: "active", CreatedAt: now, UpdatedAt: now},
		{ID: "member", Username: "member", DisplayName: "Member", Email: "member@wave-lang.dev", Status: "active", CreatedAt: now, UpdatedAt: now},
		{ID: "external", Username: "external", DisplayName: "External", Email: "external@example.net", Status: "active", CreatedAt: now, UpdatedAt: now},
		{ID: "suspended", Username: "suspended", DisplayName: "Suspended", Email: "suspended@wave-lang.dev", Status: "suspended", CreatedAt: now, UpdatedAt: now},
	} {
		if err := accounts.Create(item); err != nil {
			t.Fatal(err)
		}
	}
	if err := permission.NewRepository(database).Assign(permission.Assignment{AccountID: "owner", RoleID: "platform-owner", Scope: "platform"}); err != nil {
		t.Fatal(err)
	}

	patchBox := mailbox.Mailbox{ID: "existing-patch-box", AccountID: "service/patchs", Address: "patchs@wave-lang.dev", CreatedAt: now}
	if err := mailbox.NewRepository(database).UpsertMailbox(patchBox); err != nil {
		t.Fatal(err)
	}
	service := NewService(database, "wave-lang.dev", "service/patchs")
	service.now = func() time.Time { return now }
	if err := service.EnsureDefaults(); err != nil {
		t.Fatal(err)
	}
	lists, err := service.Lists("member")
	if err != nil || len(lists) != 3 {
		t.Fatalf("lists=%#v err=%v", lists, err)
	}
	patchList, err := service.repository.List("patchs")
	if err != nil || patchList.MailboxID != patchBox.ID || patchList.WebhookPolicy != WebhookSummary {
		t.Fatalf("patch list=%#v err=%v", patchList, err)
	}
	announce, err := service.repository.List("announce")
	if err != nil || announce.WebhookPolicy != WebhookFull || announce.WebhookPreviewLimit != 500 {
		t.Fatalf("announce=%#v err=%v", announce, err)
	}
	if _, err := service.Lists("external"); !errors.Is(err, ErrForbidden) {
		t.Fatalf("external account error=%v", err)
	}
	if err := service.Subscribe("suspended", "development", true); !errors.Is(err, ErrForbidden) {
		t.Fatalf("suspended subscription error=%v", err)
	}
	if _, err := service.Post("member", "development", PostInput{Subject: "Wave ABI", Body: "Discuss the stable Wave ABI."}); !errors.Is(err, ErrNotSubscribed) {
		t.Fatalf("unsubscribed post error=%v", err)
	}
	if err := service.Subscribe("member", "development", true); err != nil {
		t.Fatal(err)
	}
	created, err := service.Post("member", "development", PostInput{Subject: "Stable Wave ABI", Body: "Discuss a stable ABI for Wave modules and hosts."})
	if err != nil || len(created.Messages) != 1 {
		t.Fatalf("created=%#v err=%v", created, err)
	}
	root := created.Messages[0]
	stored, err := maildomain.NewRepository(database).Message(root.MessageID)
	if err != nil || stored.ThreadID != root.MessageID || stored.AuthorAccountID != "member" {
		t.Fatalf("stored=%#v err=%v", stored, err)
	}
	development, _ := service.repository.List("development")
	entry, err := mailbox.NewRepository(database).Entry(development.MailboxID, root.EntryID)
	if err != nil || entry.MessageID != stored.ID {
		t.Fatalf("entry=%#v err=%v", entry, err)
	}
	replied, err := service.Reply("member", "development", created.ID, ReplyInput{ParentMessageID: root.MessageID, Body: "The first revision should cover the native calling convention."})
	if err != nil || len(replied.Messages) != 2 || replied.Messages[1].ParentMessageID != root.MessageID {
		t.Fatalf("replied=%#v err=%v", replied, err)
	}
	replyMessage, err := maildomain.NewRepository(database).Message(replied.Messages[1].MessageID)
	if err != nil {
		t.Fatal(err)
	}
	rawReply, err := maildomain.NewRepository(database).RawMessage(replyMessage)
	if err != nil || !strings.Contains(string(rawReply), "In-Reply-To: "+root.HeaderMessageID) ||
		!strings.Contains(string(rawReply), "References: "+root.HeaderMessageID) {
		t.Fatalf("reply headers=%q err=%v", rawReply, err)
	}
	threads, err := service.Threads("member", "development", "calling convention", 30, 0)
	if err != nil || len(threads) != 1 || threads[0].MessageCount != 2 {
		t.Fatalf("threads=%#v err=%v", threads, err)
	}

	if err := service.Subscribe("member", "announce", true); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Post("member", "announce", PostInput{Subject: "Release", Body: "A release is available."}); !errors.Is(err, ErrForbidden) {
		t.Fatalf("member announcement error=%v", err)
	}
	if err := service.Subscribe("owner", "announce", true); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Post("owner", "announce", PostInput{Subject: "Wave release", Body: "A new Wave release is now available."}); err != nil {
		t.Fatal(err)
	}

	deliveries, err := maildomain.NewRepository(database).Deliveries(100)
	if err != nil || len(deliveries) != 0 {
		t.Fatalf("mailing lists must not create remote deliveries: %#v err=%v", deliveries, err)
	}
	events, err := audit.NewRepository(database).Events(20)
	if err != nil || !hasAuditAction(events, "mailing-list.post") || !hasAuditAction(events, "mailing-list.reply") || !hasAuditAction(events, "mailing-list.subscribe") {
		t.Fatalf("audit events=%#v err=%v", events, err)
	}
}

func TestMailingListWebhookPoliciesPublishWithoutAffectingStoredMessages(t *testing.T) {
	database, err := storage.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = database.Close() })
	now := time.Date(2026, 8, 23, 13, 0, 0, 0, time.UTC)
	accounts := account.NewRepository(database)
	for _, item := range []account.Account{
		{ID: "owner", Username: "owner", DisplayName: "Owner", Email: "owner@wave-lang.dev", Status: "active", CreatedAt: now, UpdatedAt: now},
		{ID: "member", Username: "member", DisplayName: "Wave Member", Email: "member@wave-lang.dev", Status: "active", CreatedAt: now, UpdatedAt: now},
	} {
		if err := accounts.Create(item); err != nil {
			t.Fatal(err)
		}
	}
	if err := permission.NewRepository(database).Assign(permission.Assignment{AccountID: "owner", RoleID: "platform-owner", Scope: "platform"}); err != nil {
		t.Fatal(err)
	}
	service := NewService(database, "wave-lang.dev", "service/patchs")
	service.now = func() time.Time { return now }
	publisher := &recordingWebhookPublisher{err: errors.New("delivery queue unavailable")}
	service.webhooks = publisher
	if err := service.EnsureDefaults(); err != nil {
		t.Fatal(err)
	}

	development, err := service.repository.List("development")
	if err != nil {
		t.Fatal(err)
	}
	development.WebhookPreviewLimit = 4
	if err := service.repository.UpsertList(development); err != nil {
		t.Fatal(err)
	}
	if err := service.Subscribe("member", "development", true); err != nil {
		t.Fatal(err)
	}
	thread, err := service.Post("member", "development", PostInput{Subject: "Unicode preview", Body: "가나다라마바사"})
	if err != nil {
		t.Fatalf("stored post must survive webhook failure: %v", err)
	}
	if _, err := service.Reply("member", "development", thread.ID, ReplyInput{Body: "123456"}); err != nil {
		t.Fatalf("stored reply must survive webhook failure: %v", err)
	}
	if len(publisher.events) != 2 {
		t.Fatalf("summary events = %#v", publisher.events)
	}
	first := publisher.events[0]
	if first.Type != webhookdomain.EventMailingListPost || first.Title != "Unicode preview" || first.Summary != "가나다라…" ||
		first.AuthorName != "Wave Member" || first.ResourceID != "mailing-list/development/thread/"+thread.ID ||
		first.URL != "/mail/lists/development/thread/"+thread.ID {
		t.Fatalf("first summary event = %#v", first)
	}
	if publisher.events[1].Summary != "1234…" || publisher.events[1].URL != first.URL {
		t.Fatalf("reply summary event = %#v", publisher.events[1])
	}
	stored, err := service.repository.Thread("development", thread.ID)
	if err != nil || len(stored.Messages) != 2 {
		t.Fatalf("stored thread after webhook failures=%#v err=%v", stored, err)
	}

	announce, err := service.repository.List("announce")
	if err != nil {
		t.Fatal(err)
	}
	announce.WebhookPreviewLimit = 2
	if err := service.repository.UpsertList(announce); err != nil {
		t.Fatal(err)
	}
	if err := service.Subscribe("owner", "announce", true); err != nil {
		t.Fatal(err)
	}
	fullBody := "This complete announcement body is published."
	if _, err := service.Post("owner", "announce", PostInput{Subject: "Wave release", Body: fullBody}); err != nil {
		t.Fatal(err)
	}
	if len(publisher.events) != 3 || publisher.events[2].Summary != fullBody {
		t.Fatalf("full-policy event = %#v", publisher.events)
	}

	development.WebhookPolicy = WebhookDisabled
	if err := service.repository.UpsertList(development); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Post("member", "development", PostInput{Subject: "Private follow-up", Body: "Internal only."}); err != nil {
		t.Fatal(err)
	}
	if len(publisher.events) != 3 {
		t.Fatalf("disabled policy published an event: %#v", publisher.events)
	}

	if err := service.Subscribe("member", "patchs", true); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Post("member", "patchs", PostInput{Subject: "Not a patch", Body: "Plain discussion text."}); !errors.Is(err, ErrInvalidPost) {
		t.Fatalf("invalid patch thread error = %v", err)
	}
	patchBody := "Keep parser ranges.\n\ndiff --git a/parser.go b/parser.go\n--- a/parser.go\n+++ b/parser.go\n@@ -1 +1 @@\n-old\n+new\n"
	patchThread, err := service.Post("member", "patchs", PostInput{Subject: "[PATCH] parser: keep ranges", Body: patchBody})
	if err != nil {
		t.Fatal(err)
	}
	if len(publisher.events) != 4 || publisher.events[3].Type != webhookdomain.EventPatchReceived ||
		publisher.events[3].ResourceID != "patch/"+patchThread.ID || publisher.events[3].URL != "/mail/lists/patchs/patch/"+patchThread.ID {
		t.Fatalf("patch event = %#v", publisher.events)
	}
	for index := 3; index < maxPostsPerAccountPerListPerDay; index++ {
		if err := service.repository.ReservePosting(development, "member", fmt.Sprintf("reserved-%d", index), now); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := service.Post("member", "development", PostInput{Subject: "Over limit", Body: "This request must be rate limited."}); !errors.Is(err, ErrRateLimited) {
		t.Fatalf("posting rate limit error = %v", err)
	}
}

func hasAuditAction(events []audit.Event, action string) bool {
	for _, event := range events {
		if event.Action == action {
			return true
		}
	}
	return false
}
