package community_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/wavefnd/wave-platform/internal/community"
	"github.com/wavefnd/wave-platform/internal/mail"
	"github.com/wavefnd/wave-platform/internal/question"
	"github.com/wavefnd/wave-platform/internal/storage"
)

func TestCommunityAndQuestionResolveTheSameMailThread(t *testing.T) {
	database, err := storage.Open(t.TempDir())
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	t.Cleanup(func() { _ = database.Close() })

	createdAt := time.Date(2026, 8, 5, 10, 0, 0, 0, time.UTC)
	mailRepository := mail.NewRepository(database)
	root := mail.Message{
		ID: "message/root", MessageID: "<root@wave-lang.dev>", ThreadID: "thread/one",
		From: "Luna <luna@wave-lang.dev>", To: []string{"community@wave-lang.dev"},
		Subject: "Wave thread", CreatedAt: createdAt, ReceivedAt: createdAt,
	}
	reply := mail.Message{
		ID: "message/reply", MessageID: "<reply@wave-lang.dev>", ThreadID: root.ThreadID,
		ParentMessageID: root.ID, From: "Wave User <user@wave-lang.dev>",
		To: []string{"community@wave-lang.dev"}, Subject: "Re: Wave thread",
		CreatedAt: createdAt.Add(time.Minute), ReceivedAt: createdAt.Add(time.Minute),
	}
	for _, item := range []struct {
		message mail.Message
		body    string
	}{{root, "Root message body"}, {reply, "Reply body"}} {
		raw := []byte(fmt.Sprintf("Subject: %s\r\nContent-Type: text/plain; charset=utf-8\r\n\r\n%s", item.message.Subject, item.body))
		if err := mailRepository.UpsertMessage(item.message, raw); err != nil {
			t.Fatalf("store mail message: %v", err)
		}
	}

	communityRepository := community.NewRepository(database)
	if err := communityRepository.UpsertThread(community.Thread{
		ID: "wave-thread", SpaceID: "general", RootMessageID: root.ID, Tags: []string{"wave"},
	}); err != nil {
		t.Fatalf("store community thread: %v", err)
	}
	communityView, err := communityRepository.View("wave-thread")
	if err != nil {
		t.Fatalf("community view: %v", err)
	}
	if communityView.Root.Body != "Root message body" || len(communityView.Replies) != 1 || communityView.Replies[0].Body != "Reply body" {
		t.Fatalf("community did not resolve mail thread: %#v", communityView)
	}

	questionRepository := question.NewRepository(database)
	if err := questionRepository.Upsert(question.Question{ID: "wave-question", RootMessageID: root.ID}); err != nil {
		t.Fatalf("store question: %v", err)
	}
	questionView, err := questionRepository.View("wave-question", "")
	if err != nil {
		t.Fatalf("question view: %v", err)
	}
	if questionView.Root.Body != communityView.Root.Body || len(questionView.Answers) != 1 {
		t.Fatalf("question did not resolve shared mail thread: %#v", questionView)
	}

}
