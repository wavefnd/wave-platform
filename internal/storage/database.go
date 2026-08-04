package storage

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/dgraph-io/badger/v4"
)

type Database struct {
	DB   *badger.DB
	Root string
}

func Open(root string) (*Database, error) {
	if root == "" {
		root = "./data"
	}

	root, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("resolve data root: %w", err)
	}

	directories := []string{
		root,
		filepath.Join(root, "badger"),
		filepath.Join(root, "blobs"),
		filepath.Join(root, "mail"),
		filepath.Join(root, "git"),
		filepath.Join(root, "git", "mirrors"),
		filepath.Join(root, "git", "temporary"),
		filepath.Join(root, "search"),
		filepath.Join(root, "temporary"),
		filepath.Join(root, "backups"),
	}

	for _, directory := range directories {
		if err := os.MkdirAll(directory, 0o750); err != nil {
			return nil, fmt.Errorf(
				"create data directory %q: %w",
				directory,
				err,
			)
		}
	}

	badgerPath := filepath.Join(root, "badger")

	if err := os.MkdirAll(badgerPath, 0o750); err != nil {
		return nil, fmt.Errorf("create data directory: %w", err)
	}

	options := badger.DefaultOptions(badgerPath).
		WithValueDir(badgerPath).
		WithLogger(nil)

	db, err := badger.Open(options)
	if err != nil {
		return nil, fmt.Errorf("open badger database: %w", err)
	}

	return &Database{
		DB:   db,
		Root: root,
	}, nil
}

func (database *Database) Health() error {
	if database == nil || database.DB == nil {
		return fmt.Errorf("database is not open")
	}

	if err := database.DB.View(func(transaction *badger.Txn) error {
		return nil
	}); err != nil {
		return fmt.Errorf("check database: %w", err)
	}

	return nil
}

func (database *Database) Close() error {
	if database == nil || database.DB == nil {
		return nil
	}

	return database.DB.Close()
}
