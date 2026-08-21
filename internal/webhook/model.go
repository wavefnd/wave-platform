package webhook

import (
	"encoding/xml"
	"time"
)

const (
	EventBlogPublished    = "blog.published"
	EventCommunityPost    = "community.post.created"
	EventFounderPost      = "founder.post.created"
	EventReleasePublished = "release.published"
	EventPatchReceived    = "patch.received"
)

type Endpoint struct {
	XMLName                xml.Name  `xml:"https://wave-lang.dev/ns/platform/webhook/v1 endpoint"`
	ID                     string    `xml:"id,attr"`
	OwnerAccountID         string    `xml:"owner-account-id"`
	Name                   string    `xml:"name"`
	Kind                   string    `xml:"kind"`
	Events                 []string  `xml:"events>event"`
	EncryptedURL           string    `xml:"encrypted-url"`
	EncryptedSigningSecret string    `xml:"encrypted-signing-secret"`
	Destination            string    `xml:"destination"`
	Enabled                bool      `xml:"enabled"`
	CreatedAt              time.Time `xml:"created-at"`
	UpdatedAt              time.Time `xml:"updated-at"`
}

type EndpointView struct {
	XMLName        xml.Name  `xml:"https://wave-lang.dev/ns/platform/api/v1 webhook"`
	ID             string    `xml:"id,attr"`
	OwnerAccountID string    `xml:"owner-account-id,omitempty"`
	Name           string    `xml:"name"`
	Kind           string    `xml:"kind"`
	Events         []string  `xml:"events>event"`
	Destination    string    `xml:"destination"`
	Enabled        bool      `xml:"enabled"`
	CreatedAt      time.Time `xml:"created-at"`
	UpdatedAt      time.Time `xml:"updated-at"`
	SigningSecret  string    `xml:"signing-secret,omitempty"`
}

type EndpointInput struct {
	ID           string   `xml:"id"`
	Name         string   `xml:"name"`
	Kind         string   `xml:"kind"`
	URL          string   `xml:"url"`
	Events       []string `xml:"events>event"`
	Enabled      bool     `xml:"enabled"`
	RotateSecret bool     `xml:"rotate-secret"`
}

type Event struct {
	ID         string    `json:"id"`
	Type       string    `json:"type"`
	Title      string    `json:"title"`
	Summary    string    `json:"summary,omitempty"`
	AuthorName string    `json:"author_name,omitempty"`
	ResourceID string    `json:"resource_id"`
	URL        string    `json:"url"`
	OccurredAt time.Time `json:"occurred_at"`
}

type Delivery struct {
	XMLName       xml.Name  `xml:"https://wave-lang.dev/ns/platform/webhook/v1 delivery"`
	ID            string    `xml:"id,attr"`
	EndpointID    string    `xml:"endpoint-id"`
	EventID       string    `xml:"event-id"`
	EventType     string    `xml:"event-type"`
	Title         string    `xml:"title"`
	Summary       string    `xml:"summary,omitempty"`
	AuthorName    string    `xml:"author-name,omitempty"`
	ResourceID    string    `xml:"resource-id"`
	ResourceURL   string    `xml:"resource-url"`
	Status        string    `xml:"status"`
	Attempts      int       `xml:"attempts"`
	HTTPStatus    int       `xml:"http-status,omitempty"`
	LastError     string    `xml:"last-error,omitempty"`
	NextAttemptAt time.Time `xml:"next-attempt-at,omitempty"`
	CreatedAt     time.Time `xml:"created-at"`
	LastAttemptAt time.Time `xml:"last-attempt-at,omitempty"`
	CompletedAt   time.Time `xml:"completed-at,omitempty"`
}
