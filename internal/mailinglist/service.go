package mailinglist

import (
	"errors"
	"fmt"
	"mime"
	stdmail "net/mail"
	"net/textproto"
	"net/url"
	"strings"
	"time"

	"github.com/wavefnd/wave-platform/internal/account"
	"github.com/wavefnd/wave-platform/internal/audit"
	"github.com/wavefnd/wave-platform/internal/identifier"
	maildomain "github.com/wavefnd/wave-platform/internal/mail"
	"github.com/wavefnd/wave-platform/internal/mailbox"
	patchdomain "github.com/wavefnd/wave-platform/internal/patcharchive"
	"github.com/wavefnd/wave-platform/internal/permission"
	"github.com/wavefnd/wave-platform/internal/storage"
	webhookdomain "github.com/wavefnd/wave-platform/internal/webhook"
)

const (
	announceMailboxAccountID    = "service/mailing-list/announce"
	developmentMailboxAccountID = "service/mailing-list/development"
	maxPostBodyRunes            = 20000
)

var (
	ErrForbidden     = errors.New("mailing list access is forbidden")
	ErrNotSubscribed = errors.New("mailing list subscription is required")
	ErrInvalidPost   = errors.New("invalid mailing list post")
	ErrRateLimited   = errors.New("mailing list posting rate limit reached")
	ErrListFull      = errors.New("mailing list storage quota reached")
)

type webhookPublisher interface {
	Publish(webhookdomain.Event) error
}

type Service struct {
	repository            *Repository
	accounts              *account.Repository
	mail                  *maildomain.Repository
	mailboxes             *mailbox.Repository
	permissions           *permission.Repository
	audit                 *audit.Repository
	mailDomain            string
	patchMailboxAccountID string
	webhooks              webhookPublisher
	now                   func() time.Time
}

func NewService(database *storage.Database, mailDomain, patchMailboxAccountID string) *Service {
	return &Service{repository: NewRepository(database), accounts: account.NewRepository(database),
		mail: maildomain.NewRepository(database), mailboxes: mailbox.NewRepository(database),
		permissions: permission.NewRepository(database), audit: audit.NewRepository(database),
		mailDomain: strings.ToLower(strings.TrimSpace(mailDomain)), patchMailboxAccountID: patchMailboxAccountID,
		now: time.Now}
}

func (service *Service) Repository() *Repository { return service.repository }

func (service *Service) SetWebhookService(webhooks *webhookdomain.Service) {
	service.webhooks = webhooks
}

func (service *Service) EnsureDefaults() error {
	if service.mailDomain == "" || strings.ContainsAny(service.mailDomain, "@/\\\r\n") {
		return errors.New("valid mailing list mail domain is required")
	}
	definitions := []struct {
		id, accountID, name, description, posting, webhook string
		preview                                            int
	}{
		{"announce", announceMailboxAccountID, "Announcements", "Official Wave project and release announcements.", PostingStaff, WebhookFull, 500},
		{"development", developmentMailboxAccountID, "Development", "Wave language, compiler, and platform development discussion.", PostingMembers, WebhookSummary, 180},
		{"patchs", service.patchMailboxAccountID, "Git patches", "Git patch submission and code review discussion.", PostingMembers, WebhookSummary, 180},
	}
	for _, definition := range definitions {
		if strings.TrimSpace(definition.accountID) == "" {
			return errors.New("mailing list service mailbox account id is required")
		}
		address := definition.id + "@" + service.mailDomain
		box, err := service.ensureMailbox(definition.accountID, address)
		if err != nil {
			return fmt.Errorf("ensure %s mailing list mailbox: %w", definition.id, err)
		}
		if _, err := service.repository.List(definition.id); err == nil {
			continue
		} else if !errors.Is(err, storage.ErrNotFound) {
			return err
		}
		item := List{ID: definition.id, MailboxID: box.ID, Address: address, Name: definition.name,
			Description: definition.description, PostingPolicy: definition.posting, WebhookPolicy: definition.webhook,
			WebhookPreviewLimit: definition.preview, CreatedAt: service.now().UTC()}
		if err := service.repository.UpsertList(item); err != nil {
			return err
		}
	}
	return nil
}

