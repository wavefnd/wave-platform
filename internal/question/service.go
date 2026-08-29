package question

import (
	"errors"
	"fmt"
	"mime"
	stdmail "net/mail"
	"net/textproto"
	"regexp"
	"strings"
	"time"

	"github.com/wavefnd/wave-platform/internal/account"
	"github.com/wavefnd/wave-platform/internal/audit"
	"github.com/wavefnd/wave-platform/internal/identifier"
	maildomain "github.com/wavefnd/wave-platform/internal/mail"
	notificationdomain "github.com/wavefnd/wave-platform/internal/notification"
	"github.com/wavefnd/wave-platform/internal/storage"
)

var (
	ErrInvalidQuestion = errors.New("invalid question")
	ErrQuestionClosed  = errors.New("question is closed")
	ErrForbidden       = errors.New("question action is forbidden")
)

var tagPattern = regexp.MustCompile(`^[\p{L}\p{N}][\p{L}\p{N}-]{0,29}$`)

type Service struct {
	repository    *Repository
	mail          *maildomain.Repository
	audit         *audit.Repository
	notifications *notificationdomain.Service
	mailDomain    string
	now           func() time.Time
}

func NewService(database *storage.Database, mailDomain string) *Service {
	return &Service{repository: NewRepository(database), mail: maildomain.NewRepository(database),
		audit: audit.NewRepository(database), mailDomain: strings.ToLower(strings.TrimSpace(mailDomain)), now: time.Now}
}

func (service *Service) SetNotificationService(notifications *notificationdomain.Service) {
	service.notifications = notifications
}

func (service *Service) Create(actor account.Account, input CreateInput) (View, error) {
	title := strings.TrimSpace(input.Title)
	body := normalizeBody(input.Body)
	version := strings.TrimSpace(input.WaveVersion)
	platform := strings.TrimSpace(input.Platform)
	if len([]rune(title)) < 10 || len([]rune(title)) > 180 || strings.ContainsAny(title, "\r\n") {
		return View{}, fmt.Errorf("%w: title must contain between 10 and 180 characters", ErrInvalidQuestion)
	}
	if len([]rune(body)) < 20 || len([]rune(body)) > 30000 {
		return View{}, fmt.Errorf("%w: body must contain between 20 and 30000 characters", ErrInvalidQuestion)
	}
	if len([]rune(version)) > 40 || len([]rune(platform)) > 80 {
		return View{}, fmt.Errorf("%w: version or platform information is too long", ErrInvalidQuestion)
	}
	tags, err := normalizeTags(input.Tags)
	if err != nil {
		return View{}, err
	}
	questionID, err := identifier.New("question")
	if err != nil {
		return View{}, err
	}
	messageID, err := identifier.New("message")
	if err != nil {
		return View{}, err
	}
	now := service.now().UTC()
	message := maildomain.Message{ID: messageID, MessageID: "<" + messageID + "@" + service.mailDomain + ">",
		ThreadID: questionID, AuthorAccountID: actor.ID, From: addressOf(actor),
		To: []string{"questions@" + service.mailDomain}, Subject: title, ReceivedAt: now, CreatedAt: now}
	if err := service.mail.UpsertMessage(message, rawMessage(message, body, "")); err != nil {
		return View{}, err
	}
	value := Question{ID: questionID, RootMessageID: message.ID, Status: "open", Tags: tags,
		WaveVersion: version, Platform: platform}
	if err := service.repository.Upsert(value); err != nil {
		return View{}, err
	}
	_ = service.auditEvent(actor.ID, "question/"+questionID, "question.create")
	return service.repository.View(questionID, actor.ID)
}

func (service *Service) Answer(actor account.Account, input AnswerInput) (View, error) {
	value, err := service.repository.Question(input.QuestionID)
	if err != nil {
		return View{}, err
	}
	if value.Status == "closed" {
		return View{}, ErrQuestionClosed
	}
	body := normalizeBody(input.Body)
	if len([]rune(body)) < 1 || len([]rune(body)) > 20000 {
		return View{}, fmt.Errorf("%w: answer must contain between 1 and 20000 characters", ErrInvalidQuestion)
	}
	root, err := service.mail.Message(value.RootMessageID)
	if err != nil {
		return View{}, err
	}
	messageID, err := identifier.New("message")
	if err != nil {
		return View{}, err
	}
	now := service.now().UTC()
	message := maildomain.Message{ID: messageID, MessageID: "<" + messageID + "@" + service.mailDomain + ">",
		ThreadID: root.ThreadID, ParentMessageID: root.ID, AuthorAccountID: actor.ID, From: addressOf(actor),
		To: root.To, Subject: replySubject(root.Subject), ReceivedAt: now, CreatedAt: now}
	if err := service.mail.UpsertMessage(message, rawMessage(message, body, root.MessageID)); err != nil {
		return View{}, err
	}
	if value.Status == "open" {
		value.Status = "answered"
		if err := service.repository.Upsert(value); err != nil {
			return View{}, err
		}
	}
	_ = service.auditEvent(actor.ID, "question/"+value.ID+"/answer/"+message.ID, "question.answer")
	if service.notifications != nil {
		_, _ = service.notifications.Notify(notificationdomain.Input{RecipientAccountID: root.AuthorAccountID,
			ActorAccountID: actor.ID, ActorName: actor.DisplayName, Kind: "question.answer", Subject: root.Subject,
			URL: "/questions/" + value.ID})
	}
	return service.repository.View(value.ID, actor.ID)
}

