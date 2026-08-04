package storage

import (
	"errors"
	"fmt"

	"github.com/dgraph-io/badger/v4"
)

var ErrNotFound = errors.New("record not found")

func (database *Database) Get(key []byte) ([]byte, error) {
	var result []byte

	err := database.DB.View(func(transaction *badger.Txn) error {
		item, err := transaction.Get(key)
		if errors.Is(err, badger.ErrKeyNotFound) {
			return ErrNotFound
		}

		if err != nil {
			return err
		}

		result, err = item.ValueCopy(nil)
		return err
	})

	if err != nil {
		return nil, fmt.Errorf("read key %q: %w", string(key), err)
	}

	return result, nil
}

func (database *Database) Set(key, value []byte) error {
	err := database.DB.Update(func(transaction *badger.Txn) error {
		return transaction.Set(key, value)
	})

	if err != nil {
		return fmt.Errorf("write key %q: %w", string(key), err)
	}

	return nil
}

func (database *Database) Delete(key []byte) error {
	err := database.DB.Update(func(transaction *badger.Txn) error {
		return transaction.Delete(key)
	})

	if err != nil {
		return fmt.Errorf("delete key %q: %w", string(key), err)
	}

	return nil
}

func (database *Database) Scan(prefix []byte, visit func(key, value []byte) error) error {
	err := database.DB.View(func(transaction *badger.Txn) error {
		iterator := transaction.NewIterator(badger.DefaultIteratorOptions)
		defer iterator.Close()

		for iterator.Seek(prefix); iterator.ValidForPrefix(prefix); iterator.Next() {
			item := iterator.Item()
			key := item.KeyCopy(nil)
			value, err := item.ValueCopy(nil)
			if err != nil {
				return err
			}
			if err := visit(key, value); err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("scan prefix %q: %w", string(prefix), err)
	}
	return nil
}