func (service *Service) Lists(actorID string) ([]ListSummary, error) {
	if _, err := service.internalAccount(actorID); err != nil {
		return nil, err
	}
	lists, err := service.repository.Lists()
	if err != nil {
		return nil, err
	}
	items := make([]ListSummary, 0, len(lists))
	for _, list := range lists {
		subscribed, err := service.repository.Subscribed(list.ID, actorID)
		if err != nil {
			return nil, err
		}
		items = append(items, listSummary(list, subscribed))
	}
	return items, nil
}

func (service *Service) Subscriptions(actorID string) ([]ListSummary, error) {
	if _, err := service.internalAccount(actorID); err != nil {
		return nil, err
	}
	subscriptions, err := service.repository.Subscriptions(actorID)
	if err != nil {
		return nil, err
	}
	items := make([]ListSummary, 0, len(subscriptions))
	for _, subscription := range subscriptions {
		list, err := service.repository.List(subscription.ListID)
		if errors.Is(err, storage.ErrNotFound) {
			continue
		}
		if err != nil {
			return nil, err
		}
		items = append(items, listSummary(list, true))
	}
	return items, nil
}

func (service *Service) Subscribe(actorID, listID string, subscribed bool) error {
	if _, err := service.internalAccount(actorID); err != nil {
		return err
	}
	list, err := service.repository.List(listID)
	if err != nil {
		return err
	}
	if err := service.repository.SetSubscription(Subscription{ListID: list.ID, AccountID: actorID,
		CreatedAt: service.now().UTC()}, subscribed); err != nil {
		return err
	}
	action := "mailing-list.subscribe"
	if !subscribed {
		action = "mailing-list.unsubscribe"
	}
	// The subscription is already committed; an audit write failure must not
	// turn a successful mutation into a retry that reverses or duplicates it.
	_ = service.appendAudit(actorID, "mailing-list/"+list.ID+"/subscription/"+actorID, action)
	return nil
}

func (service *Service) Threads(actorID, listID, query string, limit, offset int) ([]ThreadSummary, error) {
	if _, err := service.internalAccount(actorID); err != nil {
		return nil, err
	}
	if _, err := service.repository.List(listID); err != nil {
		return nil, err
	}
	return service.repository.Threads(listID, query, limit, offset)
}

func (service *Service) Thread(actorID, listID, threadID string) (ThreadView, error) {
	if _, err := service.internalAccount(actorID); err != nil {
		return ThreadView{}, err
	}
	return service.repository.Thread(listID, threadID)
}

func (service *Service) Post(actorID, listID string, input PostInput) (ThreadView, error) {
	actor, list, err := service.postingActor(actorID, listID)
	if err != nil {
		return ThreadView{}, err
	}
	subject, body, err := validatePost(input.Subject, input.Body)
	if err != nil {
		return ThreadView{}, err
	}
	if list.ID == "patchs" && !patchdomain.Valid(subject, body) {
		return ThreadView{}, fmt.Errorf("%w: new patch threads must contain a valid Git patch", ErrInvalidPost)
	}
	messageID, err := identifier.New("message")
	if err != nil {
		return ThreadView{}, err
	}
	now := service.now().UTC()
	if err := service.repository.ReservePosting(list, actor.ID, messageID, now); err != nil {
		return ThreadView{}, err
	}
	message := maildomain.Message{ID: messageID, MessageID: "<" + messageID + "@" + service.mailDomain + ">",
		ThreadID: messageID, AuthorAccountID: actor.ID, From: addressOf(actor), To: []string{list.Address},
		Subject: subject, ReceivedAt: now, CreatedAt: now}
	if err := service.mail.UpsertMessage(message, rawMessage(message, body, "", "")); err != nil {
		return ThreadView{}, err
	}
	if _, err := service.addListEntry(list, message.ID, now); err != nil {
		return ThreadView{}, err
	}
	_ = service.appendAudit(actor.ID, "mailing-list/"+list.ID+"/thread/"+message.ID, "mailing-list.post")
	service.publish(list, actor, message.ID, subject, body, true)
	return service.repository.Thread(list.ID, message.ID)
}

