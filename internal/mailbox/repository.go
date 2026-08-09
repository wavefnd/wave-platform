package mailbox

import (
	"encoding/xml"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/dgraph-io/badger/v4"
	"github.com/wavefnd/wave-platform/internal/account"
	maildomain "github.com/wavefnd/wave-platform/internal/mail"
	"github.com/wavefnd/wave-platform/internal/storage"
)

type Repository struct {
	database *storage.Database
	mail     *maildomain.Repository
}

func NewRepository(database *storage.Database) *Repository {
	return &Repository{database: database, mail: maildomain.NewRepository(database)}
}

func (repository *Repository) UpsertMailbox(mailbox Mailbox) error {
	if mailbox.ID == "" || mailbox.AccountID == "" {
		return errors.New("mailbox id and account id are required")
	}
	data, err := xml.Marshal(mailbox)
	if err != nil {
		return fmt.Errorf("encode mailbox: %w", err)
	}
	return repository.database.DB.Update(func(transaction *badger.Txn) error {
		addressKey := storage.Key("mailbox", "address", strings.ToLower(mailbox.Address))
		if existing, getErr := transaction.Get(addressKey); getErr == nil {
			value, copyErr := existing.ValueCopy(nil)
			if copyErr != nil {
				return copyErr
			}
			if string(value) != mailbox.ID {
				return account.ErrConflict
			}
		} else if !errors.Is(getErr, badger.ErrKeyNotFound) {
			return getErr
		}
		if err := transaction.Set(storage.Key("mailbox", "object", mailbox.ID), data); err != nil {
			return err
		}
		if err := transaction.Set(storage.Key("mailbox", "account", mailbox.AccountID), []byte(mailbox.ID)); err != nil {
			return err
		}
		return transaction.Set(addressKey, []byte(mailbox.ID))
	})
}

func (repository *Repository) AddAddress(mailboxID, address string) error {
	if mailboxID == "" || strings.TrimSpace(address) == "" {
		return errors.New("mailbox id and address are required")
	}
	key := storage.Key("mailbox", "address", strings.ToLower(strings.TrimSpace(address)))
	return repository.database.DB.Update(func(transaction *badger.Txn) error {
		if existing, err := transaction.Get(key); err == nil {
			value, copyErr := existing.ValueCopy(nil)
			if copyErr != nil {
				return copyErr
			}
			if string(value) != mailboxID {
				return account.ErrConflict
			}
			return nil
		} else if !errors.Is(err, badger.ErrKeyNotFound) {
			return err
		}
		return transaction.Set(key, []byte(mailboxID))
	})
}

func (repository *Repository) UpdateAddress(mailbox Mailbox, previousAddress string) error {
	data, err := xml.Marshal(mailbox)
	if err != nil {
		return fmt.Errorf("encode mailbox: %w", err)
	}
	return repository.database.DB.Update(func(transaction *badger.Txn) error {
		newKey := storage.Key("mailbox", "address", strings.ToLower(mailbox.Address))
		if existing, getErr := transaction.Get(newKey); getErr == nil {
			value, copyErr := existing.ValueCopy(nil)
			if copyErr != nil {
				return copyErr
			}
			if string(value) != mailbox.ID {
				return account.ErrConflict
			}
		} else if !errors.Is(getErr, badger.ErrKeyNotFound) {
			return getErr
		}
		if err := transaction.Set(storage.Key("mailbox", "object", mailbox.ID), data); err != nil {
			return err
		}
		if !strings.EqualFold(previousAddress, mailbox.Address) {
			if err := transaction.Delete(storage.Key("mailbox", "address", strings.ToLower(previousAddress))); err != nil {
				return err
			}
		}
		return transaction.Set(newKey, []byte(mailbox.ID))
	})
}

