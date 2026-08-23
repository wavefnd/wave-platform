package mailinglist

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	stdmail "net/mail"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/dgraph-io/badger/v4"
	maildomain "github.com/wavefnd/wave-platform/internal/mail"
	"github.com/wavefnd/wave-platform/internal/mailbox"
	"github.com/wavefnd/wave-platform/internal/storage"
)

var headerMessageIDPattern = regexp.MustCompile(`<[^<>\s]+>`)

type Repository struct {
	database  *storage.Database
	mail      *maildomain.Repository
	mailboxes *mailbox.Repository
}

const (
	maxPostsPerAccountPerListPerDay = 50
	maxMessagesPerList              = 5000
)

func NewRepository(database *storage.Database) *Repository {
	return &Repository{database: database, mail: maildomain.NewRepository(database), mailboxes: mailbox.NewRepository(database)}
}

func (repository *Repository) UpsertList(item List) error {
	if item.ID == "" || item.MailboxID == "" || item.Address == "" {
		return errors.New("mailing list id, mailbox id, and address are required")
	}
	if item.PostingPolicy != PostingMembers && item.PostingPolicy != PostingStaff {
		return errors.New("invalid mailing list posting policy")
	}
	if item.WebhookPolicy != WebhookDisabled && item.WebhookPolicy != WebhookSummary && item.WebhookPolicy != WebhookFull {
		return errors.New("invalid mailing list webhook policy")
	}
	if item.WebhookPreviewLimit < 0 || item.WebhookPreviewLimit > 2000 {
		return errors.New("invalid mailing list webhook preview limit")
	}
	data, err := xml.Marshal(item)
	if err != nil {
		return fmt.Errorf("encode mailing list: %w", err)
	}
	return repository.database.Set(storage.Key("mailing-list", "list", item.ID), data)
}

func (repository *Repository) List(id string) (List, error) {
	data, err := repository.database.Get(storage.Key("mailing-list", "list", strings.ToLower(strings.TrimSpace(id))))
	if err != nil {
		return List{}, err
	}
	var item List
	if err := xml.Unmarshal(data, &item); err != nil {
		return List{}, fmt.Errorf("decode mailing list: %w", err)
	}
	return item, nil
}

