package config

import (
	"encoding/xml"
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
)

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config %q: %w", path, err)
	}

	cfg := defaults()

	if err := xml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("decode config %q: %w", path, err)
	}

	applyEnvironment(&cfg)

	if err := validate(&cfg); err != nil {
		return nil, fmt.Errorf("validate config %q: %w", path, err)
	}

	return &cfg, nil
}

func defaults() Config {
	enabled := ModuleConfig{Enabled: true}

	return Config{
		Version: "1",
		Server: ServerConfig{
			Address:     "0.0.0.0:8080",
			Environment: "development",
		},
		Storage: StorageConfig{Root: "./data"},
		Frontend: FrontendConfig{
			Distribution: "./frontend/dist",
		},
		Wave: WaveConfig{
			Executable: "wave",
			Modules:    "./build/wave",
			Enabled:    true,
		},
		Identity: IdentityConfig{
			MailDomain:       "wave-lang.dev",
			RegistrationOpen: true,
			SessionHours:     720,
		},
		Mail: MailConfig{
			Hostname:        "mail.wave-lang.dev",
			SMTPAddress:     "0.0.0.0:2525",
			MaxMessageBytes: 25 * 1024 * 1024,
			DirectDelivery:  true,
		},
		Modules: ModulesConfig{
			Account:     enabled,
			Auth:        enabled,
			Session:     enabled,
			Permission:  enabled,
			Document:    enabled,
			Mail:        enabled,
			Mailbox:     enabled,
			MailingList: enabled,
			Community:   enabled,
			Question:    enabled,
			GitMirror:   enabled,
			Search:      enabled,
			Audit:       enabled,
		},
	}
}

func applyEnvironment(config *Config) {
	setString(&config.Server.Address, "WAVE_PLATFORM_ADDRESS")
	setString(&config.Storage.Root, "WAVE_PLATFORM_DATA_ROOT", "WAVE_PLATFORM_DATA_PATH")
	setString(&config.Frontend.Distribution, "WAVE_PLATFORM_WEB_ROOT", "WAVE_PLATFORM_WEB_PATH")
	setString(&config.Wave.Executable, "WAVE_PLATFORM_WAVE_EXECUTABLE")
	setString(&config.Wave.Modules, "WAVE_PLATFORM_WAVE_MODULE_PATH", "WAVE_PLATFORM_WAVE_PATH")
	setString(&config.Identity.MailDomain, "WAVE_MAIL_DOMAIN", "WAVE_PLATFORM_MAIL_DOMAIN")
	setString(&config.Identity.AdminDisplayName, "WAVE_ADMIN_DISPLAY_NAME")
	setString(&config.Identity.AdminUsername, "WAVE_ADMIN_USERNAME")
	setString(&config.Identity.AdminRecoveryEmail, "WAVE_ADMIN_RECOVERY_EMAIL")
	setString(&config.Identity.AdminTOTPSecret, "WAVE_ADMIN_TOTP_SECRET")
	setString(&config.Identity.AuthEncryptionKey, "WAVE_AUTH_ENCRYPTION_KEY")
	setString(&config.Identity.TOTPIssuer, "WAVE_TOTP_ISSUER")
	setString(&config.Identity.PublicURL, "WAVE_PUBLIC_URL")
	setString(&config.Identity.TurnstileSiteKey, "WAVE_TURNSTILE_SITE_KEY")
	setString(&config.Identity.TurnstileSecret, "WAVE_TURNSTILE_SECRET_KEY")
	setString(&config.Mail.Hostname, "WAVE_SMTP_HOSTNAME")
	setString(&config.Mail.SMTPAddress, "WAVE_SMTP_ADDRESS")
	setString(&config.Mail.TLSCertificate, "WAVE_SMTP_TLS_CERTIFICATE")
	setString(&config.Mail.TLSKey, "WAVE_SMTP_TLS_KEY")
	setString(&config.Mail.RelayAddress, "WAVE_SMTP_RELAY_ADDRESS")
	setString(&config.Mail.RelayUsername, "WAVE_SMTP_RELAY_USERNAME")
	setString(&config.Mail.RelayPassword, "WAVE_SMTP_RELAY_PASSWORD")
	setString(&config.Mail.DKIMDomain, "WAVE_DKIM_DOMAIN")
	setString(&config.Mail.DKIMSelector, "WAVE_DKIM_SELECTOR")
	setString(&config.Mail.DKIMPrivateKey, "WAVE_DKIM_PRIVATE_KEY")

	if value, ok := firstEnvironment("WAVE_PLATFORM_WAVE_ENABLED"); ok {
		config.Wave.Enabled = parseBoolean(value, config.Wave.Enabled)
	}
	if value, ok := firstEnvironment("WAVE_REGISTRATION_OPEN", "WAVE_PLATFORM_REGISTRATION_OPEN"); ok {
		config.Identity.RegistrationOpen = parseBoolean(value, config.Identity.RegistrationOpen)
	}
	if value, ok := firstEnvironment("WAVE_SESSION_HOURS"); ok {
		if hours, err := strconv.Atoi(value); err == nil {
			config.Identity.SessionHours = hours
		}
	}
	if value, ok := firstEnvironment("WAVE_SECURE_COOKIES"); ok {
		config.Identity.SecureCookies = parseBoolean(value, config.Identity.SecureCookies)
	}
	if value, ok := firstEnvironment("WAVE_SMTP_MAX_MESSAGE_BYTES"); ok {
		if size, err := strconv.ParseInt(value, 10, 64); err == nil {
			config.Mail.MaxMessageBytes = size
		}
	}
	if value, ok := firstEnvironment("WAVE_SMTP_RELAY_IMPLICIT_TLS"); ok {
		config.Mail.RelayImplicitTLS = parseBoolean(value, config.Mail.RelayImplicitTLS)
	}
	if value, ok := firstEnvironment("WAVE_SMTP_DIRECT_DELIVERY"); ok {
		config.Mail.DirectDelivery = parseBoolean(value, config.Mail.DirectDelivery)
	}
}

