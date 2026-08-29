package question

import (
	"errors"
	"testing"

	notificationdomain "github.com/wavefnd/wave-platform/internal/notification"
	"github.com/wavefnd/wave-platform/internal/storage"
	"github.com/wavefnd/wave-platform/internal/testsupport"
)

func TestQuestionLifecycleUsesMailThreadWithoutMailboxProjection(t *testing.T) {
	database, err := storage.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	identityService, err := testsupport.NewIdentity(database)
	if err != nil {
		t.Fatal(err)
	}
	asker, err := testsupport.Register(identityService, "Ada Lovelace")
	if err != nil {
		t.Fatal(err)
	}
	answerer, err := testsupport.Register(identityService, "Grace Hopper")
	if err != nil {
		t.Fatal(err)
	}
	service := NewService(database, "wave-lang.dev")
	notifications := notificationdomain.NewService(database)
	service.SetNotificationService(notifications)

	created, err := service.Create(asker, CreateInput{Title: "Why does generic inference fail here?",
		Body: "The compiler reports a type mismatch for this generic function call.",
		Tags: []string{"compiler", "generics"}, WaveVersion: "0.2.0-pre-beta", Platform: "Linux x86_64"})
	if err != nil {
		t.Fatal(err)
	}
	if created.Question.Status != "open" || created.Root.AuthorAccountID != asker.ID {
		t.Fatalf("created = %#v", created)
	}
	answered, err := service.Answer(answerer, AnswerInput{QuestionID: created.Question.ID, Body: "Add an explicit type argument until inference is fixed."})
	if err != nil {
		t.Fatal(err)
	}
	if answered.Question.Status != "answered" || len(answered.Answers) != 1 {
		t.Fatalf("answered = %#v", answered)
	}
	askerNotifications, unread, err := notifications.List(asker.ID, 20)
	if err != nil || unread != 1 || len(askerNotifications) != 1 || askerNotifications[0].Kind != "question.answer" {
		t.Fatalf("asker notifications=%#v unread=%d err=%v", askerNotifications, unread, err)
	}
	answerID := answered.Answers[0].ID
	if _, err := service.Vote(answerer.ID, created.Question.ID, "question", created.Question.ID, 1); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Vote(answerer.ID, created.Question.ID, "answer", answerID, 1); !errors.Is(err, ErrForbidden) {
		t.Fatalf("self vote error = %v", err)
	}
	accepted, err := service.Accept(asker, created.Question.ID, answerID, false)
	if err != nil {
		t.Fatal(err)
	}
	if accepted.Question.Status != "resolved" || accepted.Question.AcceptedMessageID != answerID || !accepted.Answers[0].Accepted {
		t.Fatalf("accepted = %#v", accepted)
	}
	answererNotifications, unread, err := notifications.List(answerer.ID, 20)
	if err != nil || unread != 1 || len(answererNotifications) != 1 || answererNotifications[0].Kind != "question.accepted" {
		t.Fatalf("answerer notifications=%#v unread=%d err=%v", answererNotifications, unread, err)
	}
	if _, err := service.Accept(answerer, created.Question.ID, "", false); !errors.Is(err, ErrForbidden) {
		t.Fatalf("non-owner accept error = %v", err)
	}

	for _, accountID := range []string{asker.ID, answerer.ID} {
		for _, folder := range []string{"Inbox", "Sent"} {
			items, err := identityService.MailboxItems(accountID, folder)
			if err != nil || len(items) != 0 {
				t.Fatalf("account=%s folder=%s items=%d err=%v", accountID, folder, len(items), err)
			}
		}
	}
}

func TestQuestionQueryFiltersUnansweredAndSearch(t *testing.T) {
	database, err := storage.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	identityService, err := testsupport.NewIdentity(database)
	if err != nil {
		t.Fatal(err)
	}
	asker, err := testsupport.Register(identityService, "Ada Lovelace")
	if err != nil {
		t.Fatal(err)
	}
	service := NewService(database, "wave-lang.dev")
	created, err := service.Create(asker, CreateInput{Title: "How can I call a native C library?",
		Body: "I need to bind a native library and pass a string safely from Wave.", Tags: []string{"ffi"}})
	if err != nil {
		t.Fatal(err)
	}
	items, err := NewRepository(database).Query("native", "unanswered", "ffi", 30, 0, asker.ID)
	if err != nil || len(items) != 1 || items[0].ID != created.Question.ID {
		t.Fatalf("items=%#v err=%v", items, err)
	}
}
