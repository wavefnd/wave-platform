package mail

import (
	"bytes"
	"encoding/base64"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"mime/quotedprintable"
	stdmail "net/mail"
	"net/textproto"
	"sort"
	"strings"
	"time"

	"github.com/wavefnd/wave-platform/internal/storage"
)

type Repository struct {
	database *storage.Database
}

func (repository *Repository) UpsertDelivery(delivery Delivery) error {
	if delivery.ID == "" || delivery.MessageID == "" || delivery.Recipient == "" {
		return errors.New("mail delivery id, message id, and recipient are required")
	}
	data, err := xml.Marshal(delivery)
	if err != nil {
		return fmt.Errorf("encode mail delivery: %w", err)
	}
	return repository.database.Set(storage.Key("mail", "delivery", "object", delivery.ID), data)
}

func (repository *Repository) Delivery(id string) (Delivery, error) {
	data, err := repository.database.Get(storage.Key("mail", "delivery", "object", id))
	if err != nil {
		return Delivery{}, err
	}
	var delivery Delivery
	if err := xml.Unmarshal(data, &delivery); err != nil {
		return Delivery{}, fmt.Errorf("decode mail delivery: %w", err)
	}
	return delivery, nil
}

func (repository *Repository) DeliveriesByMessage(messageID string) ([]Delivery, error) {
	return repository.deliveries(func(delivery Delivery) bool { return delivery.MessageID == messageID })
}

func (repository *Repository) PendingDeliveries(now time.Time, limit int) ([]Delivery, error) {
	items, err := repository.deliveries(func(delivery Delivery) bool {
		return (delivery.Status == "queued" || delivery.Status == "deferred") &&
			(delivery.NextAttemptAt.IsZero() || !delivery.NextAttemptAt.After(now))
	})
	if err != nil {
		return nil, err
	}
	if limit > 0 && len(items) > limit {
		items = items[:limit]
	}
	return items, nil
}

func (repository *Repository) deliveries(include func(Delivery) bool) ([]Delivery, error) {
	items := make([]Delivery, 0)
	err := repository.database.Scan(storage.Prefix("mail", "delivery", "object"), func(_, value []byte) error {
		var delivery Delivery
		if err := xml.Unmarshal(value, &delivery); err != nil {
			return fmt.Errorf("decode mail delivery: %w", err)
		}
		if include(delivery) {
			items = append(items, delivery)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(items, func(left, right int) bool {
		if items[left].CreatedAt.Equal(items[right].CreatedAt) {
			return items[left].ID < items[right].ID
		}
		return items[left].CreatedAt.Before(items[right].CreatedAt)
	})
	return items, nil
}

func NewRepository(database *storage.Database) *Repository {
	return &Repository{database: database}
}

func (repository *Repository) UpsertMessage(message Message, raw []byte) error {
	if message.ID == "" {
		return errors.New("mail message id is required")
	}
	if message.ThreadID == "" {
		message.ThreadID = message.ID
	}
	hash, err := repository.database.PutBlob(raw, "message/rfc822")
	if err != nil {
		return err
	}
	message.RawMessageKey = hash

	data, err := xml.Marshal(message)
	if err != nil {
		return fmt.Errorf("encode mail message: %w", err)
	}
	if err := repository.database.Set(storage.Key("mail", "message", message.ID), data); err != nil {
		return err
	}
	return repository.database.Set(storage.Key("mail", "thread", message.ThreadID, message.ID), []byte(message.ID))
}

func (repository *Repository) MessagesByThread(threadID string) ([]Message, error) {
	messages := make([]Message, 0)
	err := repository.database.Scan(storage.Prefix("mail", "thread", threadID), func(_, value []byte) error {
		message, err := repository.Message(string(value))
		if err != nil {
			return err
		}
		messages = append(messages, message)
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(messages, func(left, right int) bool {
		if messages[left].CreatedAt.Equal(messages[right].CreatedAt) {
			return messages[left].ID < messages[right].ID
		}
		return messages[left].CreatedAt.Before(messages[right].CreatedAt)
	})
	return messages, nil
}

func (repository *Repository) Message(id string) (Message, error) {
	data, err := repository.database.Get(storage.Key("mail", "message", id))
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return Message{}, storage.ErrNotFound
		}
		return Message{}, err
	}

	var message Message
	if err := xml.Unmarshal(data, &message); err != nil {
		return Message{}, fmt.Errorf("decode mail message: %w", err)
	}
	return message, nil
}

func (repository *Repository) RawMessage(message Message) ([]byte, error) {
	return repository.database.Blob(message.RawMessageKey)
}

func (repository *Repository) Body(message Message) (string, error) {
	raw, err := repository.RawMessage(message)
	if err != nil {
		return "", err
	}
	parsed, err := stdmail.ReadMessage(bytes.NewReader(raw))
	if err != nil {
		return fallbackBody(message.ID, raw)
	}
	body, err := decodeMIMEBody(textproto.MIMEHeader(parsed.Header), parsed.Body)
	if err != nil {
		return "", fmt.Errorf("decode mail message %q body: %w", message.ID, err)
	}
	return body, nil
}

func fallbackBody(messageID string, raw []byte) (string, error) {
	separator := []byte("\r\n\r\n")
	position := bytes.Index(raw, separator)
	if position < 0 {
		separator = []byte("\n\n")
		position = bytes.Index(raw, separator)
	}
	if position < 0 {
		return "", fmt.Errorf("mail message %q has no MIME body", messageID)
	}
	return string(raw[position+len(separator):]), nil
}

func decodeMIMEBody(header textproto.MIMEHeader, body io.Reader) (string, error) {
	mediaType, parameters, err := mime.ParseMediaType(header.Get("Content-Type"))
	if err != nil || mediaType == "" {
		mediaType = "text/plain"
	}
	if strings.HasPrefix(strings.ToLower(mediaType), "multipart/") {
		boundary := parameters["boundary"]
		if boundary == "" {
			return "", errors.New("multipart message has no boundary")
		}
		reader := multipart.NewReader(body, boundary)
		var fallback string
		for {
			part, partErr := reader.NextPart()
			if errors.Is(partErr, io.EOF) {
				break
			}
			if partErr != nil {
				return "", partErr
			}
			value, decodeErr := decodeMIMEBody(textproto.MIMEHeader(part.Header), part)
			_ = part.Close()
			if decodeErr != nil {
				continue
			}
			partType, _, _ := mime.ParseMediaType(part.Header.Get("Content-Type"))
			if strings.EqualFold(partType, "text/plain") {
				return value, nil
			}
			if fallback == "" && strings.HasPrefix(strings.ToLower(partType), "text/") {
				fallback = value
			}
		}
		return fallback, nil
	}

	var decoded io.Reader = body
	switch strings.ToLower(strings.TrimSpace(header.Get("Content-Transfer-Encoding"))) {
	case "base64":
		decoded = base64.NewDecoder(base64.StdEncoding, body)
	case "quoted-printable":
		decoded = quotedprintable.NewReader(body)
	}
	data, err := io.ReadAll(decoded)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