func (repository *Repository) MailboxByAccount(accountID string) (Mailbox, error) {
	id, err := repository.database.Get(storage.Key("mailbox", "account", accountID))
	if err != nil {
		return Mailbox{}, err
	}
	data, err := repository.database.Get(storage.Key("mailbox", "object", string(id)))
	if err != nil {
		return Mailbox{}, err
	}
	var item Mailbox
	if err := xml.Unmarshal(data, &item); err != nil {
		return Mailbox{}, fmt.Errorf("decode mailbox: %w", err)
	}
	return item, nil
}

func (repository *Repository) MailboxByAddress(address string) (Mailbox, error) {
	id, err := repository.database.Get(storage.Key("mailbox", "address", strings.ToLower(strings.TrimSpace(address))))
	if err != nil {
		return Mailbox{}, err
	}
	data, err := repository.database.Get(storage.Key("mailbox", "object", string(id)))
	if err != nil {
		return Mailbox{}, err
	}
	var item Mailbox
	if err := xml.Unmarshal(data, &item); err != nil {
		return Mailbox{}, fmt.Errorf("decode mailbox: %w", err)
	}
	return item, nil
}

func (repository *Repository) DeleteMailbox(mailbox Mailbox) error {
	for _, key := range [][]byte{
		storage.Key("mailbox", "object", mailbox.ID),
		storage.Key("mailbox", "account", mailbox.AccountID),
		storage.Key("mailbox", "address", mailbox.Address),
	} {
		if err := repository.database.Delete(key); err != nil {
			return err
		}
	}
	return nil
}

func (repository *Repository) AddEntry(entry Entry) error {
	if entry.ID == "" || entry.MailboxID == "" || entry.MessageID == "" {
		return errors.New("mailbox entry id, mailbox id, and message id are required")
	}
	if _, err := repository.mail.Message(entry.MessageID); err != nil {
		return fmt.Errorf("mailbox entry message: %w", err)
	}
	data, err := xml.Marshal(entry)
	if err != nil {
		return fmt.Errorf("encode mailbox entry: %w", err)
	}
	return repository.database.Set(storage.Key("mailbox", "entry", entry.MailboxID, entry.ID), data)
}

func (repository *Repository) Entry(mailboxID, entryID string) (Entry, error) {
	data, err := repository.database.Get(storage.Key("mailbox", "entry", mailboxID, entryID))
	if err != nil {
		return Entry{}, err
	}
	var entry Entry
	if err := xml.Unmarshal(data, &entry); err != nil {
		return Entry{}, fmt.Errorf("decode mailbox entry: %w", err)
	}
	return entry, nil
}

func (repository *Repository) UpdateEntry(entry Entry) error {
	if entry.ID == "" || entry.MailboxID == "" || entry.MessageID == "" {
		return errors.New("mailbox entry id, mailbox id, and message id are required")
	}
	data, err := xml.Marshal(entry)
	if err != nil {
		return fmt.Errorf("encode mailbox entry: %w", err)
	}
	return repository.database.Set(storage.Key("mailbox", "entry", entry.MailboxID, entry.ID), data)
}

func (repository *Repository) Entries(mailboxID, folder string) ([]Entry, error) {
	entries := make([]Entry, 0)
	err := repository.database.Scan(storage.Prefix("mailbox", "entry", mailboxID), func(_, value []byte) error {
		var entry Entry
		if err := xml.Unmarshal(value, &entry); err != nil {
			return fmt.Errorf("decode mailbox entry: %w", err)
		}
		if folder == "" || entry.Folder == folder {
			entries = append(entries, entry)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(entries, func(left, right int) bool { return entries[left].CreatedAt.After(entries[right].CreatedAt) })
	return entries, nil
}

func (repository *Repository) Message(entry Entry) (maildomain.Message, error) {
	message, err := repository.mail.Message(entry.MessageID)
	if errors.Is(err, storage.ErrNotFound) {
		return maildomain.Message{}, storage.ErrNotFound
	}
	return message, err
}
