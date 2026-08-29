package notification

import (
	"encoding/xml"
	"fmt"
	"sort"

	"github.com/wavefnd/wave-platform/internal/storage"
)

type Repository struct{ database *storage.Database }

func NewRepository(database *storage.Database) *Repository { return &Repository{database: database} }

func (repository *Repository) Upsert(item Item) error {
	data, err := xml.Marshal(item)
	if err != nil {
		return fmt.Errorf("encode notification: %w", err)
	}
	return repository.database.Set(storage.Key("notification", "account", item.RecipientAccountID, item.ID), data)
}

func (repository *Repository) Item(accountID, notificationID string) (Item, error) {
	data, err := repository.database.Get(storage.Key("notification", "account", accountID, notificationID))
	if err != nil {
		return Item{}, err
	}
	var item Item
	if err := xml.Unmarshal(data, &item); err != nil {
		return Item{}, fmt.Errorf("decode notification: %w", err)
	}
	return item, nil
}

func (repository *Repository) List(accountID string, limit int) ([]Item, error) {
	items := make([]Item, 0)
	err := repository.database.Scan(storage.Prefix("notification", "account", accountID), func(_, value []byte) error {
		var item Item
		if err := xml.Unmarshal(value, &item); err != nil {
			return fmt.Errorf("decode notification: %w", err)
		}
		items = append(items, item)
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.SliceStable(items, func(left, right int) bool {
		if items[left].CreatedAt.Equal(items[right].CreatedAt) {
			return items[left].ID > items[right].ID
		}
		return items[left].CreatedAt.After(items[right].CreatedAt)
	})
	if limit > 0 && len(items) > limit {
		items = items[:limit]
	}
	return items, nil
}

func (repository *Repository) Delete(accountID, notificationID string) error {
	return repository.database.Delete(storage.Key("notification", "account", accountID, notificationID))
}
