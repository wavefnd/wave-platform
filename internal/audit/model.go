package audit

import (
	"encoding/xml"
	"time"
)

type Event struct {
	XMLName    xml.Name  `xml:"https://wave-lang.dev/ns/platform/audit/v1 event"`
	Sequence   uint64    `xml:"sequence"`
	ID         string    `xml:"id,attr"`
	ActorID    string    `xml:"actor-id"`
	ResourceID string    `xml:"resource-id"`
	Action     string    `xml:"action"`
	Result     string    `xml:"result"`
	RequestID  string    `xml:"request-id,omitempty"`
	OccurredAt time.Time `xml:"occurred-at"`
}