func (service *Service) Reply(actorID, listID, threadID string, input ReplyInput) (ThreadView, error) {
	actor, list, err := service.postingActor(actorID, listID)
	if err != nil {
		return ThreadView{}, err
	}
	thread, err := service.repository.Thread(list.ID, threadID)
	if err != nil {
		return ThreadView{}, err
	}
	_, body, err := validatePost(thread.Subject, input.Body)
	if err != nil {
		return ThreadView{}, err
	}
	parent := thread.Messages[len(thread.Messages)-1]
	if requested := strings.TrimSpace(input.ParentMessageID); requested != "" {
		found := false
		for _, message := range thread.Messages {
			if message.MessageID == requested {
				parent, found = message, true
				break
			}
		}
		if !found {
			return ThreadView{}, fmt.Errorf("%w: reply parent does not belong to this thread", ErrInvalidPost)
		}
	}
	messageID, err := identifier.New("message")
	if err != nil {
		return ThreadView{}, err
	}
	now := service.now().UTC()
	if err := service.repository.ReservePosting(list, actor.ID, messageID, now); err != nil {
		return ThreadView{}, err
	}
	subject := replySubject(thread.Subject)
	message := maildomain.Message{ID: messageID, MessageID: "<" + messageID + "@" + service.mailDomain + ">",
		ThreadID: thread.ID, ParentMessageID: parent.MessageID, AuthorAccountID: actor.ID, From: addressOf(actor),
		To: []string{list.Address}, Subject: subject, ReceivedAt: now, CreatedAt: now}
	if err := service.mail.UpsertMessage(message, rawMessage(message, body, parent.HeaderMessageID, threadReferences(thread, parent))); err != nil {
		return ThreadView{}, err
	}
	if _, err := service.addListEntry(list, message.ID, now); err != nil {
		return ThreadView{}, err
	}
	_ = service.appendAudit(actor.ID, "mailing-list/"+list.ID+"/thread/"+thread.ID+"/message/"+message.ID, "mailing-list.reply")
	service.publish(list, actor, thread.ID, thread.Subject, body, false)
	return service.repository.Thread(list.ID, thread.ID)
}

func (service *Service) publish(list List, actor account.Account, threadID, title, body string, root bool) {
	if service.webhooks == nil || list.WebhookPolicy == WebhookDisabled {
		return
	}
	summary := body
	if list.WebhookPolicy == WebhookSummary {
		summary = truncateRunes(body, list.WebhookPreviewLimit)
	}
	eventType := webhookdomain.EventMailingListPost
	resourceID := "mailing-list/" + list.ID + "/thread/" + threadID
	resourceURL := "/mail/lists/" + url.PathEscape(list.ID) + "/thread/" + url.PathEscape(threadID)
	if root && list.ID == "patchs" {
		eventType = webhookdomain.EventPatchReceived
		resourceID = "patch/" + threadID
		resourceURL = "/mail/lists/patchs/patch/" + url.PathEscape(threadID)
	}
	_ = service.webhooks.Publish(webhookdomain.Event{
		Type:       eventType,
		Title:      title,
		Summary:    summary,
		AuthorName: actor.DisplayName,
		ResourceID: resourceID,
		URL:        resourceURL,
	})
}

func (service *Service) postingActor(actorID, listID string) (account.Account, List, error) {
	actor, err := service.internalAccount(actorID)
	if err != nil {
		return account.Account{}, List{}, err
	}
	list, err := service.repository.List(listID)
	if err != nil {
		return account.Account{}, List{}, err
	}
	subscribed, err := service.repository.Subscribed(list.ID, actor.ID)
	if err != nil {
		return account.Account{}, List{}, err
	}
	if !subscribed {
		return account.Account{}, List{}, ErrNotSubscribed
	}
	if list.PostingPolicy == PostingStaff {
		owner, err := service.permissions.HasRole(actor.ID, "platform-owner")
		if err != nil {
			return account.Account{}, List{}, err
		}
		administrator, err := service.permissions.HasRole(actor.ID, "platform-admin")
		if err != nil {
			return account.Account{}, List{}, err
		}
		if !owner && !administrator {
			return account.Account{}, List{}, ErrForbidden
		}
	}
	return actor, list, nil
}

func (service *Service) internalAccount(accountID string) (account.Account, error) {
	item, err := service.accounts.Account(accountID)
	if err != nil {
		return account.Account{}, err
	}
	parsed, parseErr := stdmail.ParseAddress(strings.TrimSpace(item.Email))
	parts := []string(nil)
	if parseErr == nil {
		parts = strings.Split(strings.ToLower(parsed.Address), "@")
	}
	if item.Status != "active" || len(parts) != 2 || parts[0] == "" || parts[1] != service.mailDomain {
		return account.Account{}, ErrForbidden
	}
	return item, nil
}

