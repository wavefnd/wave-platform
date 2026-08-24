package document

import (
	"encoding/xml"
	"time"
)

const XMLNamespace = "https://wave-lang.dev/ns/platform/document/v1"

type Document struct {
	XMLName xml.Name `xml:"https://wave-lang.dev/ns/platform/document/v1 document"`

	ID                  string    `xml:"id,attr"`
	TranslationSetID    string    `xml:"translation-set-id"`
	Path                string    `xml:"path"`
	Locale              string    `xml:"locale"`
	Group               string    `xml:"group"`
	GroupOrder          int       `xml:"group-order"`
	Order               int       `xml:"order"`
	Title               string    `xml:"title"`
	Summary             string    `xml:"summary"`
	Status              string    `xml:"status"`
	DraftRevisionID     string    `xml:"draft-revision-id,omitempty"`
	PublishedRevisionID string    `xml:"published-revision-id,omitempty"`
	CreatedAt           time.Time `xml:"created-at"`
	UpdatedAt           time.Time `xml:"updated-at"`
}

type Revision struct {
	XMLName xml.Name `xml:"https://wave-lang.dev/ns/platform/document/v1 revision"`

	ID          string    `xml:"id,attr"`
	DocumentID  string    `xml:"document-id"`
	AuthorID    string    `xml:"author-id"`
	ContentHash string    `xml:"content-hash"`
	ContentXML  []byte    `xml:"content-xml"`
	CreatedAt   time.Time `xml:"created-at"`
}

type Content struct {
	XMLName  xml.Name `xml:"https://wave-lang.dev/ns/platform/document/v1 content"`
	Markdown string   `xml:"markdown"`
	Blocks   []Block  `xml:"block"`
}

type Block struct {
	Kind     string     `xml:"kind,attr"`
	Anchor   string     `xml:"anchor,attr,omitempty"`
	Level    int        `xml:"level,attr,omitempty"`
	Language string     `xml:"language,attr,omitempty"`
	Title    string     `xml:"title,attr,omitempty"`
	Text     string     `xml:",chardata"`
	Items    []string   `xml:"item"`
	Rows     []TableRow `xml:"row"`
}

type TableRow struct {
	Header bool     `xml:"header,attr,omitempty"`
	Cells  []string `xml:"cell"`
}

type Summary struct {
	ID      string `xml:"id"`
	Path    string `xml:"path"`
	Locale  string `xml:"locale"`
	Group   string `xml:"group"`
	Order   int    `xml:"order"`
	Title   string `xml:"title"`
	Summary string `xml:"summary"`
}

type View struct {
	XMLName xml.Name `xml:"https://wave-lang.dev/ns/platform/api/v1 document"`
	Summary
	UpdatedAt string  `xml:"updated-at"`
	Markdown  string  `xml:"content>markdown"`
	Blocks    []Block `xml:"content>block"`
}
