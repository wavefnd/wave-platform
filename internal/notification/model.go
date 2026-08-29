package notification

import "time"

const MaxStoredPerAccount = 200

type Item struct {
	ID                 string     `xml:"id"`
	RecipientAccountID string     `xml:"recipient-account-id"`
	ActorAccountID     string     `xml:"actor-account-id,omitempty"`
	ActorName          string     `xml:"actor-name,omitempty"`
	Kind               string     `xml:"kind"`
	Subject            string     `xml:"subject"`
	Detail             string     `xml:"detail,omitempty"`
	URL                string     `xml:"url"`
	CreatedAt          time.Time  `xml:"created-at"`
	ReadAt             *time.Time `xml:"read-at,omitempty"`
}

type Input struct {
	RecipientAccountID string
	ActorAccountID     string
	ActorName          string
	Kind               string
	Subject            string
	Detail             string
	URL                string
}
