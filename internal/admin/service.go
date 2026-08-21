package admin

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
	"time"

	"github.com/wavefnd/wave-platform/internal/account"
	"github.com/wavefnd/wave-platform/internal/audit"
	"github.com/wavefnd/wave-platform/internal/auth"
	"github.com/wavefnd/wave-platform/internal/gitmirror"
	"github.com/wavefnd/wave-platform/internal/identifier"
	maildomain "github.com/wavefnd/wave-platform/internal/mail"
	"github.com/wavefnd/wave-platform/internal/permission"
	"github.com/wavefnd/wave-platform/internal/session"
	"github.com/wavefnd/wave-platform/internal/storage"
)

var (
	ErrForbidden     = errors.New("management action is forbidden")
	ErrInvalidStatus = errors.New("invalid account status")
	ErrSelfAction    = errors.New("administrators cannot change their own access")
)

const GitSyncInterval = 15 * time.Minute

type Service struct {
	database         *storage.Database
	accounts         *account.Repository
	audit            *audit.Repository
	deliveries       *maildomain.Repository
	permissions      *permission.Repository
	sessions         *session.Repository
	gitMirror        *gitmirror.Service
	registrationOpen bool
	turnstileEnabled bool
	now              func() time.Time
}

func NewService(database *storage.Database, gitMirror *gitmirror.Service, registrationOpen, turnstileEnabled bool) *Service {
	return &Service{database: database, accounts: account.NewRepository(database), audit: audit.NewRepository(database),
		deliveries: maildomain.NewRepository(database), permissions: permission.NewRepository(database),
		sessions: session.NewRepository(database), gitMirror: gitMirror, registrationOpen: registrationOpen,
		turnstileEnabled: turnstileEnabled, now: time.Now}
}

func (service *Service) Snapshot(ctx context.Context, query string) (Snapshot, error) {
	items, err := service.accounts.Accounts()
	if err != nil {
		return Snapshot{}, fmt.Errorf("list accounts: %w", err)
	}
	factors, err := service.factors()
	if err != nil {
		return Snapshot{}, err
	}
	result := Snapshot{GeneratedAt: service.now().UTC(), SyncInterval: GitSyncInterval.String(), LunaStevTimeZone: service.LunaStevTimeZone()}
	query = strings.ToLower(strings.TrimSpace(query))
	for _, item := range items {
		factor, hasFactor := factors[item.ID]
		owner, err := service.permissions.HasRole(item.ID, "platform-owner")
		if err != nil {
			return Snapshot{}, err
		}
		administrator, err := service.permissions.HasRole(item.ID, "platform-admin")
		if err != nil {
			return Snapshot{}, err
		}
		if owner {
			administrator = true
		}
		sourceMaintainer, err := service.permissions.HasRole(item.ID, "source-maintainer")
		if err != nil {
			return Snapshot{}, err
		}
		if owner {
			sourceMaintainer = true
		}
		rfcMaintainer, err := service.permissions.HasRole(item.ID, "rfc-maintainer")
		if err != nil {
			return Snapshot{}, err
		}
		if owner {
			rfcMaintainer = true
		}
		if item.Status == "active" {
			result.Security.ActiveAccounts++
		} else {
			result.Security.SuspendedAccounts++
		}
		if hasFactor {
			result.Security.TOTPAccounts++
		}
		if factor.RecoveryVerified {
			result.Security.VerifiedRecoveries++
		}
		searchable := strings.ToLower(item.Username + " " + item.DisplayName + " " + item.Email)
		if query != "" && !strings.Contains(searchable, query) {
			continue
		}
		result.Accounts = append(result.Accounts, AccountView{ID: item.ID, Username: item.Username,
			DisplayName: item.DisplayName, Email: item.Email, Status: item.Status, Owner: owner,
			Administrator: administrator, SourceMaintainer: sourceMaintainer, RFCMaintainer: rfcMaintainer, TOTPEnabled: hasFactor, RecoveryVerified: factor.RecoveryVerified,
			CreatedAt: item.CreatedAt, UpdatedAt: item.UpdatedAt})
	}
	result.Security.RegistrationOpen = service.registrationOpen
	result.Security.TurnstileEnabled = service.turnstileEnabled
	result.Deliveries, err = service.deliveries.Deliveries(50)
	if err != nil {
		return Snapshot{}, fmt.Errorf("list mail deliveries: %w", err)
	}
	for _, delivery := range result.Deliveries {
		switch delivery.Status {
		case "queued":
			result.Mail.Queued++
		case "delivering":
			result.Mail.Delivering++
		case "deferred":
			result.Mail.Deferred++
		case "failed":
			result.Mail.Failed++
		case "delivered":
			result.Mail.Delivered++
		}
	}
	if service.gitMirror != nil {
		result.Repositories, err = service.gitMirror.Repositories(ctx)
		if err != nil {
			return Snapshot{}, fmt.Errorf("list Git mirrors: %w", err)
		}
	}
	result.AuditEvents, err = service.audit.Events(100)
	if err != nil {
		return Snapshot{}, fmt.Errorf("list audit events: %w", err)
	}
	result.Storage = service.storageStatus()
	return result, nil
}

