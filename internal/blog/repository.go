package blog

import (
	"encoding/xml"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/wavefnd/wave-platform/internal/storage"
)

type Repository struct{ database *storage.Database }

func NewRepository(database *storage.Database) *Repository { return &Repository{database: database} }

func (repository *Repository) Upsert(item Post) error {
	data, err := xml.Marshal(item)
	if err != nil {
		return fmt.Errorf("encode blog post: %w", err)
	}
	return repository.database.Set(storage.Key("blog", "post", item.Slug), data)
}

func (repository *Repository) Post(slug string, includeDrafts bool) (Post, error) {
	data, err := repository.database.Get(storage.Key("blog", "post", strings.ToLower(strings.TrimSpace(slug))))
	if err != nil {
		return Post{}, err
	}
	var item Post
	if err := xml.Unmarshal(data, &item); err != nil {
		return Post{}, fmt.Errorf("decode blog post: %w", err)
	}
	if !includeDrafts && item.Status != "published" {
		return Post{}, storage.ErrNotFound
	}
	item.Category = NormalizeCategory(item.Category)
	return item, nil
}

func (repository *Repository) Posts(includeDrafts bool, limit int) ([]Post, error) {
	return repository.PostsByCategory("", includeDrafts, limit)
}

func (repository *Repository) PostsByCategory(category string, includeDrafts bool, limit int) ([]Post, error) {
	category = strings.ToLower(strings.TrimSpace(category))
	items := make([]Post, 0)
	err := repository.database.Scan(storage.Prefix("blog", "post"), func(_, value []byte) error {
		var item Post
		if err := xml.Unmarshal(value, &item); err != nil {
			return fmt.Errorf("decode blog post: %w", err)
		}
		item.Category = NormalizeCategory(item.Category)
		if (!includeDrafts && item.Status != "published") || (category != "" && item.Category != category) {
			return nil
		}
		items = append(items, item)
		return nil
	})
	if err != nil && !errors.Is(err, storage.ErrNotFound) {
		return nil, err
	}
	sort.Slice(items, func(left, right int) bool {
		leftDate, rightDate := items[left].PublishedAt, items[right].PublishedAt
		if includeDrafts {
			leftDate, rightDate = items[left].UpdatedAt.Format(timeLayout), items[right].UpdatedAt.Format(timeLayout)
		}
		return leftDate > rightDate
	})
	if limit > 0 && len(items) > limit {
		items = items[:limit]
	}
	return items, nil
}

const timeLayout = "2006-01-02T15:04:05.999999999Z07:00"

func SummaryOf(item Post, includeStatus bool) Summary {
	status := ""
	if includeStatus {
		status = item.Status
	}
	return Summary{Slug: item.Slug, Category: NormalizeCategory(item.Category), Title: item.Title, Summary: item.Summary,
		Status: status, AuthorName: item.AuthorName, PublishedAt: item.PublishedAt, UpdatedAt: item.UpdatedAt.Format(timeLayout)}
}
