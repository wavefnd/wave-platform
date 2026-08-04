package account

import (
	"encoding/xml"
	"time"
)

type Account struct {
	XMLName     xml.Name  `xml:"https://wave-lang.dev/ns/platform/account/v1 account"`
	ID          string    `xml:"id,attr"`
	Username    string    `xml:"username"`
	DisplayName string    `xml:"display-name"`
	Email       string    `xml:"email"`
	Status      string    `xml:"status"`
	CreatedAt   time.Time `xml:"created-at"`
	UpdatedAt   time.Time `xml:"updated-at"`
}
