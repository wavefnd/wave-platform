package community

import (
	"encoding/xml"
	"fmt"

	maildomain "github.com/wavefnd/wave-platform/internal/mail"
	"github.com/wavefnd/wave-platform/internal/mailbox"
	"github.com/wavefnd/wave-platform/internal/storage"
)

// CleanupMailboxProjections removes legacy personal mailbox entries created for
// community posts and replies. Community content remains in the shared mail
// message store and is projected only through the Community domain.
func CleanupMailboxProjections(database *storage.Database) (int, error) {
	mailRepository := maildomain.NewRepository(database)
	communityMessageIDs := make(map[string]bool)
	if err := database.Scan(storage.Prefix("community", "thread", "object"), func(_, value []byte) error {
		var thread Thread
		if err := xml.Unmarshal(value, &thread); err != nil {
			return fmt.Errorf("decode community thread: %w", err)
		}
		root, err := mailRepository.Message(thread.RootMessageID)
		if err != nil {
			return err
		}
		messages, err := mailRepository.MessagesByThread(root.ThreadID)
		if err != nil {
			return err
		}
		for _, message := range messages {
			communityMessageIDs[message.ID] = true
		}
		return nil
	}); err != nil {
		return 0, err
	}

	keys := make([][]byte, 0)
	if err := database.Scan(storage.Prefix("mailbox", "entry"), func(key, value []byte) error {
		var entry mailbox.Entry
		if err := xml.Unmarshal(value, &entry); err != nil {
			return fmt.Errorf("decode mailbox entry: %w", err)
		}
		if communityMessageIDs[entry.MessageID] {
			keys = append(keys, key)
		}
		return nil
	}); err != nil {
		return 0, err
	}
	for _, key := range keys {
		if err := database.Delete(key); err != nil {
			return 0, err
		}
	}
	return len(keys), nil
}
