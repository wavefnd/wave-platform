package account

import (
	"encoding/xml"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/dgraph-io/badger/v4"
	"github.com/wavefnd/wave-platform/internal/storage"
)

var ErrConflict = errors.New("account identifier already exists")

type Repository struct {
	database *storage.Database
}

func NewRepository(database *storage.Database) *Repository {
	return &Repository{database: database}
}

func (repository *Repository) Create(item Account) error {
	if item.ID == "" || item.Username == "" || item.Email == "" {
		return errors.New("account id, username, and email are required")
	}
	data, err := xml.Marshal(item)
	if err != nil {
		return fmt.Errorf("encode account: %w", err)
	}
	username := strings.ToLower(item.Username)
	email := strings.ToLower(item.Email)
	err = repository.database.DB.Update(func(transaction *badger.Txn) error {
		for _, key := range [][]byte{
			storage.Key("account", "object", item.ID),
			storage.Key("account", "name", username),
			storage.Key("account", "email", email),
			storage.Key("account", "profile-alias", username),
		} {
			if _, getErr := transaction.Get(key); getErr == nil {
				return ErrConflict
			} else if !errors.Is(getErr, badger.ErrKeyNotFound) {
				return getErr
			}
		}
		if err := transaction.Set(storage.Key("account", "object", item.ID), data); err != nil {
			return err
		}
		if err := transaction.Set(storage.Key("account", "name", username), []byte(item.ID)); err != nil {
			return err
		}
		return transaction.Set(storage.Key("account", "email", email), []byte(item.ID))
	})
	if err != nil {
		return fmt.Errorf("create account: %w", err)
	}
	return nil
}

func (repository *Repository) Account(id string) (Account, error) {
	return repository.read(storage.Key("account", "object", id))
}

func (repository *Repository) Accounts() ([]Account, error) {
	items := make([]Account, 0)
	err := repository.database.Scan(storage.Prefix("account", "object"), func(_, value []byte) error {
		var item Account
		if err := xml.Unmarshal(value, &item); err != nil {
			return fmt.Errorf("decode account: %w", err)
		}
		items = append(items, item)
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(items, func(left, right int) bool {
		return items[left].CreatedAt.After(items[right].CreatedAt)
	})
	return items, nil
}

func (repository *Repository) Update(item Account) error {
	if item.ID == "" || item.Username == "" || item.Email == "" {
		return errors.New("account id, username, and email are required")
	}
	previous, err := repository.Account(item.ID)
	if err != nil {
		return err
	}
	data, err := xml.Marshal(item)
	if err != nil {
		return fmt.Errorf("encode account: %w", err)
	}
	username := strings.ToLower(item.Username)
	email := strings.ToLower(item.Email)
	previousUsername := strings.ToLower(previous.Username)
	previousEmail := strings.ToLower(previous.Email)
	err = repository.database.DB.Update(func(transaction *badger.Txn) error {
		for _, indexed := range []struct {
			key      []byte
			previous string
			current  string
		}{
			{storage.Key("account", "name", username), previousUsername, username},
			{storage.Key("account", "email", email), previousEmail, email},
		} {
			if indexed.previous == indexed.current {
				continue
			}
			if existing, getErr := transaction.Get(indexed.key); getErr == nil {
				value, copyErr := existing.ValueCopy(nil)
				if copyErr != nil {
					return copyErr
				}
				if string(value) != item.ID {
					return ErrConflict
				}
			} else if !errors.Is(getErr, badger.ErrKeyNotFound) {
				return getErr
			}
		}
		if previousUsername != username {
			if existing, getErr := transaction.Get(storage.Key("account", "profile-alias", username)); getErr == nil {
				value, copyErr := existing.ValueCopy(nil)
				if copyErr != nil {
					return copyErr
				}
				if string(value) != item.ID {
					return ErrConflict
				}
			} else if !errors.Is(getErr, badger.ErrKeyNotFound) {
				return getErr
			}
		}
		if err := transaction.Set(storage.Key("account", "object", item.ID), data); err != nil {
			return err
		}
		if previousUsername != username {
			if err := transaction.Delete(storage.Key("account", "name", previousUsername)); err != nil {
				return err
			}
			if err := transaction.Set(storage.Key("account", "name", username), []byte(item.ID)); err != nil {
				return err
			}
		}
		if previousEmail != email {
			if err := transaction.Delete(storage.Key("account", "email", previousEmail)); err != nil {
				return err
			}
			if err := transaction.Set(storage.Key("account", "email", email), []byte(item.ID)); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("update account: %w", err)
	}
	return nil
}

func (repository *Repository) ByUsername(username string) (Account, error) {
	return repository.byIndex(storage.Key("account", "name", strings.ToLower(strings.TrimSpace(username))))
}

func (repository *Repository) ByEmail(email string) (Account, error) {
	return repository.byIndex(storage.Key("account", "email", strings.ToLower(strings.TrimSpace(email))))
}

func (repository *Repository) AddProfileAlias(localPart, accountID string) error {
	if strings.TrimSpace(localPart) == "" || accountID == "" {
		return errors.New("profile alias and account id are required")
	}
	return repository.database.Set(storage.Key("account", "profile-alias", strings.ToLower(strings.TrimSpace(localPart))), []byte(accountID))
}

func (repository *Repository) ByProfileAlias(localPart string) (Account, error) {
	return repository.byIndex(storage.Key("account", "profile-alias", strings.ToLower(strings.TrimSpace(localPart))))
}

func (repository *Repository) byIndex(key []byte) (Account, error) {
	id, err := repository.database.Get(key)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return Account{}, storage.ErrNotFound
		}
		return Account{}, err
	}
	return repository.Account(string(id))
}

func (repository *Repository) read(key []byte) (Account, error) {
	data, err := repository.database.Get(key)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return Account{}, storage.ErrNotFound
		}
		return Account{}, err
	}
	var item Account
	if err := xml.Unmarshal(data, &item); err != nil {
		return Account{}, fmt.Errorf("decode account: %w", err)
	}
	return item, nil
}

func (repository *Repository) Delete(item Account) error {
	return repository.database.DB.Update(func(transaction *badger.Txn) error {
		for _, key := range [][]byte{
			storage.Key("account", "object", item.ID),
			storage.Key("account", "name", strings.ToLower(item.Username)),
			storage.Key("account", "email", strings.ToLower(item.Email)),
		} {
			if err := transaction.Delete(key); err != nil {
				return err
			}
		}
		return nil
	})
}