func (service *Service) LunaStevTimeZone() string {
	data, err := service.database.Get(storage.Key("setting", "lunastev", "time-zone"))
	if err != nil || strings.TrimSpace(string(data)) == "" {
		return "Asia/Seoul"
	}
	return string(data)
}

func (service *Service) UpdateLunaStevTimeZone(actorID, value string) error {
	value = strings.TrimSpace(value)
	if len(value) > 64 {
		return errors.New("invalid time zone")
	}
	if _, err := time.LoadLocation(value); err != nil {
		return errors.New("invalid time zone")
	}
	if err := service.database.Set(storage.Key("setting", "lunastev", "time-zone"), []byte(value)); err != nil {
		return err
	}
	return service.appendAudit(actorID, "setting/lunastev/time-zone", "admin.setting.update")
}

func (service *Service) UpdateAccountStatus(actorID, accountID, status string) error {
	status = strings.ToLower(strings.TrimSpace(status))
	if status != "active" && status != "suspended" {
		return ErrInvalidStatus
	}
	if actorID == accountID {
		return ErrSelfAction
	}
	if err := service.canManage(actorID, accountID); err != nil {
		return err
	}
	item, err := service.accounts.Account(accountID)
	if err != nil {
		return err
	}
	if item.Status == status {
		return nil
	}
	item.Status = status
	item.UpdatedAt = service.now().UTC()
	if err := service.accounts.Update(item); err != nil {
		return err
	}
	if status == "suspended" {
		if err := service.sessions.DeleteByAccount(accountID); err != nil {
			return err
		}
	}
	return service.appendAudit(actorID, "account/"+accountID, "admin.account."+status)
}

func (service *Service) UpdateAdministrator(actorID, accountID string, enabled bool) error {
	if actorID == accountID {
		return ErrSelfAction
	}
	actorOwner, err := service.permissions.HasRole(actorID, "platform-owner")
	if err != nil {
		return err
	}
	if !actorOwner {
		return ErrForbidden
	}
	if _, err := service.accounts.Account(accountID); err != nil {
		return err
	}
	targetOwner, err := service.permissions.HasRole(accountID, "platform-owner")
	if err != nil {
		return err
	}
	if targetOwner {
		return ErrForbidden
	}
	if enabled {
		err = service.permissions.Assign(permission.Assignment{AccountID: accountID, RoleID: "platform-admin", Scope: "platform"})
	} else {
		err = service.permissions.Unassign(accountID, "platform-admin")
	}
	if err != nil {
		return err
	}
	if !enabled {
		_ = service.sessions.DeleteByAccount(accountID)
	}
	action := "admin.role.remove"
	if enabled {
		action = "admin.role.assign"
	}
	return service.appendAudit(actorID, "account/"+accountID+"/role/platform-admin", action)
}

