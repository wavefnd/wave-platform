package admin

import (
	"encoding/xml"
	"time"

	"github.com/wavefnd/wave-platform/internal/audit"
	"github.com/wavefnd/wave-platform/internal/gitmirror"
	maildomain "github.com/wavefnd/wave-platform/internal/mail"
)

type AccountView struct {
	ID               string    `xml:"id,attr"`
	Username         string    `xml:"username"`
	DisplayName      string    `xml:"display-name"`
	Email            string    `xml:"email"`
	Status           string    `xml:"status"`
	Owner            bool      `xml:"owner"`
	Administrator    bool      `xml:"administrator"`
	SourceMaintainer bool      `xml:"source-maintainer"`
	RFCMaintainer    bool      `xml:"rfc-maintainer"`
	TOTPEnabled      bool      `xml:"totp-enabled"`
	RecoveryVerified bool      `xml:"recovery-verified"`
	CreatedAt        time.Time `xml:"created-at"`
	UpdatedAt        time.Time `xml:"updated-at"`
}

type SecurityStatus struct {
	ActiveAccounts     int  `xml:"active-accounts"`
	SuspendedAccounts  int  `xml:"suspended-accounts"`
	TOTPAccounts       int  `xml:"totp-accounts"`
	VerifiedRecoveries int  `xml:"verified-recoveries"`
	RegistrationOpen   bool `xml:"registration-open"`
	TurnstileEnabled   bool `xml:"turnstile-enabled"`
}

type StorageStatus struct {
	DatabaseBytes int64  `xml:"database-bytes"`
	ValueLogBytes int64  `xml:"value-log-bytes"`
	FilesBytes    int64  `xml:"files-bytes"`
	Health        string `xml:"health"`
}

type MailStatus struct {
	Queued     int `xml:"queued"`
	Delivering int `xml:"delivering"`
	Deferred   int `xml:"deferred"`
	Failed     int `xml:"failed"`
	Delivered  int `xml:"delivered"`
}

type Snapshot struct {
	XMLName          xml.Name               `xml:"https://wave-lang.dev/ns/platform/api/v1 administration"`
	Accounts         []AccountView          `xml:"accounts>account"`
	Security         SecurityStatus         `xml:"security"`
	Storage          StorageStatus          `xml:"storage"`
	Mail             MailStatus             `xml:"mail-status"`
	Deliveries       []maildomain.Delivery  `xml:"deliveries>delivery"`
	Repositories     []gitmirror.Repository `xml:"git-mirrors>repository"`
	SyncInterval     string                 `xml:"git-sync-interval"`
	AuditEvents      []audit.Event          `xml:"audit-log>event"`
	GeneratedAt      time.Time              `xml:"generated-at"`
	LunaStevTimeZone string                 `xml:"lunastev-time-zone"`
}

type AccountStatusInput struct {
	Status string `xml:"status"`
}

type AccountRoleInput struct {
	Administrator bool `xml:"administrator"`
}

type SourceMaintainerInput struct {
	Enabled bool `xml:"enabled"`
}

type RFCMaintainerInput struct {
	Enabled bool `xml:"enabled"`
}

type TimeZoneInput struct {
	TimeZone string `xml:"time-zone"`
}
