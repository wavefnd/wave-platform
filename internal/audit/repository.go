package audit

import (
	"encoding/binary"
	"encoding/xml"
	"errors"
	"fmt"

	"github.com/dgraph-io/badger/v4"
	"github.com/wavefnd/wave-platform/internal/storage"
)

type Repository struct{ database *storage.Database }

func NewRepository(database *storage.Database) *Repository { return &Repository{database: database} }

func (repository *Repository) Append(event Event) error {
	if event.ID == "" || event.Action == "" || event.ResourceID == "" {
		return errors.New("audit event id, action, and resource are required")
	}
	return repository.database.DB.Update(func(transaction *badger.Txn) error {
		sequenceKey := storage.Key("audit", "state", "sequence")
		var sequence uint64
		item, err := transaction.Get(sequenceKey)
		if err == nil {
			value, copyErr := item.ValueCopy(nil)
			if copyErr != nil {
				return copyErr
			}
			if len(value) != 8 {
				return errors.New("invalid audit sequence")
			}
			sequence = binary.BigEndian.Uint64(value)
		} else if !errors.Is(err, badger.ErrKeyNotFound) {
			return err
		}
		sequence++
		event.Sequence = sequence
		data, err := xml.Marshal(event)
		if err != nil {
			return fmt.Errorf("encode audit event: %w", err)
		}
		sequenceValue := make([]byte, 8)
		binary.BigEndian.PutUint64(sequenceValue, sequence)
		if err := transaction.Set(sequenceKey, sequenceValue); err != nil {
			return err
		}
		return transaction.Set(storage.Key("audit", "event", fmt.Sprintf("%020d", sequence)), data)
	})
}