func (service *Service) ensureMailbox(accountID, address string) (mailbox.Mailbox, error) {
	box, err := service.mailboxes.MailboxByAccount(accountID)
	if errors.Is(err, storage.ErrNotFound) {
		id, idErr := identifier.New("mailbox")
		if idErr != nil {
			return mailbox.Mailbox{}, idErr
		}
		box = mailbox.Mailbox{ID: id, AccountID: accountID, Address: address, CreatedAt: service.now().UTC()}
		if err := service.mailboxes.UpsertMailbox(box); err != nil {
			return mailbox.Mailbox{}, err
		}
	} else if err != nil {
		return mailbox.Mailbox{}, err
	}
	if !strings.EqualFold(box.Address, address) {
		return mailbox.Mailbox{}, errors.New("mailing list service mailbox has an unexpected address")
	}
	return box, nil
}

func (service *Service) addListEntry(list List, messageID string, now time.Time) (mailbox.Entry, error) {
	entryID, err := identifier.New("mailbox-entry")
	if err != nil {
		return mailbox.Entry{}, err
	}
	entry := mailbox.Entry{ID: entryID, MailboxID: list.MailboxID, MessageID: messageID, Folder: "Inbox", CreatedAt: now}
	return entry, service.repository.AddEntry(list, entry)
}

func (service *Service) appendAudit(actorID, resourceID, action string) error {
	id, err := identifier.New("audit")
	if err != nil {
		return err
	}
	return service.audit.Append(audit.Event{ID: id, ActorID: "account/" + actorID, ResourceID: resourceID,
		Action: action, Result: "success", OccurredAt: service.now().UTC()})
}

func listSummary(list List, subscribed bool) ListSummary {
	return ListSummary{ID: list.ID, Address: list.Address, Name: list.Name, Description: list.Description,
		PostingPolicy: list.PostingPolicy, WebhookPolicy: list.WebhookPolicy,
		WebhookPreviewLimit: list.WebhookPreviewLimit, Subscribed: subscribed}
}

func validatePost(subject, body string) (string, string, error) {
	subject = strings.TrimSpace(subject)
	body = strings.TrimSpace(strings.ReplaceAll(body, "\r\n", "\n"))
	if len([]rune(subject)) < 1 || len([]rune(subject)) > 180 || strings.ContainsAny(subject, "\r\n") {
		return "", "", fmt.Errorf("%w: subject must contain between 1 and 180 characters", ErrInvalidPost)
	}
	if len([]rune(body)) < 1 || len([]rune(body)) > maxPostBodyRunes {
		return "", "", fmt.Errorf("%w: body must contain between 1 and %d characters", ErrInvalidPost, maxPostBodyRunes)
	}
	return subject, body, nil
}

func truncateRunes(value string, limit int) string {
	characters := []rune(strings.TrimSpace(value))
	if limit > 0 && len(characters) > limit {
		return strings.TrimSpace(string(characters[:limit])) + "…"
	}
	return string(characters)
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

func threadReferences(thread ThreadView, parent MessageView) string {
	byID := make(map[string]MessageView, len(thread.Messages))
	for _, message := range thread.Messages {
		byID[message.MessageID] = message
	}
	chain := make([]string, 0, len(thread.Messages))
	current := parent
	seen := map[string]bool{}
	for current.MessageID != "" && !seen[current.MessageID] {
		seen[current.MessageID] = true
		if current.HeaderMessageID != "" {
			chain = append(chain, current.HeaderMessageID)
		}
		if current.ParentMessageID == "" {
			break
		}
		current = byID[current.ParentMessageID]
	}
	for left, right := 0, len(chain)-1; left < right; left, right = left+1, right-1 {
		chain[left], chain[right] = chain[right], chain[left]
	}
	return strings.Join(chain, " ")
}

func rawMessage(message maildomain.Message, body, inReplyTo, references string) []byte {
	headers := textproto.MIMEHeader{}
	headers.Set("From", message.From)
	headers.Set("To", strings.Join(message.To, ", "))
	headers.Set("Subject", mime.QEncoding.Encode("utf-8", message.Subject))
	headers.Set("Date", message.CreatedAt.Format(time.RFC1123Z))
	headers.Set("Message-ID", message.MessageID)
	if inReplyTo != "" {
		headers.Set("In-Reply-To", inReplyTo)
	}
	if references != "" {
		headers.Set("References", references)
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
