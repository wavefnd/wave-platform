package release

import "encoding/xml"

const XMLNamespace = "https://wave-lang.dev/ns/platform/release/v1"

type Release struct {
	XMLName xml.Name `xml:"https://wave-lang.dev/ns/platform/release/v1 release"`

	ID          string `xml:"id,attr"`
	Slug        string `xml:"slug"`
	Title       string `xml:"title"`
	PublishedAt string `xml:"published-at"`
	Summary     string `xml:"summary"`
	MessageID   string `xml:"message-id"`
	Content     string `xml:"content,omitempty"`
	Source      string `xml:"migration-source"`
}

type Summary struct {
	Slug        string `xml:"slug"`
	Title       string `xml:"title"`
	PublishedAt string `xml:"published-at"`
	Summary     string `xml:"summary"`
}
