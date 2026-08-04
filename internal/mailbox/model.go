package mailbox

import (
	"encoding/xml"
	"time"
)

type Mailbox struct {
	XMLName   xml.Name  `xml:"https://wave-lang.dev/ns/platform/mailbox/v1 mailbox"`
	ID        string    `xml:"id,attr"`
	AccountID string    `xml:"account-id"`
	Address   string    `xml:"address"`
	Quota     int64     `xml:"quota"`
	CreatedAt time.Time `xml:"created-at"`
}

type Entry struct {
	XMLName   xml.Name  `xml:"https://wave-lang.dev/ns/platform/mailbox/v1 entry"`
	ID        string    `xml:"id,attr"`
	MailboxID string    `xml:"mailbox-id"`
	MessageID string    `xml:"message-id"`
	Folder    string    `xml:"folder"`
	Flags     []string  `xml:"flags>flag,omitempty"`
	CreatedAt time.Time `xml:"created-at"`
}
