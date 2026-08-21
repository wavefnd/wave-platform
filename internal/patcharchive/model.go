package patcharchive

import (
	"encoding/xml"
	"time"
)

type Patch struct {
	XMLName     xml.Name  `xml:"https://wave-lang.dev/ns/platform/api/v1 patch"`
	ID          string    `xml:"id,attr"`
	MessageID   string    `xml:"message-id"`
	Subject     string    `xml:"subject"`
	Title       string    `xml:"title"`
	AuthorName  string    `xml:"author-name"`
	AuthorEmail string    `xml:"author-email"`
	Body        string    `xml:"body,omitempty"`
	Preview     string    `xml:"preview"`
	Version     int       `xml:"version,omitempty"`
	Part        int       `xml:"part,omitempty"`
	Total       int       `xml:"total,omitempty"`
	Files       []string  `xml:"files>file,omitempty"`
	ReceivedAt  time.Time `xml:"received-at"`
}
