package config

import "encoding/xml"

type Config struct {
	XMLName xml.Name `xml:"https://wave-lang.dev/ns/platform/config/v1 platform"`

	Version string `xml:"version,attr"`

	Server   ServerConfig   `xml:"server"`
	Storage  StorageConfig  `xml:"storage"`
	Frontend FrontendConfig `xml:"frontend"`
	Wave     WaveConfig     `xml:"wave"`
	Identity IdentityConfig `xml:"identity"`
	Mail     MailConfig     `xml:"mail"`
	Modules  ModulesConfig  `xml:"modules"`
}

// MailConfig controls the SMTP edges of the platform. Messages and delivery
// state remain in the shared storage layer; these values only describe how the
// single platform process accepts and transfers Internet mail.
type MailConfig struct {
	Hostname         string `xml:"hostname"`
	SMTPAddress      string `xml:"smtp-address"`
	MaxMessageBytes  int64  `xml:"max-message-bytes"`
	TLSCertificate   string `xml:"tls-certificate"`
	TLSKey           string `xml:"tls-key"`
	RelayAddress     string `xml:"relay-address"`
	RelayUsername    string `xml:"-"`
	RelayPassword    string `xml:"-"`
	RelayImplicitTLS bool   `xml:"relay-implicit-tls"`
	DirectDelivery   bool   `xml:"direct-delivery"`
	DKIMDomain       string `xml:"dkim-domain"`
	DKIMSelector     string `xml:"dkim-selector"`
	DKIMPrivateKey   string `xml:"dkim-private-key"`
}

type ServerConfig struct {
	Address     string `xml:"address"`
	Environment string `xml:"environment"`
}

type StorageConfig struct {
	Root string `xml:"root"`
}

type FrontendConfig struct {
	Distribution string `xml:"distribution"`
}

type WaveConfig struct {
	Executable string `xml:"executable"`
	Modules    string `xml:"modules"`
	Enabled    bool   `xml:"enabled"`
}

type IdentityConfig struct {
	MailDomain         string `xml:"mail-domain"`
	RegistrationOpen   bool   `xml:"registration-open"`
	SessionHours       int    `xml:"session-hours"`
	SecureCookies      bool   `xml:"secure-cookies"`
	AdminDisplayName   string `xml:"-"`
	AdminUsername      string `xml:"-"`
	AdminRecoveryEmail string `xml:"-"`
	AdminTOTPSecret    string `xml:"-"`
	AuthEncryptionKey  string `xml:"-"`
	TOTPIssuer         string `xml:"-"`
	PublicURL          string `xml:"-"`
	TurnstileSiteKey   string `xml:"-"`
	TurnstileSecret    string `xml:"-"`
}

type ModuleConfig struct {
	Enabled bool `xml:"enabled,attr"`
}

type ModulesConfig struct {
	Account     ModuleConfig `xml:"account"`
	Auth        ModuleConfig `xml:"auth"`
	Session     ModuleConfig `xml:"session"`
	Permission  ModuleConfig `xml:"permission"`
	Document    ModuleConfig `xml:"document"`
	Mail        ModuleConfig `xml:"mail"`
	Mailbox     ModuleConfig `xml:"mailbox"`
	MailingList ModuleConfig `xml:"mailingList"`
	Community   ModuleConfig `xml:"community"`
	Question    ModuleConfig `xml:"question"`
	GitMirror   ModuleConfig `xml:"gitMirror"`
	Search      ModuleConfig `xml:"search"`
	Audit       ModuleConfig `xml:"audit"`
}

type ModuleStatus struct {
	Name    string
	Enabled bool
}

func (config ModulesConfig) Statuses() []ModuleStatus {
	return []ModuleStatus{
		{Name: "account", Enabled: config.Account.Enabled},
		{Name: "auth", Enabled: config.Auth.Enabled},
		{Name: "session", Enabled: config.Session.Enabled},
		{Name: "permission", Enabled: config.Permission.Enabled},
		{Name: "document", Enabled: config.Document.Enabled},
		{Name: "mail", Enabled: config.Mail.Enabled},
		{Name: "mailbox", Enabled: config.Mailbox.Enabled},
		{Name: "mailing-list", Enabled: config.MailingList.Enabled},
		{Name: "community", Enabled: config.Community.Enabled},
		{Name: "question", Enabled: config.Question.Enabled},
		{Name: "git-mirror", Enabled: config.GitMirror.Enabled},
		{Name: "search", Enabled: config.Search.Enabled},
		{Name: "audit", Enabled: config.Audit.Enabled},
	}
}