func setString(destination *string, names ...string) {
	if value, ok := firstEnvironment(names...); ok {
		*destination = value
	}
}

func firstEnvironment(names ...string) (string, bool) {
	for _, name := range names {
		if value, ok := os.LookupEnv(name); ok && strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value), true
		}
	}

	return "", false
}

func parseBoolean(value string, fallback bool) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		return fallback
	}
}

func validate(config *Config) error {
	if config.Version != "1" {
		return fmt.Errorf("unsupported platform config version %q", config.Version)
	}

	if _, _, err := net.SplitHostPort(config.Server.Address); err != nil {
		return fmt.Errorf("invalid server address %q: %w", config.Server.Address, err)
	}

	if config.Server.Environment == "" {
		return fmt.Errorf("server environment is required")
	}

	if config.Storage.Root == "" {
		return fmt.Errorf("storage root is required")
	}

	if config.Frontend.Distribution == "" {
		return fmt.Errorf("frontend distribution is required")
	}
	if config.Identity.MailDomain == "" || !strings.Contains(config.Identity.MailDomain, ".") {
		return fmt.Errorf("valid identity mail domain is required")
	}
	if config.Identity.SessionHours < 1 || config.Identity.SessionHours > 24*365 {
		return fmt.Errorf("identity session hours must be between 1 and 8760")
	}
	if config.Identity.PublicURL != "" {
		publicURL, err := url.Parse(config.Identity.PublicURL)
		if err != nil || (publicURL.Scheme != "http" && publicURL.Scheme != "https") || publicURL.Host == "" || publicURL.RawQuery != "" || publicURL.Fragment != "" {
			return fmt.Errorf("WAVE_PUBLIC_URL must be an absolute HTTP(S) URL without a query or fragment")
		}
		config.Identity.PublicURL = strings.TrimRight(publicURL.String(), "/")
	}
	if config.Server.Environment == "production" && config.Identity.AuthEncryptionKey == "" {
		return fmt.Errorf("WAVE_AUTH_ENCRYPTION_KEY is required in production")
	}
	if (config.Identity.TurnstileSiteKey == "") != (config.Identity.TurnstileSecret == "") {
		return fmt.Errorf("Turnstile site key and secret key must be configured together")
	}
	adminTOTPValues := 0
	for _, value := range []string{config.Identity.AdminDisplayName, config.Identity.AdminRecoveryEmail, config.Identity.AdminTOTPSecret} {
		if value != "" {
			adminTOTPValues++
		}
	}
	if adminTOTPValues != 0 && adminTOTPValues != 3 {
		return fmt.Errorf("admin display name, recovery email, and TOTP secret must be configured together")
	}
	if config.Mail.Hostname == "" || !strings.Contains(config.Mail.Hostname, ".") {
		return fmt.Errorf("valid SMTP hostname is required")
	}
	if config.Mail.SMTPAddress != "" {
		if _, _, err := net.SplitHostPort(config.Mail.SMTPAddress); err != nil {
			return fmt.Errorf("invalid SMTP address %q: %w", config.Mail.SMTPAddress, err)
		}
	}
	if config.Mail.MaxMessageBytes < 1024 || config.Mail.MaxMessageBytes > 100*1024*1024 {
		return fmt.Errorf("SMTP max message bytes must be between 1024 and 104857600")
	}
	if (config.Mail.TLSCertificate == "") != (config.Mail.TLSKey == "") {
		return fmt.Errorf("SMTP TLS certificate and key must be configured together")
	}
	dkimConfigured := 0
	for _, value := range []string{config.Mail.DKIMDomain, config.Mail.DKIMSelector, config.Mail.DKIMPrivateKey} {
		if value != "" {
			dkimConfigured++
		}
	}
	if dkimConfigured != 0 && dkimConfigured != 3 {
		return fmt.Errorf("DKIM domain, selector, and private key must be configured together")
	}

	return nil
}
