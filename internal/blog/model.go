package blog

import (
	"encoding/xml"
	"time"
)

const XMLNamespace = "https://wave-lang.dev/ns/platform/blog/v1"

type Post struct {
	XMLName xml.Name `xml:"https://wave-lang.dev/ns/platform/blog/v1 post"`

	Slug            string    `xml:"slug"`
	Locale          string    `xml:"locale"`
	Title           string    `xml:"title"`
	Summary         string    `xml:"summary"`
	Content         string    `xml:"content"`
	Status          string    `xml:"status"`
	AuthorAccountID string    `xml:"author-account-id"`
	AuthorName      string    `xml:"author-name"`
	PublishedAt     string    `xml:"published-at,omitempty"`
	CreatedAt       time.Time `xml:"created-at"`
	UpdatedAt       time.Time `xml:"updated-at"`
}

type Summary struct {
	Slug        string `xml:"slug"`
	Locale      string `xml:"locale"`
	Title       string `xml:"title"`
	Summary     string `xml:"summary"`
	Status      string `xml:"status,omitempty"`
	AuthorName  string `xml:"author-name"`
	PublishedAt string `xml:"published-at,omitempty"`
	UpdatedAt   string `xml:"updated-at"`
}

type Input struct {
	Slug    string `xml:"slug"`
	Locale  string `xml:"locale"`
	Title   string `xml:"title"`
	Summary string `xml:"summary"`
	Content string `xml:"content"`
	Status  string `xml:"status"`
}
