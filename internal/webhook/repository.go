package webhook

import (
	"encoding/xml"
	"fmt"
	"sort"
	"time"

	"github.com/wavefnd/wave-platform/internal/storage"
)

type Repository struct{ database *storage.Database }

func NewRepository(database *storage.Database) *Repository { return &Repository{database: database} }

func (repository *Repository) PutEndpoint(value Endpoint) error {
	data, err := xml.Marshal(value)
	if err != nil {
		return fmt.Errorf("encode webhook endpoint: %w", err)
	}
	return repository.database.Set(storage.Key("webhook", "endpoint", value.ID), data)
}

func (repository *Repository) Endpoint(id string) (Endpoint, error) {
	data, err := repository.database.Get(storage.Key("webhook", "endpoint", id))
	if err != nil {
		return Endpoint{}, err
	}
	var value Endpoint
	if err := xml.Unmarshal(data, &value); err != nil {
		return Endpoint{}, fmt.Errorf("decode webhook endpoint: %w", err)
	}
	return value, nil
}

func (repository *Repository) Endpoints() ([]Endpoint, error) {
	items := make([]Endpoint, 0)
	err := repository.database.Scan(storage.Prefix("webhook", "endpoint"), func(_, data []byte) error {
		var value Endpoint
		if err := xml.Unmarshal(data, &value); err != nil {
			return fmt.Errorf("decode webhook endpoint: %w", err)
		}
		items = append(items, value)
		return nil
	})
	sort.Slice(items, func(i, j int) bool { return items[i].CreatedAt.Before(items[j].CreatedAt) })
	return items, err
}

func (repository *Repository) DeleteEndpoint(id string) error {
	return repository.database.Delete(storage.Key("webhook", "endpoint", id))
}

func (repository *Repository) PutDelivery(value Delivery) error {
	data, err := xml.Marshal(value)
	if err != nil {
		return fmt.Errorf("encode webhook delivery: %w", err)
	}
	return repository.database.Set(storage.Key("webhook", "delivery", value.ID), data)
}

func (repository *Repository) Deliveries(limit int) ([]Delivery, error) {
	items := make([]Delivery, 0)
	err := repository.database.Scan(storage.Prefix("webhook", "delivery"), func(_, data []byte) error {
		var value Delivery
		if err := xml.Unmarshal(data, &value); err != nil {
			return fmt.Errorf("decode webhook delivery: %w", err)
		}
		items = append(items, value)
		return nil
	})
	sort.Slice(items, func(i, j int) bool { return items[i].CreatedAt.After(items[j].CreatedAt) })
	if limit > 0 && len(items) > limit {
		items = items[:limit]
	}
	return items, err
}

func (repository *Repository) Pending(now time.Time, limit int) ([]Delivery, error) {
	items, err := repository.Deliveries(0)
	if err != nil {
		return nil, err
	}
	result := make([]Delivery, 0)
	for _, item := range items {
		if (item.Status == "queued" || item.Status == "deferred") && (item.NextAttemptAt.IsZero() || !item.NextAttemptAt.After(now)) {
			result = append(result, item)
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].CreatedAt.Before(result[j].CreatedAt) })
	if limit > 0 && len(result) > limit {
		result = result[:limit]
	}
	return result, nil
}
