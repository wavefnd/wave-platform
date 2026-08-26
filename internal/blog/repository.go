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
	item.CommentPolicy = NormalizeCommentPolicy(item.Category, item.CommentPolicy)
	if item.Category == "roadmap" {
		item.Summary = SummaryFromContent(item.Content)
	}
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
		item.CommentPolicy = NormalizeCommentPolicy(item.Category, item.CommentPolicy)
		if item.Category == "roadmap" {
			item.Summary = SummaryFromContent(item.Content)
		}
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
		if category == "roadmap" {
			if items[left].RoadmapOrder != items[right].RoadmapOrder {
				return items[left].RoadmapOrder < items[right].RoadmapOrder
			}
			return items[left].Slug < items[right].Slug
		}
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

func (repository *Repository) AddComment(item Comment) error {
	data, err := xml.Marshal(item)
	if err != nil {
		return fmt.Errorf("encode blog comment: %w", err)
	}
	return repository.database.Set(storage.Key("blog", "comment", item.PostSlug, item.ID), data)
}

func (repository *Repository) Comment(slug, id string) (Comment, error) {
	data, err := repository.database.Get(storage.Key("blog", "comment", strings.ToLower(strings.TrimSpace(slug)), strings.TrimSpace(id)))
	if err != nil {
		return Comment{}, err
	}
	var item Comment
	if err := xml.Unmarshal(data, &item); err != nil {
		return Comment{}, fmt.Errorf("decode blog comment: %w", err)
	}
	return item, nil
}

func (repository *Repository) Comments(slug string, includeHidden bool) ([]Comment, error) {
	items := make([]Comment, 0)
	err := repository.database.Scan(storage.Prefix("blog", "comment", strings.ToLower(strings.TrimSpace(slug))), func(_, data []byte) error {
		var item Comment
		if err := xml.Unmarshal(data, &item); err != nil {
			return fmt.Errorf("decode blog comment: %w", err)
		}
		if !includeHidden && item.Status != "visible" {
			return nil
		}
		items = append(items, item)
		return nil
	})
	if err != nil && !errors.Is(err, storage.ErrNotFound) {
		return nil, err
	}
	sort.Slice(items, func(left, right int) bool { return items[left].CreatedAt.Before(items[right].CreatedAt) })
	return items, nil
}

const timeLayout = "2006-01-02T15:04:05.999999999Z07:00"

func SummaryOf(item Post, includeStatus bool) Summary {
	status := ""
	if includeStatus {
		status = item.Status
	}
	summary := item.Summary
	if NormalizeCategory(item.Category) == "roadmap" {
		summary = SummaryFromContent(item.Content)
	}
	return Summary{Slug: item.Slug, Category: NormalizeCategory(item.Category), RoadmapStatus: item.RoadmapStatus,
		RoadmapOrder: item.RoadmapOrder, TargetDate: item.TargetDate, Title: item.Title, Summary: summary,
		Status: status, CommentPolicy: NormalizeCommentPolicy(item.Category, item.CommentPolicy), AuthorName: item.AuthorName,
		PublishedAt: item.PublishedAt, UpdatedAt: item.UpdatedAt.Format(timeLayout)}
}
