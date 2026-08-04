package mail

import (
	"encoding/xml"
	"time"
)

type Message struct {
	XMLName xml.Name `xml:"https://wave-lang.dev/ns/platform/mail/v1 message"`

	ID              string    `xml:"id,attr"`
	MessageID       string    `xml:"message-id"`
	ThreadID        string    `xml:"thread-id"`
	ParentMessageID string    `xml:"parent-message-id,omitempty"`
	AuthorAccountID string    `xml:"author-account-id,omitempty"`
	From            string    `xml:"from"`
	To              []string  `xml:"to"`
	Cc              []string  `xml:"cc,omitempty"`
	Subject         string    `xml:"subject"`
	RawMessageKey   string    `xml:"raw-message-key"`
	ReceivedAt      time.Time `xml:"received-at"`
	CreatedAt       time.Time `xml:"created-at"`
}

type Delivery struct {
	XMLName xml.Name `xml:"https://wave-lang.dev/ns/platform/mail/v1 delivery"`

	ID            string    `xml:"id,attr"`
	MessageID     string    `xml:"message-id"`
	Sender        string    `xml:"sender"`
	Recipient     string    `xml:"recipient"`
	Destination   string    `xml:"destination"`
	Status        string    `xml:"status"`
	Attempts      int       `xml:"attempts"`
	NextAttemptAt time.Time `xml:"next-attempt-at,omitempty"`
	LastAttemptAt time.Time `xml:"last-attempt-at,omitempty"`
	LastError     string    `xml:"last-error,omitempty"`
	CreatedAt     time.Time `xml:"created-at"`
	CompletedAt   time.Time `xml:"completed-at,omitempty"`
}
