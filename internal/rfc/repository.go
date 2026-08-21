package rfc

import (
	"encoding/binary"
	"encoding/xml"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/dgraph-io/badger/v4"
	"github.com/wavefnd/wave-platform/internal/storage"
)

type Repository struct{ database *storage.Database }

func NewRepository(database *storage.Database) *Repository { return &Repository{database: database} }

func (repository *Repository) NextNumber() (uint64, error) {
	var number uint64
	err := repository.database.DB.Update(func(transaction *badger.Txn) error {
		key := storage.Key("rfc", "state", "sequence")
		item, err := transaction.Get(key)
		if err == nil {
			value, copyErr := item.ValueCopy(nil)
			if copyErr != nil {
				return copyErr
			}
			if len(value) != 8 {
				return errors.New("invalid RFC sequence")
			}
			number = binary.BigEndian.Uint64(value)
		} else if !errors.Is(err, badger.ErrKeyNotFound) {
			return err
		}
		number++
		value := make([]byte, 8)
		binary.BigEndian.PutUint64(value, number)
		return transaction.Set(key, value)
	})
	if err != nil {
		return 0, fmt.Errorf("allocate RFC number: %w", err)
	}
	return number, nil
}

func (repository *Repository) Upsert(item Proposal) error {
	data, err := xml.Marshal(item)
	if err != nil {
		return fmt.Errorf("encode RFC: %w", err)
	}
	return repository.database.Set(storage.Key("rfc", "proposal", fmt.Sprintf("%020d", item.Number)), data)
}

func (repository *Repository) Proposal(number uint64) (Proposal, error) {
	data, err := repository.database.Get(storage.Key("rfc", "proposal", fmt.Sprintf("%020d", number)))
	if err != nil {
		return Proposal{}, err
	}
	var item Proposal
	if err := xml.Unmarshal(data, &item); err != nil {
		return Proposal{}, fmt.Errorf("decode RFC: %w", err)
	}
	comments, err := repository.Comments(number)
	if err != nil {
		return Proposal{}, err
	}
	item.Comments, item.CommentCount = comments, len(comments)
	return item, nil
}

func (repository *Repository) Proposals(query, status string) ([]Proposal, error) {
	query, status = strings.ToLower(strings.TrimSpace(query)), strings.ToLower(strings.TrimSpace(status))
	items := make([]Proposal, 0)
	err := repository.database.Scan(storage.Prefix("rfc", "proposal"), func(_, data []byte) error {
		var item Proposal
		if err := xml.Unmarshal(data, &item); err != nil {
			return fmt.Errorf("decode RFC: %w", err)
		}
		if status != "" && item.Status != status {
			return nil
		}
		if query != "" && !strings.Contains(strings.ToLower(item.Title+" "+item.Summary+" "+item.AuthorName), query) {
			return nil
		}
		comments, err := repository.Comments(item.Number)
		if err != nil {
			return err
		}
		item.Content, item.Comments, item.CommentCount = "", nil, len(comments)
		items = append(items, item)
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(items, func(left, right int) bool { return items[left].Number > items[right].Number })
	return items, nil
}

func (repository *Repository) AddComment(item Comment) error {
	data, err := xml.Marshal(item)
	if err != nil {
		return fmt.Errorf("encode RFC comment: %w", err)
	}
	return repository.database.Set(storage.Key("rfc", "comment", fmt.Sprintf("%020d", item.ProposalNumber), item.ID), data)
}

func (repository *Repository) Comments(number uint64) ([]Comment, error) {
	items := make([]Comment, 0)
	err := repository.database.Scan(storage.Prefix("rfc", "comment", fmt.Sprintf("%020d", number)), func(_, data []byte) error {
		var item Comment
		if err := xml.Unmarshal(data, &item); err != nil {
			return fmt.Errorf("decode RFC comment: %w", err)
		}
		items = append(items, item)
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(items, func(left, right int) bool { return items[left].CreatedAt.Before(items[right].CreatedAt) })
	return items, nil
}
