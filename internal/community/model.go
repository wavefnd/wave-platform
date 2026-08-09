package community

import (
	"encoding/xml"
	"time"
)

const XMLNamespace = "https://wave-lang.dev/ns/platform/community/v1"

// Space and Thread contain community metadata only. The actual post and reply
// content is always owned by mail.Message.
type Space struct {
	XMLName xml.Name `xml:"https://wave-lang.dev/ns/platform/community/v1 space"`

	ID         string `xml:"id,attr"`
	Slug       string `xml:"slug"`
	Name       string `xml:"name"`
	Visibility string `xml:"visibility"`
	// PostingPolicy is "members" for public boards and "owner" for the
	// owner's writing and development log. Replies are governed separately.
	PostingPolicy string `xml:"posting-policy"`
}

type Thread struct {
	XMLName xml.Name `xml:"https://wave-lang.dev/ns/platform/community/v1 thread"`

	ID            string   `xml:"id,attr"`
	SpaceID       string   `xml:"space-id"`
	RootMessageID string   `xml:"root-message-id"`
	Tags          []string `xml:"tags>tag,omitempty"`
	Pinned        bool     `xml:"pinned"`
	Locked        bool     `xml:"locked"`
}

type ThreadSummary struct {
	ID              string   `xml:"id"`
	SpaceID         string   `xml:"space-id"`
	Title           string   `xml:"title"`
	Author          string   `xml:"author"`
	AuthorAccountID string   `xml:"author-account-id,omitempty"`
	Excerpt         string   `xml:"excerpt"`
	CreatedAt       string   `xml:"created-at"`
	ReplyCount      int      `xml:"reply-count"`
	ViewCount       uint64   `xml:"view-count"`
	Score           int64    `xml:"score"`
	ViewerVote      int      `xml:"viewer-vote"`
	LastActivityAt  string   `xml:"last-activity-at"`
	Tags            []string `xml:"tags>tag,omitempty"`
	Pinned          bool     `xml:"pinned"`
	Locked          bool     `xml:"locked"`
}

type MessageView struct {
	ID              string `xml:"id"`
	ParentMessageID string `xml:"parent-message-id,omitempty"`
	AuthorAccountID string `xml:"author-account-id,omitempty"`
	Author          string `xml:"author"`
	CreatedAt       string `xml:"created-at"`
	Body            string `xml:"body"`
	Score           int64  `xml:"score"`
	ViewerVote      int    `xml:"viewer-vote"`
}

type ThreadView struct {
	XMLName xml.Name `xml:"https://wave-lang.dev/ns/platform/community/v1 thread-view"`

	Thread     Thread        `xml:"thread"`
	Title      string        `xml:"title"`
	Root       MessageView   `xml:"root"`
	Replies    []MessageView `xml:"replies>reply"`
	Score      int64         `xml:"score"`
	ViewCount  uint64        `xml:"view-count"`
	ViewerVote int           `xml:"viewer-vote"`
	Subscribed bool          `xml:"subscribed"`
}

type Vote struct {
	XMLName    xml.Name  `xml:"https://wave-lang.dev/ns/platform/community/v1 vote"`
	TargetType string    `xml:"target-type"`
	TargetID   string    `xml:"target-id"`
	AccountID  string    `xml:"account-id"`
	Value      int       `xml:"value"`
	UpdatedAt  time.Time `xml:"updated-at"`
}

type Subscription struct {
	XMLName   xml.Name  `xml:"https://wave-lang.dev/ns/platform/community/v1 subscription"`
	ThreadID  string    `xml:"thread-id"`
	AccountID string    `xml:"account-id"`
	CreatedAt time.Time `xml:"created-at"`
}

type CreatePostInput struct {
	SpaceID string
	Title   string
	Body    string
	Tags    []string
}

type CreateReplyInput struct {
	ThreadID        string
	ParentMessageID string
	Body            string
}
