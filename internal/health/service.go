package health

import (
	"encoding/xml"
)

type Status struct {
	XMLName xml.Name `xml:"https://wave-lang.dev/ns/platform/api/v1 health"`

	Status    string `xml:"status"`
	Ready     bool   `xml:"ready"`
	Timestamp string `xml:"timestamp"`
	Database  string `xml:"database"`
}
