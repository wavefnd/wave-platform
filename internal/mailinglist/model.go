package mailinglist

import (
	"encoding/xml"
	"time"
)

const XMLNamespace = "https://wave-lang.dev/ns/platform/mailing-list/v1"

const (
	PostingMembers = "members"
	PostingStaff   = "staff"

	WebhookDisabled = "disabled"
	WebhookSummary  = "summary"
	WebhookFull     = "full"
)

// List describes an internal Wave mailing list. MailboxID points at the
// service mailbox whose mailbox.Entry values reference the canonical
// mail.Message objects. Message bodies are never duplicated here.
type List struct {
	XMLName xml.Name `xml:"https://wave-lang.dev/ns/platform/mailing-list/v1 list"`

	ID                  string    `xml:"id,attr"`
	MailboxID           string    `xml:"mailbox-id"`
	Address             string    `xml:"address"`
	Name                string    `xml:"name"`
	Description         string    `xml:"description"`
	PostingPolicy       string    `xml:"posting-policy"`
	WebhookPolicy       string    `xml:"webhook-policy"`
	WebhookPreviewLimit int       `xml:"webhook-preview-limit,omitempty"`
	CreatedAt           time.Time `xml:"created-at"`
}

// Subscription intentionally stores an internal account identifier rather
// than an email address. This prevents an Internet address from becoming a
// list delivery target.
type Subscription struct {
	XMLName xml.Name `xml:"https://wave-lang.dev/ns/platform/mailing-list/v1 subscription"`

	ListID    string    `xml:"list-id"`
	AccountID string    `xml:"account-id"`
	CreatedAt time.Time `xml:"created-at"`
}

type ListSummary struct {
	ID                  string `xml:"id,attr"`
	Address             string `xml:"address"`
	Name                string `xml:"name"`
	Description         string `xml:"description"`
	PostingPolicy       string `xml:"posting-policy"`
	WebhookPolicy       string `xml:"webhook-policy"`
	WebhookPreviewLimit int    `xml:"webhook-preview-limit,omitempty"`
	Subscribed          bool   `xml:"subscribed"`
}

type ThreadSummary struct {
	ID              string    `xml:"id,attr"`
	ListID          string    `xml:"list-id"`
	RootMessageID   string    `xml:"root-message-id"`
	Subject         string    `xml:"subject"`
	Preview         string    `xml:"preview"`
	Author          string    `xml:"author"`
	AuthorAccountID string    `xml:"author-account-id,omitempty"`
	MessageCount    int       `xml:"message-count"`
	CreatedAt       time.Time `xml:"created-at"`
	LastActivityAt  time.Time `xml:"last-activity-at"`
}

// MessageView exposes identifiers for both canonical records used by a list:
// MessageID is mail.Message.ID and EntryID is mailbox.Entry.ID.
type MessageView struct {
	ID              string    `xml:"id,attr"`
	EntryID         string    `xml:"entry-id"`
	MessageID       string    `xml:"message-id"`
	HeaderMessageID string    `xml:"header-message-id"`
	ParentMessageID string    `xml:"parent-message-id,omitempty"`
	AuthorAccountID string    `xml:"author-account-id,omitempty"`
	From            string    `xml:"from"`
	To              []string  `xml:"to"`
	Subject         string    `xml:"subject"`
	Body            string    `xml:"body"`
	CreatedAt       time.Time `xml:"created-at"`
}

type ThreadView struct {
	XMLName xml.Name `xml:"https://wave-lang.dev/ns/platform/mailing-list/v1 thread"`

	ID       string        `xml:"id,attr"`
	ListID   string        `xml:"list-id"`
	Address  string        `xml:"address"`
	Subject  string        `xml:"subject"`
	Messages []MessageView `xml:"messages>message"`
}

type PostInput struct {
	Subject string
	Body    string
}

type ReplyInput struct {
	ParentMessageID string
	Body            string
}