func (service *Service) Vote(actorID, questionID, targetType, targetID string, value int) (int64, error) {
	question, err := service.repository.Question(questionID)
	if err != nil {
		return 0, err
	}
	var authorID string
	if targetType == "question" {
		if targetID != question.ID {
			return 0, fmt.Errorf("%w: vote target does not belong to the question", ErrInvalidQuestion)
		}
		root, err := service.mail.Message(question.RootMessageID)
		if err != nil {
			return 0, err
		}
		authorID = root.AuthorAccountID
	} else if targetType == "answer" {
		answer, err := service.mail.Message(targetID)
		if err != nil || answer.ThreadID != question.ID || answer.ID == question.RootMessageID {
			return 0, fmt.Errorf("%w: vote target does not belong to the question", ErrInvalidQuestion)
		}
		authorID = answer.AuthorAccountID
	} else {
		return 0, fmt.Errorf("%w: invalid vote target", ErrInvalidQuestion)
	}
	if authorID == actorID {
		return 0, fmt.Errorf("%w: authors cannot vote on their own content", ErrForbidden)
	}
	return service.repository.SetVote(targetType, targetID, actorID, value)
}

func (service *Service) Accept(actor account.Account, questionID, answerID string, administrator bool) (View, error) {
	value, err := service.repository.Question(questionID)
	if err != nil {
		return View{}, err
	}
	root, err := service.mail.Message(value.RootMessageID)
	if err != nil {
		return View{}, err
	}
	if root.AuthorAccountID != actor.ID && !administrator {
		return View{}, ErrForbidden
	}
	answerID = strings.TrimSpace(answerID)
	acceptedAuthorID := ""
	if answerID != "" {
		answer, err := service.mail.Message(answerID)
		if err != nil || answer.ThreadID != value.ID || answer.ID == value.RootMessageID {
			return View{}, fmt.Errorf("%w: accepted answer does not belong to the question", ErrInvalidQuestion)
		}
		value.AcceptedMessageID = answer.ID
		value.Status = "resolved"
		acceptedAuthorID = answer.AuthorAccountID
	} else {
		value.AcceptedMessageID = ""
		messages, err := service.mail.MessagesByThread(root.ThreadID)
		if err != nil {
			return View{}, err
		}
		value.Status = "open"
		if len(messages) > 1 {
			value.Status = "answered"
		}
	}
	if err := service.repository.Upsert(value); err != nil {
		return View{}, err
	}
	_ = service.auditEvent(actor.ID, "question/"+value.ID, "question.accept")
	if service.notifications != nil && acceptedAuthorID != "" {
		_, _ = service.notifications.Notify(notificationdomain.Input{RecipientAccountID: acceptedAuthorID,
			ActorAccountID: actor.ID, ActorName: actor.DisplayName, Kind: "question.accepted", Subject: root.Subject,
			URL: "/questions/" + value.ID})
	}
	return service.repository.View(value.ID, actor.ID)
}

func (service *Service) auditEvent(actorID, resourceID, action string) error {
	id, err := identifier.New("audit")
	if err != nil {
		return err
	}
	return service.audit.Append(audit.Event{ID: id, ActorID: "account/" + actorID, ResourceID: resourceID,
		Action: action, Result: "success", OccurredAt: service.now().UTC()})
}

func normalizeBody(value string) string {
	return strings.TrimSpace(strings.ReplaceAll(value, "\r\n", "\n"))
}

func normalizeTags(values []string) ([]string, error) {
	if len(values) < 1 || len(values) > 5 {
		return nil, fmt.Errorf("%w: between one and five tags are required", ErrInvalidQuestion)
	}
	seen := map[string]bool{}
	result := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.ToLower(strings.TrimSpace(value))
		if value == "" {
			continue
		}
		if !tagPattern.MatchString(value) {
			return nil, fmt.Errorf("%w: tags may contain letters, numbers, and hyphens", ErrInvalidQuestion)
		}
		if !seen[value] {
			seen[value] = true
			result = append(result, value)
		}
	}
	if len(result) == 0 {
		return nil, fmt.Errorf("%w: at least one tag is required", ErrInvalidQuestion)
	}
	return result, nil
}

func addressOf(actor account.Account) string {
	return (&stdmail.Address{Name: actor.DisplayName, Address: actor.Email}).String()
}

func replySubject(subject string) string {
	if strings.HasPrefix(strings.ToLower(subject), "re: ") {
		return subject
	}
	return "Re: " + subject
}

func rawMessage(message maildomain.Message, body, inReplyTo string) []byte {
	headers := textproto.MIMEHeader{}
	headers.Set("From", message.From)
	headers.Set("To", strings.Join(message.To, ", "))
	headers.Set("Subject", mime.QEncoding.Encode("utf-8", message.Subject))
	headers.Set("Date", message.CreatedAt.Format(time.RFC1123Z))
	headers.Set("Message-ID", message.MessageID)
	if inReplyTo != "" {
		headers.Set("In-Reply-To", inReplyTo)
		headers.Set("References", inReplyTo)
	}
	headers.Set("MIME-Version", "1.0")
	headers.Set("Content-Type", "text/plain; charset=utf-8")
	headers.Set("Content-Transfer-Encoding", "8bit")
	order := []string{"From", "To", "Subject", "Date", "Message-ID", "In-Reply-To", "References", "MIME-Version", "Content-Type", "Content-Transfer-Encoding"}
	var raw strings.Builder
	for _, name := range order {
		if value := headers.Get(name); value != "" {
			raw.WriteString(name + ": " + value + "\r\n")
		}
	}
	raw.WriteString("\r\n")
	raw.WriteString(strings.ReplaceAll(body, "\n", "\r\n"))
	return []byte(raw.String())
}
