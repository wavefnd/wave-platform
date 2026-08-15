package admin

import (
	"errors"
	"testing"
	"time"

	"github.com/wavefnd/wave-platform/internal/account"
	"github.com/wavefnd/wave-platform/internal/audit"
	"github.com/wavefnd/wave-platform/internal/permission"
	"github.com/wavefnd/wave-platform/internal/session"
	"github.com/wavefnd/wave-platform/internal/storage"
)

func TestAccountManagementProtectsOwnersAndWritesAuditLog(t *testing.T) {
	database, err := storage.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = database.Close() })
	now := time.Now().UTC().Truncate(time.Second)
	accounts := account.NewRepository(database)
	permissions := permission.NewRepository(database)
	for _, item := range []account.Account{
		{ID: "owner", Username: "owner", DisplayName: "Owner", Email: "owner@wave.test", Status: "active", CreatedAt: now, UpdatedAt: now},
		{ID: "admin", Username: "admin", DisplayName: "Admin", Email: "admin@wave.test", Status: "active", CreatedAt: now, UpdatedAt: now},
		{ID: "member", Username: "member", DisplayName: "Member", Email: "member@wave.test", Status: "active", CreatedAt: now, UpdatedAt: now},
	} {
		if err := accounts.Create(item); err != nil {
			t.Fatal(err)
		}
	}
	if err := permissions.Assign(permission.Assignment{AccountID: "owner", RoleID: "platform-owner", Scope: "platform"}); err != nil {
		t.Fatal(err)
	}
	if err := permissions.Assign(permission.Assignment{AccountID: "admin", RoleID: "platform-admin", Scope: "platform"}); err != nil {
		t.Fatal(err)
	}

	service := NewService(database, nil, true, false)
	service.now = func() time.Time { return now.Add(time.Hour) }
	if err := service.UpdateAccountStatus("admin", "owner", "suspended"); !errors.Is(err, ErrForbidden) {
		t.Fatalf("administrator changing owner status: %v", err)
	}
	if err := service.UpdateAccountStatus("admin", "admin", "suspended"); !errors.Is(err, ErrSelfAction) {
		t.Fatalf("administrator changing own status: %v", err)
	}

	sessions := session.NewRepository(database)
	current := session.Session{ID: "session-member", AccountID: "member", CreatedAt: now, ExpiresAt: now.Add(24 * time.Hour)}
	if err := sessions.Put(current); err != nil {
		t.Fatal(err)
	}
	if err := service.UpdateAccountStatus("admin", "member", "suspended"); err != nil {
		t.Fatal(err)
	}
	member, err := accounts.Account("member")
	if err != nil {
		t.Fatal(err)
	}
	if member.Status != "suspended" || !member.UpdatedAt.Equal(now.Add(time.Hour)) {
		t.Fatalf("updated member = %#v", member)
	}
	if _, err := sessions.Session(current.ID); !errors.Is(err, storage.ErrNotFound) {
		t.Fatalf("suspended account session still exists: %v", err)
	}
	events, err := audit.NewRepository(database).Events(10)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 || events[0].Action != "admin.account.suspended" || events[0].ActorID != "account/admin" {
		t.Fatalf("audit events = %#v", events)
	}
}

func TestOnlyOwnerCanChangeAdministratorRole(t *testing.T) {
	database, err := storage.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = database.Close() })
	now := time.Now().UTC()
	accounts := account.NewRepository(database)
	permissions := permission.NewRepository(database)
	for _, id := range []string{"owner", "admin", "member"} {
		if err := accounts.Create(account.Account{ID: id, Username: id, DisplayName: id, Email: id + "@wave.test", Status: "active", CreatedAt: now, UpdatedAt: now}); err != nil {
			t.Fatal(err)
		}
	}
	_ = permissions.Assign(permission.Assignment{AccountID: "owner", RoleID: "platform-owner", Scope: "platform"})
	_ = permissions.Assign(permission.Assignment{AccountID: "admin", RoleID: "platform-admin", Scope: "platform"})
	service := NewService(database, nil, false, false)
	if err := service.UpdateAdministrator("admin", "member", true); !errors.Is(err, ErrForbidden) {
		t.Fatalf("administrator assigned role: %v", err)
	}
	if err := service.UpdateAdministrator("owner", "member", true); err != nil {
		t.Fatal(err)
	}
	assigned, err := permissions.HasRole("member", "platform-admin")
	if err != nil || !assigned {
		t.Fatalf("administrator role assigned=%v err=%v", assigned, err)
	}
	if err := service.UpdateAdministrator("owner", "member", false); err != nil {
		t.Fatal(err)
	}
	assigned, err = permissions.HasRole("member", "platform-admin")
	if err != nil || assigned {
		t.Fatalf("administrator role assigned=%v err=%v", assigned, err)
	}
	if err := service.UpdateAdministrator("owner", "owner", false); !errors.Is(err, ErrSelfAction) {
		t.Fatalf("owner changed own role: %v", err)
	}
}

func TestLunaStevTimeZoneDefaultsToSeoulAndIsAudited(t *testing.T) {
	database, err := storage.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = database.Close() })
	service := NewService(database, nil, true, false)
	if value := service.LunaStevTimeZone(); value != "Asia/Seoul" {
		t.Fatalf("default time zone = %q", value)
	}
	if err := service.UpdateLunaStevTimeZone("owner", "America/New_York"); err != nil {
		t.Fatal(err)
	}
	if value := service.LunaStevTimeZone(); value != "America/New_York" {
		t.Fatalf("updated time zone = %q", value)
	}
	if err := service.UpdateLunaStevTimeZone("owner", "not/a-zone"); err == nil {
		t.Fatal("invalid time zone was accepted")
	}
	events, err := audit.NewRepository(database).Events(10)
	if err != nil || len(events) != 1 || events[0].Action != "admin.setting.update" {
		t.Fatalf("events=%#v err=%v", events, err)
	}
}
