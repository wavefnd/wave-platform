package release

import (
	"encoding/xml"
	"errors"
	"fmt"
	"sort"

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

func (repository *Repository) Upsert(item Release) error {
	item.Content = ""
	data, err := xml.Marshal(item)
	if err != nil {
		return fmt.Errorf("encode language release: %w", err)
	}
	return repository.database.Set(storage.Key("release", "object", item.Slug), data)
}

func (repository *Repository) Releases(limit int) ([]Release, error) {
	items := make([]Release, 0)
	err := repository.database.Scan(storage.Prefix("release", "object"), func(_, value []byte) error {
		var item Release
		if err := xml.Unmarshal(value, &item); err != nil {
			return fmt.Errorf("decode language release: %w", err)
		}
		items = append(items, item)
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(items, func(left, right int) bool {
		return items[left].PublishedAt > items[right].PublishedAt
	})
	if limit > 0 && len(items) > limit {
		items = items[:limit]
	}
	return items, nil
}

func (repository *Repository) Release(slug string) (Release, error) {
	data, err := repository.database.Get(storage.Key("release", "object", slug))
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return Release{}, storage.ErrNotFound
		}
		return Release{}, err
	}

	var item Release
	if err := xml.Unmarshal(data, &item); err != nil {
		return Release{}, fmt.Errorf("decode language release: %w", err)
	}
	message, err := repository.mail.Message(item.MessageID)
	if err != nil {
		return Release{}, err
	}
	body, err := repository.mail.Body(message)
	if err != nil {
		return Release{}, err
	}
	item.Content = body
	return item, nil
}
