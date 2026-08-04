package permission

import "encoding/xml"

type Role struct {
	XMLName     xml.Name `xml:"https://wave-lang.dev/ns/platform/permission/v1 role"`
	ID          string   `xml:"id,attr"`
	Name        string   `xml:"name"`
	Description string   `xml:"description"`
	Permissions []string `xml:"permissions>permission"`
}

type Assignment struct {
	XMLName   xml.Name `xml:"https://wave-lang.dev/ns/platform/permission/v1 assignment"`
	AccountID string   `xml:"account-id"`
	RoleID    string   `xml:"role-id"`
	Scope     string   `xml:"scope"`
}

type Decision struct {
	Allowed bool
	Reason  string
}
