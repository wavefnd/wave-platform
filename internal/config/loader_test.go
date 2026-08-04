package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadAppliesDefaultsAndEnvironment(t *testing.T) {
	root := t.TempDir()
	configPath := filepath.Join(root, "platform.xml")
	dataRoot := filepath.Join(root, "data-override")

	configuration := `<?xml version="1.0" encoding="UTF-8"?>
<platform xmlns="https://wave-lang.dev/ns/platform/config/v1" version="1">
    <server><environment>test</environment></server>
    <modules><mail enabled="false"/></modules>
</platform>`

	if err := os.WriteFile(configPath, []byte(configuration), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	t.Setenv("WAVE_PLATFORM_ADDRESS", "127.0.0.1:9090")
	t.Setenv("WAVE_PLATFORM_DATA_ROOT", dataRoot)
	t.Setenv("WAVE_MAIL_DOMAIN", "mail.wave.test")
	t.Setenv("WAVE_ADMIN_DISPLAY_NAME", "Wave Administrator")
	t.Setenv("WAVE_ADMIN_RECOVERY_EMAIL", "owner@example.com")
	t.Setenv("WAVE_ADMIN_TOTP_SECRET", "JBSWY3DPEHPK3PXP")
	t.Setenv("WAVE_AUTH_ENCRYPTION_KEY", "MDEyMzQ1Njc4OWFiY2RlZjAxMjM0NTY3ODlhYmNkZWY=")
	t.Setenv("WAVE_PUBLIC_URL", "https://wave.example/")

	config, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if config.Server.Address != "127.0.0.1:9090" {
		t.Fatalf("address = %q", config.Server.Address)
	}
	if config.Storage.Root != dataRoot {
		t.Fatalf("data root = %q", config.Storage.Root)
	}
	if config.Modules.Mail.Enabled {
		t.Fatal("mail module should be disabled")
	}
	if !config.Modules.Document.Enabled {
		t.Fatal("document module should keep the enabled default")
	}
	if config.Wave.Executable != "wave" || !config.Wave.Enabled {
		t.Fatalf("unexpected Wave defaults: %+v", config.Wave)
	}
	if config.Identity.MailDomain != "mail.wave.test" || config.Identity.AdminDisplayName != "Wave Administrator" {
		t.Fatalf("unexpected identity config: %+v", config.Identity)
	}
	if config.Identity.PublicURL != "https://wave.example" {
		t.Fatalf("public URL = %q", config.Identity.PublicURL)
	}
}

func TestLoadRejectsUnsupportedVersion(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "platform.xml")
	if err := os.WriteFile(configPath, []byte(`<platform xmlns="https://wave-lang.dev/ns/platform/config/v1" version="2"/>`), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	if _, err := Load(configPath); err == nil {
		t.Fatal("Load() should reject an unsupported version")
	}
}

func TestLoadRejectsInvalidPublicURL(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "platform.xml")
	configuration := `<platform xmlns="https://wave-lang.dev/ns/platform/config/v1" version="1"><server><environment>test</environment></server></platform>`
	if err := os.WriteFile(configPath, []byte(configuration), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	t.Setenv("WAVE_PUBLIC_URL", "javascript:alert(1)")
	if _, err := Load(configPath); err == nil {
		t.Fatal("Load() should reject an invalid public URL")
	}
}
