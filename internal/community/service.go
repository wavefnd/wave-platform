package community

import (
	"errors"
	"fmt"
	"mime"
	stdmail "net/mail"
	"net/textproto"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/wavefnd/wave-platform/internal/account"
	"github.com/wavefnd/wave-platform/internal/audit"
	"github.com/wavefnd/wave-platform/internal/identifier"
	maildomain "github.com/wavefnd/wave-platform/internal/mail"
	notificationdomain "github.com/wavefnd/wave-platform/internal/notification"
	"github.com/wavefnd/wave-platform/internal/permission"
	"github.com/wavefnd/wave-platform/internal/storage"
	webhookdomain "github.com/wavefnd/wave-platform/internal/webhook"
)

var (
	ErrInvalidPost       = errors.New("invalid community post")
	ErrThreadLocked      = errors.New("community thread is locked")
	ErrPostingRestricted = errors.New("only the platform owner can publish in this space")
	ErrEnglishRequired   = errors.New("community posts and comments must be written in English")
)

var (
	tagPattern                   = regexp.MustCompile(`^[\p{L}\p{N}][\p{L}\p{N}-]{0,29}$`)
	englishTagPattern            = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,29}$`)
	lunaStevMarkdownImagePattern = regexp.MustCompile(`!\[[^\]\r\n]*\]\((/media/lunastev/image-[0-9]+-[0-9a-f]{32}\.webp)[ \t]*\)`)
)

type Service struct {
	repository    *Repository
	mail          *maildomain.Repository
	audit         *audit.Repository
	permission    *permission.Repository
	webhooks      *webhookdomain.Service
	notifications *notificationdomain.Service
	mailDomain    string
	now           func() time.Time
}

func NewService(database *storage.Database, mailDomain string) *Service {
	return &Service{repository: NewRepository(database), mail: maildomain.NewRepository(database),
		audit: audit.NewRepository(database), permission: permission.NewRepository(database),
		mailDomain: strings.ToLower(strings.TrimSpace(mailDomain)), now: time.Now}
}

func (service *Service) SetWebhookService(webhooks *webhookdomain.Service) {
	service.webhooks = webhooks
}

func (service *Service) SetNotificationService(notifications *notificationdomain.Service) {
	service.notifications = notifications
}

func (service *Service) CreatePost(actor account.Account, input CreatePostInput) (ThreadView, error) {
	input.SpaceID = strings.TrimSpace(input.SpaceID)
	input.Title = strings.TrimSpace(input.Title)
	input.Body = normalizeBody(input.Body)
	space, err := service.repository.Space(input.SpaceID)
	if err != nil {
		return ThreadView{}, fmt.Errorf("%w: unknown space", ErrInvalidPost)
	}
	if space.PostingPolicy == "owner" {
		allowed, roleErr := service.permission.HasRole(actor.ID, "platform-owner")
		if roleErr != nil {
			return ThreadView{}, roleErr
		}
		if !allowed {
			return ThreadView{}, ErrPostingRestricted
		}
	}
	englishOnly := space.PostingPolicy != "owner"
	if englishOnly && (!englishProse(input.Title) || !englishProse(withoutMarkdownCode(input.Body))) {
		return ThreadView{}, ErrEnglishRequired
	}
	if len([]rune(input.Title)) < 5 || len([]rune(input.Title)) > 180 || strings.ContainsAny(input.Title, "\r\n") {
		return ThreadView{}, fmt.Errorf("%w: title must contain between 5 and 180 characters", ErrInvalidPost)
	}
	if len([]rune(input.Body)) < 1 || len([]rune(input.Body)) > 20000 {
		return ThreadView{}, fmt.Errorf("%w: body must contain between 1 and 20000 characters", ErrInvalidPost)
	}
	tags, err := normalizeTags(input.Tags, englishOnly)
	if err != nil {
		return ThreadView{}, err
	}
	threadID, err := identifier.New("community")
	if err != nil {
		return ThreadView{}, err
	}
	messageID, err := identifier.New("message")
	if err != nil {
		return ThreadView{}, err
	}
	now := service.now().UTC()
	message := maildomain.Message{ID: messageID, MessageID: "<" + messageID + "@" + service.mailDomain + ">",
		ThreadID: threadID, AuthorAccountID: actor.ID, From: addressOf(actor),
		To: []string{"community+" + input.SpaceID + "@" + service.mailDomain}, Subject: input.Title,
		ReceivedAt: now, CreatedAt: now}
	if err := service.mail.UpsertMessage(message, rawMessage(message, input.Body, "")); err != nil {
		return ThreadView{}, err
	}
	thread := Thread{ID: threadID, SpaceID: input.SpaceID, RootMessageID: message.ID, Tags: tags}
	if err := service.repository.UpsertThread(thread); err != nil {
		return ThreadView{}, err
	}
	if err := service.repository.SetSubscribed(thread.ID, actor.ID, true); err != nil {
		return ThreadView{}, err
	}
	_ = service.auditEvent(actor.ID, "community/thread/"+thread.ID, "community.post")
	if service.webhooks != nil {
		eventType, path := webhookdomain.EventCommunityPost, "/community/thread/"+thread.ID
		imageURL := ""
		if space.PostingPolicy == "owner" {
			eventType, path = webhookdomain.EventFounderPost, "/lunastev/thread/"+thread.ID
			imageURL = firstLunaStevImage(input.Body)
		} else if space.ID == "showcase" {
			path = "/community/showcase/" + thread.ID
		}
		_ = service.webhooks.Publish(webhookdomain.Event{Type: eventType, Title: input.Title, Summary: input.Body, AuthorName: actor.DisplayName,
			ImageURL: imageURL, ResourceID: "community/thread/" + thread.ID, URL: path, OccurredAt: now})
	}
	return service.repository.ViewFor(thread.ID, actor.ID)
}

func firstLunaStevImage(markdown string) string {
	match := lunaStevMarkdownImagePattern.FindStringSubmatch(markdown)
	if len(match) == 2 {
		return match[1]
	}
	return ""
}

func (service *Service) AddReply(actor account.Account, input CreateReplyInput) (ThreadView, error) {
	thread, err := service.repository.Thread(input.ThreadID)
	if err != nil {
		return ThreadView{}, err
	}
	if thread.Locked {
		return ThreadView{}, ErrThreadLocked
	}
	space, err := service.repository.Space(thread.SpaceID)
	if err != nil {
		return ThreadView{}, err
	}
	body := normalizeBody(input.Body)
	if len([]rune(body)) < 1 || len([]rune(body)) > 10000 {
		return ThreadView{}, fmt.Errorf("%w: reply must contain between 1 and 10000 characters", ErrInvalidPost)
	}
	if space.PostingPolicy != "owner" && !englishProse(withoutMarkdownCode(body)) {
		return ThreadView{}, ErrEnglishRequired
	}
	root, err := service.mail.Message(thread.RootMessageID)
	if err != nil {
		return ThreadView{}, err
	}
	parentID := strings.TrimSpace(input.ParentMessageID)
	if parentID == "" {
		parentID = root.ID
	}
	parent, err := service.mail.Message(parentID)
	if err != nil || parent.ThreadID != root.ThreadID {
		return ThreadView{}, fmt.Errorf("%w: reply parent does not belong to the post", ErrInvalidPost)
	}
	messageID, err := identifier.New("message")
	if err != nil {
		return ThreadView{}, err
	}
	now := service.now().UTC()
	message := maildomain.Message{ID: messageID, MessageID: "<" + messageID + "@" + service.mailDomain + ">",
		ThreadID: root.ThreadID, ParentMessageID: parent.ID, AuthorAccountID: actor.ID,
		From: addressOf(actor), To: root.To, Subject: replySubject(root.Subject), ReceivedAt: now, CreatedAt: now}
	if err := service.mail.UpsertMessage(message, rawMessage(message, body, parent.MessageID)); err != nil {
		return ThreadView{}, err
	}
	if err := service.repository.SetSubscribed(thread.ID, actor.ID, true); err != nil {
		return ThreadView{}, err
	}
	_ = service.auditEvent(actor.ID, "community/thread/"+thread.ID, "community.reply")
	if service.notifications != nil {
		path := "/community/thread/" + thread.ID
		if space.PostingPolicy == "owner" {
			path = "/lunastev/thread/" + thread.ID
		} else if space.ID == "showcase" {
			path = "/community/showcase/" + thread.ID
		}
		if subscribers, subscriberErr := service.repository.Subscribers(thread.ID); subscriberErr == nil {
			for _, accountID := range subscribers {
				_, _ = service.notifications.Notify(notificationdomain.Input{RecipientAccountID: accountID,
					ActorAccountID: actor.ID, ActorName: actor.DisplayName, Kind: "community.reply",
					Subject: root.Subject, URL: path})
			}
		}
	}
	return service.repository.ViewFor(thread.ID, actor.ID)
}

func (service *Service) Vote(actorID, threadID, targetType, targetID string, value int) (int64, error) {
	thread, err := service.repository.Thread(threadID)
	if err != nil {
		return 0, err
	}
	if targetType == "thread" {
		if targetID != thread.ID {
			return 0, fmt.Errorf("%w: vote target does not belong to the post", ErrInvalidPost)
		}
	} else if targetType == "message" {
		message, err := service.mail.Message(targetID)
		if err != nil || message.ThreadID != thread.ID || message.ID == thread.RootMessageID {
			return 0, fmt.Errorf("%w: vote target does not belong to the post", ErrInvalidPost)
		}
	} else {
		return 0, fmt.Errorf("%w: invalid vote target", ErrInvalidPost)
	}
	return service.repository.SetVote(targetType, targetID, actorID, value)
}

func (service *Service) Subscribe(actorID, threadID string, subscribed bool) error {
	if _, err := service.repository.Thread(threadID); err != nil {
		return err
	}
	return service.repository.SetSubscribed(threadID, actorID, subscribed)
}

func (service *Service) auditEvent(actorID, resourceID, action string) error {
	id, err := identifier.New("audit")
	if err != nil {
		return err
	}
	return service.audit.Append(audit.Event{ID: id, ActorID: "account/" + actorID, ResourceID: resourceID,
		Action: action, Result: "success", OccurredAt: service.now().UTC()})
}

func addressOf(actor account.Account) string {
	return (&stdmail.Address{Name: actor.DisplayName, Address: actor.Email}).String()
}

func normalizeBody(value string) string {
	return strings.TrimSpace(strings.ReplaceAll(value, "\r\n", "\n"))
}

func normalizeTags(values []string, englishOnly bool) ([]string, error) {
	if len(values) > 5 {
		return nil, fmt.Errorf("%w: at most five tags are allowed", ErrInvalidPost)
	}
	seen := make(map[string]bool)
	result := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.ToLower(strings.TrimSpace(value))
		if value == "" {
			continue
		}
		pattern := tagPattern
		if englishOnly {
			pattern = englishTagPattern
		}
		if !pattern.MatchString(value) {
			return nil, fmt.Errorf("%w: tags may contain letters, numbers, and hyphens", ErrInvalidPost)
		}
		if !seen[value] {
			seen[value] = true
			result = append(result, value)
		}
	}
	return result, nil
}

func englishProse(value string) bool {
	for _, character := range value {
		if unicode.IsLetter(character) && !unicode.In(character, unicode.Latin) {
			return false
		}
	}
	return true
}

func withoutMarkdownCode(value string) string {
	var result strings.Builder
	inFence := false
	for _, line := range strings.Split(value, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "```") {
			inFence = !inFence
			continue
		}
		if inFence {
			continue
		}
		inInlineCode := false
		for _, character := range line {
			if character == '`' {
				inInlineCode = !inInlineCode
				continue
			}
			if !inInlineCode {
				result.WriteRune(character)
			}
		}
		result.WriteByte('\n')
	}
	return result.String()
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
	headers.Set("Content-Type", "text/markdown; charset=utf-8")
	headers.Set("Content-Transfer-Encoding", "8bit")
	order := []string{"From", "To", "Subject", "Date", "Message-ID", "In-Reply-To", "References", "MIME-Version", "Content-Type", "Content-Transfer-Encoding"}
	var raw strings.Builder
	for _, name := range order {
		if value := headers.Get(name); value != "" {
			raw.WriteString(name)
			raw.WriteString(": ")
			raw.WriteString(value)
			raw.WriteString("\r\n")
		}
	}
	raw.WriteString("\r\n")
	raw.WriteString(strings.ReplaceAll(body, "\n", "\r\n"))
	return []byte(raw.String())
}
