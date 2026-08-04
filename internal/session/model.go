package session

import (
	"encoding/xml"
	"time"
)

type Session struct {
	XMLName    xml.Name  `xml:"https://wave-lang.dev/ns/platform/session/v1 session"`
	ID         string    `xml:"id,attr"`
	AccountID  string    `xml:"account-id"`
	CreatedAt  time.Time `xml:"created-at"`
	ExpiresAt  time.Time `xml:"expires-at"`
	LastSeenAt time.Time `xml:"last-seen-at"`
	UserAgent  string    `xml:"user-agent,omitempty"`
}
