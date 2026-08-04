package question

import (
	"encoding/xml"
	"time"
)

const XMLNamespace = "https://wave-lang.dev/ns/platform/question/v1"

// Question owns Q&A metadata. The question and answer bodies are immutable
// mail.Message objects in one mail thread and are never projected into a
// user's personal mailbox.
type Question struct {
	XMLName           xml.Name `xml:"https://wave-lang.dev/ns/platform/question/v1 question"`
	ID                string   `xml:"id,attr"`
	RootMessageID     string   `xml:"root-message-id"`
	AcceptedMessageID string   `xml:"accepted-message-id,omitempty"`
	Status            string   `xml:"status"`
	Tags              []string `xml:"tags>tag,omitempty"`
	WaveVersion       string   `xml:"wave-version,omitempty"`
	Platform          string   `xml:"platform,omitempty"`
}

type Summary struct {
	ID              string   `xml:"id"`
	Title           string   `xml:"title"`
	Excerpt         string   `xml:"excerpt"`
	Author          string   `xml:"author"`
	AuthorAccountID string   `xml:"author-account-id,omitempty"`
	CreatedAt       string   `xml:"created-at"`
	LastActivityAt  string   `xml:"last-activity-at"`
	Tags            []string `xml:"tags>tag,omitempty"`
	WaveVersion     string   `xml:"wave-version,omitempty"`
	Platform        string   `xml:"platform,omitempty"`
	Status          string   `xml:"status"`
	Score           int64    `xml:"score"`
	ViewerVote      int      `xml:"viewer-vote"`
	AnswerCount     int      `xml:"answer-count"`
	ViewCount       uint64   `xml:"view-count"`
	Accepted        bool     `xml:"accepted"`
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
	Accepted        bool   `xml:"accepted"`
}

type View struct {
	XMLName    xml.Name      `xml:"https://wave-lang.dev/ns/platform/question/v1 question-view"`
	Question   Question      `xml:"question"`
	Title      string        `xml:"title"`
	Root       MessageView   `xml:"root"`
	Answers    []MessageView `xml:"answers>answer"`
	Score      int64         `xml:"score"`
	ViewCount  uint64        `xml:"view-count"`
	ViewerVote int           `xml:"viewer-vote"`
}

type Vote struct {
	XMLName    xml.Name  `xml:"https://wave-lang.dev/ns/platform/question/v1 vote"`
	TargetType string    `xml:"target-type"`
	TargetID   string    `xml:"target-id"`
	AccountID  string    `xml:"account-id"`
	Value      int       `xml:"value"`
	UpdatedAt  time.Time `xml:"updated-at"`
}

type CreateInput struct {
	Title       string
	Body        string
	Tags        []string
	WaveVersion string
	Platform    string
}

type AnswerInput struct {
	QuestionID string
	Body       string
}