func (service *Service) UpdateSourceMaintainer(actorID, accountID string, enabled bool) error {
	actorOwner, err := service.permissions.HasRole(actorID, "platform-owner")
	if err != nil {
		return err
	}
	if !actorOwner {
		return ErrForbidden
	}
	if _, err := service.accounts.Account(accountID); err != nil {
		return err
	}
	targetOwner, err := service.permissions.HasRole(accountID, "platform-owner")
	if err != nil {
		return err
	}
	if targetOwner {
		return ErrForbidden
	}
	if enabled {
		err = service.permissions.Assign(permission.Assignment{AccountID: accountID, RoleID: "source-maintainer", Scope: "source"})
	} else {
		err = service.permissions.Unassign(accountID, "source-maintainer")
	}
	if err != nil {
		return err
	}
	action := "admin.source-maintainer.remove"
	if enabled {
		action = "admin.source-maintainer.assign"
	}
	return service.appendAudit(actorID, "account/"+accountID+"/role/source-maintainer", action)
}

func (service *Service) UpdateRFCMaintainer(actorID, accountID string, enabled bool) error {
	actorOwner, err := service.permissions.HasRole(actorID, "platform-owner")
	if err != nil {
		return err
	}
	if !actorOwner {
		return ErrForbidden
	}
	if _, err := service.accounts.Account(accountID); err != nil {
		return err
	}
	targetOwner, err := service.permissions.HasRole(accountID, "platform-owner")
	if err != nil {
		return err
	}
	if targetOwner {
		return ErrForbidden
	}
	if enabled {
		err = service.permissions.Assign(permission.Assignment{AccountID: accountID, RoleID: "rfc-maintainer", Scope: "rfc"})
	} else {
		err = service.permissions.Unassign(accountID, "rfc-maintainer")
	}
	if err != nil {
		return err
	}
	action := "admin.rfc-maintainer.remove"
	if enabled {
		action = "admin.rfc-maintainer.assign"
	}
	return service.appendAudit(actorID, "account/"+accountID+"/role/rfc-maintainer", action)
}

func (service *Service) canManage(actorID, targetID string) error {
	actorOwner, err := service.permissions.HasRole(actorID, "platform-owner")
	if err != nil {
		return err
	}
	targetOwner, err := service.permissions.HasRole(targetID, "platform-owner")
	if err != nil {
		return err
	}
	if targetOwner && !actorOwner {
		return ErrForbidden
	}
	return nil
}

func (service *Service) factors() (map[string]auth.TOTPFactor, error) {
	result := map[string]auth.TOTPFactor{}
	err := service.database.Scan(storage.Prefix("auth", "totp", "factor"), func(_, value []byte) error {
		var factor auth.TOTPFactor
		if err := xml.Unmarshal(value, &factor); err != nil {
			return fmt.Errorf("decode TOTP factor: %w", err)
		}
		result[factor.AccountID] = factor
		return nil
	})
	return result, err
}

func (service *Service) storageStatus() StorageStatus {
	result := StorageStatus{Health: "ready"}
	if err := service.database.Health(); err != nil {
		result.Health = "unavailable"
	}
	result.DatabaseBytes, result.ValueLogBytes = service.database.DB.Size()
	_ = filepath.WalkDir(service.database.Root, func(_ string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return nil
		}
		if info, infoErr := entry.Info(); infoErr == nil {
			result.FilesBytes += info.Size()
		}
		return nil
	})
	return result
}

func (service *Service) appendAudit(actorID, resourceID, action string) error {
	id, err := identifier.New("audit")
	if err != nil {
		return err
	}
	return service.audit.Append(audit.Event{ID: id, ActorID: "account/" + actorID, ResourceID: resourceID,
		Action: action, Result: "success", OccurredAt: service.now().UTC()})
}
