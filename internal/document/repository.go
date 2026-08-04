package document

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

func (repository *Repository) UpsertDocument(value Document) error {
	if value.ID == "" || value.Path == "" || (value.Locale != "en" && value.Locale != "ko") {
		return errors.New("document id, path, and supported locale are required")
	}
	data, err := xml.Marshal(value)
	if err != nil {
		return fmt.Errorf("encode document: %w", err)
	}
	if err := repository.database.Set(storage.Key("document", "object", value.ID), data); err != nil {
		return err
	}
	return repository.database.Set(storage.Key("document", "path", value.Locale, value.Path), []byte(value.ID))
}

func (repository *Repository) PutRevision(value Revision) error {
	if value.ID == "" || value.DocumentID == "" || len(value.ContentXML) == 0 {
		return errors.New("revision id, document id, and content are required")
	}
	var content Content
	if err := xml.Unmarshal(value.ContentXML, &content); err != nil {
		return fmt.Errorf("validate document content: %w", err)
	}
	data, err := xml.Marshal(value)
	if err != nil {
		return fmt.Errorf("encode document revision: %w", err)
	}
	return repository.database.Set(storage.Key("document", "revision", value.DocumentID, value.ID), data)
}

func (repository *Repository) Document(id string) (Document, error) {
	data, err := repository.database.Get(storage.Key("document", "object", id))
	if err != nil {
		return Document{}, err
	}
	var value Document
	if err := xml.Unmarshal(data, &value); err != nil {
		return Document{}, fmt.Errorf("decode document: %w", err)
	}
	return value, nil
}

func (repository *Repository) Revision(documentID, revisionID string) (Revision, error) {
	data, err := repository.database.Get(storage.Key("document", "revision", documentID, revisionID))
	if err != nil {
		return Revision{}, err
	}
	var value Revision
	if err := xml.Unmarshal(data, &value); err != nil {
		return Revision{}, fmt.Errorf("decode document revision: %w", err)
	}
	return value, nil
}

func (repository *Repository) Summaries(locale string) ([]Summary, error) {
	items := make([]Document, 0)
	err := repository.database.Scan(storage.Prefix("document", "object"), func(_, data []byte) error {
		var value Document
		if err := xml.Unmarshal(data, &value); err != nil {
			return fmt.Errorf("decode document: %w", err)
		}
		if value.Locale == locale && value.Status == "published" {
			items = append(items, value)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.SliceStable(items, func(left, right int) bool {
		if items[left].GroupOrder != items[right].GroupOrder {
			return items[left].GroupOrder < items[right].GroupOrder
		}
		if items[left].Order != items[right].Order {
			return items[left].Order < items[right].Order
		}
		return items[left].Path < items[right].Path
	})
	result := make([]Summary, 0, len(items))
	for _, value := range items {
		result = append(result, summaryOf(value))
	}
	return result, nil
}

func (repository *Repository) Published(locale, path string) (View, error) {
	path = strings.Trim(strings.TrimSpace(path), "/")
	id, err := repository.database.Get(storage.Key("document", "path", locale, path))
	if err != nil {
		return View{}, err
	}
	value, err := repository.Document(string(id))
	if err != nil {
		return View{}, err
	}
	if value.Status != "published" || value.PublishedRevisionID == "" {
		return View{}, storage.ErrNotFound
	}
	revision, err := repository.Revision(value.ID, value.PublishedRevisionID)
	if err != nil {
		return View{}, err
	}
	var content Content
	if err := xml.Unmarshal(revision.ContentXML, &content); err != nil {
		return View{}, fmt.Errorf("decode document content: %w", err)
	}
	return View{Summary: summaryOf(value), SourceRevision: value.SourceRevision,
		UpdatedAt: value.UpdatedAt.Format("2006-01-02"), Markdown: content.Markdown, Blocks: content.Blocks}, nil
}

func summaryOf(value Document) Summary {
	return Summary{ID: value.ID, Path: value.Path, Locale: value.Locale, Group: value.Group,
		Order: value.Order, Title: value.Title, Summary: value.Summary, Version: value.Version}
}