func (repository *Repository) Lists() ([]List, error) {
	items := make([]List, 0)
	err := repository.database.Scan(storage.Prefix("mailing-list", "list"), func(_, data []byte) error {
		var item List
		if err := xml.Unmarshal(data, &item); err != nil {
			return fmt.Errorf("decode mailing list: %w", err)
		}
		items = append(items, item)
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(items, func(left, right int) bool { return items[left].ID < items[right].ID })
	return items, nil
}

func (repository *Repository) SetSubscription(item Subscription, subscribed bool) error {
	if item.ListID == "" || item.AccountID == "" {
		return errors.New("mailing list and account ids are required")
	}
	key := storage.Key("mailing-list", "subscription", item.AccountID, item.ListID)
	if !subscribed {
		return repository.database.Delete(key)
	}
	data, err := xml.Marshal(item)
	if err != nil {
		return fmt.Errorf("encode mailing list subscription: %w", err)
	}
	return repository.database.Set(key, data)
}

func (repository *Repository) Subscribed(listID, accountID string) (bool, error) {
	_, err := repository.database.Get(storage.Key("mailing-list", "subscription", accountID, listID))
	if errors.Is(err, storage.ErrNotFound) {
		return false, nil
	}
	return err == nil, err
}

func (repository *Repository) Subscriptions(accountID string) ([]Subscription, error) {
	items := make([]Subscription, 0)
	err := repository.database.Scan(storage.Prefix("mailing-list", "subscription", accountID), func(_, data []byte) error {
		var item Subscription
		if err := xml.Unmarshal(data, &item); err != nil {
			return fmt.Errorf("decode mailing list subscription: %w", err)
		}
		items = append(items, item)
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(items, func(left, right int) bool { return items[left].ListID < items[right].ListID })
	return items, nil
}

func (repository *Repository) AddEntry(list List, entry mailbox.Entry) error {
	if entry.MailboxID != list.MailboxID {
		return errors.New("mailing list entry belongs to another mailbox")
	}
	return repository.mailboxes.AddEntry(entry)
}

// ReservePosting enforces bounded list growth before any message data is
// written. The reservation and both counters are checked in one Badger
// transaction so concurrent requests from the same account cannot race past
// the daily limit.
func (repository *Repository) ReservePosting(list List, accountID, messageID string, at time.Time) error {
	if list.ID == "" || list.MailboxID == "" || accountID == "" || messageID == "" {
		return errors.New("mailing list posting reservation is incomplete")
	}
	day := at.UTC().Format("2006-01-02")
	postingPrefix := storage.Prefix("mailing-list", "posting", day, list.ID, accountID)
	entryPrefix := storage.Prefix("mailbox", "entry", list.MailboxID)
	return repository.database.DB.Update(func(transaction *badger.Txn) error {
		options := badger.DefaultIteratorOptions
		options.PrefetchValues = false
		iterator := transaction.NewIterator(options)
		defer iterator.Close()
		postingCount := 0
		for iterator.Seek(postingPrefix); iterator.ValidForPrefix(postingPrefix); iterator.Next() {
			postingCount++
			if postingCount >= maxPostsPerAccountPerListPerDay {
				return ErrRateLimited
			}
		}
		entryCount := 0
		for iterator.Seek(entryPrefix); iterator.ValidForPrefix(entryPrefix); iterator.Next() {
			entryCount++
			if entryCount >= maxMessagesPerList {
				return ErrListFull
			}
		}
		return transaction.Set(storage.Key("mailing-list", "posting", day, list.ID, accountID, messageID), []byte(at.UTC().Format(time.RFC3339Nano)))
	})
}

func (repository *Repository) Threads(listID, query string, limit, offset int) ([]ThreadSummary, error) {
	views, err := repository.threadViews(listID)
	if err != nil {
		return nil, err
	}
	query = strings.ToLower(strings.TrimSpace(query))
	items := make([]ThreadSummary, 0, len(views))
	for _, view := range views {
		if len(view.Messages) == 0 {
			continue
		}
		root := view.Messages[0]
		if query != "" {
			var haystack strings.Builder
			for _, message := range view.Messages {
				haystack.WriteString(" " + message.Subject + " " + message.From + " " + message.Body)
			}
			if !strings.Contains(strings.ToLower(haystack.String()), query) {
				continue
			}
		}
		items = append(items, ThreadSummary{ID: view.ID, ListID: listID, RootMessageID: root.MessageID,
			Subject: view.Subject, Preview: preview(root.Body, 220), Author: root.From,
			AuthorAccountID: root.AuthorAccountID, MessageCount: len(view.Messages), CreatedAt: root.CreatedAt,
			LastActivityAt: view.Messages[len(view.Messages)-1].CreatedAt})
	}
	sort.SliceStable(items, func(left, right int) bool {
		if items[left].LastActivityAt.Equal(items[right].LastActivityAt) {
			return items[left].ID < items[right].ID
		}
		return items[left].LastActivityAt.After(items[right].LastActivityAt)
	})
	if offset >= len(items) {
		return []ThreadSummary{}, nil
	}
	if offset > 0 {
		items = items[offset:]
	}
	if limit > 0 && len(items) > limit {
		items = items[:limit]
	}
	return items, nil
}

func (repository *Repository) Thread(listID, threadID string) (ThreadView, error) {
	views, err := repository.threadViews(listID)
	if err != nil {
		return ThreadView{}, err
	}
	for _, view := range views {
		if view.ID == threadID {
			return view, nil
		}
	}
	return ThreadView{}, storage.ErrNotFound
}

type indexedMessage struct {
	entry          mailbox.Entry
	message        maildomain.Message
	body           string
	headerParentID string
}

func (repository *Repository) threadViews(listID string) ([]ThreadView, error) {
	list, err := repository.List(listID)
	if err != nil {
		return nil, err
	}
	entries, err := repository.mailboxes.Entries(list.MailboxID, "")
	if err != nil {
		return nil, err
	}
	indexed := make(map[string]indexedMessage, len(entries))
	byHeaderID := make(map[string]string, len(entries))
	for _, entry := range entries {
		message, err := repository.mailboxes.Message(entry)
		if err != nil {
			return nil, err
		}
		body, err := repository.mail.Body(message)
		if err != nil {
			return nil, err
		}
		parentHeader := ""
		if message.ParentMessageID == "" {
			parentHeader, _ = repository.headerParent(message)
		}
		indexed[message.ID] = indexedMessage{entry: entry, message: message, body: body, headerParentID: parentHeader}
		if normalized := normalizeHeaderMessageID(message.MessageID); normalized != "" {
			byHeaderID[normalized] = message.ID
		}
	}

	parentByID := make(map[string]string, len(indexed))
	for id, item := range indexed {
		parent := item.message.ParentMessageID
		if parent == "" && item.headerParentID != "" {
			parent = byHeaderID[normalizeHeaderMessageID(item.headerParentID)]
		}
		if _, exists := indexed[parent]; exists && parent != id {
			parentByID[id] = parent
		}
	}
	rootFor := func(id string) string {
		seen := map[string]bool{}
		current := id
		for current != "" && !seen[current] {
			seen[current] = true
			parent := parentByID[current]
			if parent == "" {
				item := indexed[current]
				if item.message.ThreadID != "" && item.message.ThreadID != current {
					if _, exists := indexed[item.message.ThreadID]; exists {
						current = item.message.ThreadID
						continue
					}
				}
				return current
			}
			current = parent
		}
		return id
	}

	groups := make(map[string][]MessageView)
	for id, item := range indexed {
		parent := parentByID[id]
		groups[rootFor(id)] = append(groups[rootFor(id)], MessageView{ID: item.message.ID, EntryID: item.entry.ID,
			MessageID: item.message.ID, HeaderMessageID: item.message.MessageID, ParentMessageID: parent,
			AuthorAccountID: item.message.AuthorAccountID, From: item.message.From, To: item.message.To,
			Subject: item.message.Subject, Body: item.body, CreatedAt: messageTime(item.message)})
	}
	views := make([]ThreadView, 0, len(groups))
	for rootID, messages := range groups {
		sort.SliceStable(messages, func(left, right int) bool {
			if messages[left].CreatedAt.Equal(messages[right].CreatedAt) {
				return messages[left].ID < messages[right].ID
			}
			return messages[left].CreatedAt.Before(messages[right].CreatedAt)
		})
		rootIndex := 0
		for index := range messages {
			if messages[index].ID == rootID {
				rootIndex = index
				break
			}
		}
		if rootIndex != 0 {
			root := messages[rootIndex]
			messages = append([]MessageView{root}, append(messages[:rootIndex], messages[rootIndex+1:]...)...)
		}
		views = append(views, ThreadView{ID: rootID, ListID: list.ID, Address: list.Address,
			Subject: messages[0].Subject, Messages: messages})
	}
	return views, nil
}

func (repository *Repository) headerParent(message maildomain.Message) (string, error) {
	raw, err := repository.mail.RawMessage(message)
	if err != nil {
		return "", err
	}
	parsed, err := stdmail.ReadMessage(bytes.NewReader(raw))
	if err != nil {
		return "", err
	}
	if value := lastHeaderMessageID(parsed.Header.Get("In-Reply-To")); value != "" {
		return value, nil
	}
	return lastHeaderMessageID(parsed.Header.Get("References")), nil
}

func lastHeaderMessageID(value string) string {
	matches := headerMessageIDPattern.FindAllString(value, -1)
	if len(matches) == 0 {
		return ""
	}
	return matches[len(matches)-1]
}

func normalizeHeaderMessageID(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func messageTime(message maildomain.Message) time.Time {
	if !message.CreatedAt.IsZero() {
		return message.CreatedAt
	}
	return message.ReceivedAt
}

func preview(value string, limit int) string {
	characters := []rune(strings.Join(strings.Fields(value), " "))
	if len(characters) > limit {
		return string(characters[:limit]) + "…"
	}
	return string(characters)
}
