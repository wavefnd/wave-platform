package rfc

import (
	"encoding/xml"
	"time"
)

const XMLNamespace = "https://wave-lang.dev/ns/platform/rfc/v1"

type Proposal struct {
	XMLName         xml.Name  `xml:"https://wave-lang.dev/ns/platform/rfc/v1 proposal"`
	Number          uint64    `xml:"number,attr"`
	Title           string    `xml:"title"`
	Summary         string    `xml:"summary"`
	Content         string    `xml:"content,omitempty"`
	Status          string    `xml:"status"`
	AuthorAccountID string    `xml:"author-account-id"`
	AuthorName      string    `xml:"author-name"`
	CommentCount    int       `xml:"comment-count"`
	Comments        []Comment `xml:"comments>comment,omitempty"`
	CreatedAt       time.Time `xml:"created-at"`
	UpdatedAt       time.Time `xml:"updated-at"`
}

type Comment struct {
	XMLName         xml.Name  `xml:"https://wave-lang.dev/ns/platform/rfc/v1 comment"`
	ID              string    `xml:"id,attr"`
	ProposalNumber  uint64    `xml:"proposal-number,omitempty"`
	AuthorAccountID string    `xml:"author-account-id"`
	AuthorName      string    `xml:"author-name"`
	Body            string    `xml:"body"`
	CreatedAt       time.Time `xml:"created-at"`
}

type ProposalInput struct {
	Title   string `xml:"title"`
	Content string `xml:"content"`
}

type StatusInput struct {
	Status string `xml:"status"`
}

type CommentInput struct {
	Body string `xml:"body"`
}
