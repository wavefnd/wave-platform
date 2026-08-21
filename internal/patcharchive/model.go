package patcharchive

import (
	"encoding/xml"
	"time"
)

type Patch struct {
	XMLName            xml.Name        `xml:"https://wave-lang.dev/ns/platform/api/v1 patch"`
	ID                 string          `xml:"id,attr"`
	MessageID          string          `xml:"message-id"`
	Subject            string          `xml:"subject"`
	Title              string          `xml:"title"`
	AuthorName         string          `xml:"author-name"`
	AuthorEmail        string          `xml:"author-email"`
	Body               string          `xml:"body,omitempty"`
	Preview            string          `xml:"preview"`
	Version            int             `xml:"version,omitempty"`
	Part               int             `xml:"part,omitempty"`
	Total              int             `xml:"total,omitempty"`
	Files              []string        `xml:"files>file,omitempty"`
	SHA256             string          `xml:"sha256,omitempty"`
	ReviewStatus       string          `xml:"review-status"`
	TargetRepository   string          `xml:"target-repository,omitempty"`
	AssigneeAccountID  string          `xml:"assignee-account-id,omitempty"`
	AssigneeName       string          `xml:"assignee-name,omitempty"`
	ReviewUpdatedAt    time.Time       `xml:"review-updated-at,omitempty"`
	SeriesCount        int             `xml:"series-count,omitempty"`
	ReviewCommentCount int             `xml:"review-comment-count,omitempty"`
	ReviewComments     []ReviewComment `xml:"review-comments>comment,omitempty"`
	ReceivedAt         time.Time       `xml:"received-at"`
}

type Review struct {
	XMLName           xml.Name  `xml:"https://wave-lang.dev/ns/platform/patch-review/v1 review"`
	PatchID           string    `xml:"patch-id"`
	Status            string    `xml:"status"`
	TargetRepository  string    `xml:"target-repository,omitempty"`
	AssigneeAccountID string    `xml:"assignee-account-id,omitempty"`
	UpdatedBy         string    `xml:"updated-by"`
	UpdatedAt         time.Time `xml:"updated-at"`
}

type ReviewInput struct {
	Status           string `xml:"status"`
	TargetRepository string `xml:"target-repository"`
}

type ReviewComment struct {
	XMLName         xml.Name  `xml:"https://wave-lang.dev/ns/platform/patch-review/v1 comment"`
	ID              string    `xml:"id,attr"`
	PatchID         string    `xml:"patch-id,omitempty"`
	AuthorAccountID string    `xml:"author-account-id"`
	AuthorName      string    `xml:"author-name,omitempty"`
	Path            string    `xml:"path,omitempty"`
	Line            int       `xml:"line,omitempty"`
	LineText        string    `xml:"line-text,omitempty"`
	Body            string    `xml:"body"`
	Resolved        bool      `xml:"resolved"`
	CreatedAt       time.Time `xml:"created-at"`
	UpdatedAt       time.Time `xml:"updated-at"`
}

type ReviewCommentInput struct {
	Line int    `xml:"line"`
	Body string `xml:"body"`
}

type ReviewCommentResolutionInput struct {
	Resolved bool `xml:"resolved"`
}
