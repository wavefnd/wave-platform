package session

import (
	"encoding/xml"
	"errors"
	"fmt"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/wavefnd/wave-platform/internal/storage"
)

type Repository struct{ database *storage.Database }

func NewRepository(database *storage.Database) *Repository { return &Repository{database: database} }

func (repository *Repository) Put(item Session) error {
	if item.ID == "" || item.AccountID == "" {
		return errors.New("session id and account id are required")
	}
	data, err := xml.Marshal(item)
	if err != nil {
		return fmt.Errorf("encode session: %w", err)
	}
	ttl := time.Until(item.ExpiresAt)
	if ttl <= 0 {
		return errors.New("session expiry must be in the future")
	}
	return repository.database.DB.Update(func(transaction *badger.Txn) error {
		entry := badger.NewEntry(storage.Key("session", "object", item.ID), data).WithTTL(ttl)
		if err := transaction.SetEntry(entry); err != nil {
			return err
		}
		index := badger.NewEntry(storage.Key("session", "account", item.AccountID, item.ID), []byte(item.ID)).WithTTL(ttl)
		return transaction.SetEntry(index)
	})
}

func (repository *Repository) Session(id string) (Session, error) {
	data, err := repository.database.Get(storage.Key("session", "object", id))
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return Session{}, storage.ErrNotFound
		}
		return Session{}, err
	}
	var item Session
	if err := xml.Unmarshal(data, &item); err != nil {
		return Session{}, fmt.Errorf("decode session: %w", err)
	}
	return item, nil
}

func (repository *Repository) Delete(item Session) error {
	return repository.database.DB.Update(func(transaction *badger.Txn) error {
		if err := transaction.Delete(storage.Key("session", "object", item.ID)); err != nil {
			return err
		}
		return transaction.Delete(storage.Key("session", "account", item.AccountID, item.ID))
	})
}

func (repository *Repository) DeleteByAccount(accountID string) error {
	var sessionIDs []string
	err := repository.database.Scan(storage.Prefix("session", "account", accountID), func(_, value []byte) error {
		sessionIDs = append(sessionIDs, string(value))
		return nil
	})
	if err != nil {
		return err
	}
	for _, sessionID := range sessionIDs {
		item, err := repository.Session(sessionID)
		if errors.Is(err, storage.ErrNotFound) {
			continue
		}
		if err != nil {
			return err
		}
		if err := repository.Delete(item); err != nil {
			return err
		}
	}
	return nil
}
