package permission

import (
	"encoding/xml"
	"errors"
	"fmt"

	"github.com/wavefnd/wave-platform/internal/storage"
)

type Repository struct{ database *storage.Database }

func NewRepository(database *storage.Database) *Repository { return &Repository{database: database} }

func (repository *Repository) PutRole(role Role) error {
	if role.ID == "" {
		return errors.New("role id is required")
	}
	data, err := xml.Marshal(role)
	if err != nil {
		return fmt.Errorf("encode role: %w", err)
	}
	return repository.database.Set(storage.Key("permission", "role", role.ID), data)
}

func (repository *Repository) Assign(assignment Assignment) error {
	if assignment.AccountID == "" || assignment.RoleID == "" {
		return errors.New("assignment account and role ids are required")
	}
	data, err := xml.Marshal(assignment)
	if err != nil {
		return fmt.Errorf("encode assignment: %w", err)
	}
	return repository.database.Set(storage.Key("permission", "assignment", assignment.AccountID, assignment.RoleID), data)
}

func (repository *Repository) HasRole(accountID, roleID string) (bool, error) {
	_, err := repository.database.Get(storage.Key("permission", "assignment", accountID, roleID))
	if errors.Is(err, storage.ErrNotFound) {
		return false, nil
	}
	return err == nil, err
}

func (repository *Repository) Unassign(accountID, roleID string) error {
	return repository.database.Delete(storage.Key("permission", "assignment", accountID, roleID))
}
