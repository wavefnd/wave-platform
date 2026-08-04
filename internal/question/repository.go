package question

import (
	"encoding/binary"
	"encoding/xml"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/dgraph-io/badger/v4"
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

func (repository *Repository) Upsert(value Question) error {
	if value.ID == "" || value.RootMessageID == "" {
		return errors.New("question id and root message id are required")
	}
	if _, err := repository.mail.Message(value.RootMessageID); err != nil {
		return fmt.Errorf("question root message: %w", err)
	}
	data, err := xml.Marshal(value)
	if err != nil {
		return fmt.Errorf("encode question: %w", err)
	}
	return repository.database.Set(storage.Key("question", "object", value.ID), data)
}

func (repository *Repository) Question(id string) (Question, error) {
	data, err := repository.database.Get(storage.Key("question", "object", id))
	if err != nil {
		return Question{}, err
	}
	var value Question
	if err := xml.Unmarshal(data, &value); err != nil {
		return Question{}, fmt.Errorf("decode question: %w", err)
	}
	return value, nil
}

func (repository *Repository) Query(query, order, tag string, limit, offset int, viewerAccountID string) ([]Summary, error) {
	query = strings.ToLower(strings.TrimSpace(query))
	tag = strings.ToLower(strings.TrimSpace(tag))
	items := make([]Summary, 0)
	err := repository.database.Scan(storage.Prefix("question", "object"), func(_, data []byte) error {
		var value Question
		if err := xml.Unmarshal(data, &value); err != nil {
			return fmt.Errorf("decode question: %w", err)
		}
		summary, err := repository.summary(value, viewerAccountID)
		if err != nil {
			return err
		}
		if order == "unanswered" && summary.AnswerCount != 0 {
			return nil
		}
		if tag != "" && !contains(summary.Tags, tag) {
			return nil
		}
		if query != "" {
			haystack := strings.ToLower(summary.Title + " " + summary.Excerpt + " " + strings.Join(summary.Tags, " ") + " " + summary.WaveVersion)
			if !strings.Contains(haystack, query) {
				return nil
			}
		}
		items = append(items, summary)
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.SliceStable(items, func(left, right int) bool {
		if order == "active" {
			return items[left].LastActivityAt > items[right].LastActivityAt
		}
		return items[left].CreatedAt > items[right].CreatedAt
	})
	if offset > 0 {
		if offset >= len(items) {
			return []Summary{}, nil
		}
		items = items[offset:]
	}
	if limit > 0 && len(items) > limit {
		items = items[:limit]
	}
	return items, nil
}

func (repository *Repository) View(id, viewerAccountID string) (View, error) {
	value, err := repository.Question(id)
	if err != nil {
		return View{}, err
	}
	root, err := repository.mail.Message(value.RootMessageID)
	if err != nil {
		return View{}, err
	}
	messages, err := repository.mail.MessagesByThread(root.ThreadID)
	if err != nil {
		return View{}, err
	}
	rootView, err := repository.messageView(root, "question", value.ID, viewerAccountID)
	if err != nil {
		return View{}, err
	}
	result := View{Question: value, Title: root.Subject, Root: rootView, Score: rootView.Score,
		ViewerVote: rootView.ViewerVote}
	result.ViewCount, err = repository.ViewCount(id)
	if err != nil {
		return View{}, err
	}
	for _, message := range messages {
		if message.ID == value.RootMessageID {
			continue
		}
		answer, err := repository.messageView(message, "answer", message.ID, viewerAccountID)
		if err != nil {
			return View{}, err
		}
		answer.Accepted = message.ID == value.AcceptedMessageID
		result.Answers = append(result.Answers, answer)
	}
	sort.SliceStable(result.Answers, func(left, right int) bool {
		if result.Answers[left].Accepted != result.Answers[right].Accepted {
			return result.Answers[left].Accepted
		}
		if result.Answers[left].Score != result.Answers[right].Score {
			return result.Answers[left].Score > result.Answers[right].Score
		}
		return result.Answers[left].CreatedAt < result.Answers[right].CreatedAt
	})
	return result, nil
}

func (repository *Repository) summary(value Question, viewerAccountID string) (Summary, error) {
	root, err := repository.mail.Message(value.RootMessageID)
	if err != nil {
		return Summary{}, err
	}
	body, err := repository.mail.Body(root)
	if err != nil {
		return Summary{}, err
	}
	messages, err := repository.mail.MessagesByThread(root.ThreadID)
	if err != nil {
		return Summary{}, err
	}
	lastActivity := root.CreatedAt
	answerCount := 0
	for _, message := range messages {
		if message.ID != value.RootMessageID {
			answerCount++
		}
		if message.CreatedAt.After(lastActivity) {
			lastActivity = message.CreatedAt
		}
	}
	score, err := repository.Score("question", value.ID)
	if err != nil {
		return Summary{}, err
	}
	viewerVote := 0
	if viewerAccountID != "" {
		viewerVote, err = repository.VoteValue("question", value.ID, viewerAccountID)
		if err != nil {
			return Summary{}, err
		}
	}
	views, err := repository.ViewCount(value.ID)
	if err != nil {
		return Summary{}, err
	}
	return Summary{ID: value.ID, Title: root.Subject, Excerpt: excerpt(body, 220), Author: root.From,
		AuthorAccountID: root.AuthorAccountID, CreatedAt: root.CreatedAt.Format(time.RFC3339),
		LastActivityAt: lastActivity.Format(time.RFC3339), Tags: value.Tags, WaveVersion: value.WaveVersion,
		Platform: value.Platform, Status: value.Status, Score: score, ViewerVote: viewerVote,
		AnswerCount: answerCount, ViewCount: views, Accepted: value.AcceptedMessageID != ""}, nil
}

func (repository *Repository) messageView(message maildomain.Message, targetType, targetID, viewerAccountID string) (MessageView, error) {
	body, err := repository.mail.Body(message)
	if err != nil {
		return MessageView{}, err
	}
	score, err := repository.Score(targetType, targetID)
	if err != nil {
		return MessageView{}, err
	}
	viewerVote := 0
	if viewerAccountID != "" {
		viewerVote, err = repository.VoteValue(targetType, targetID, viewerAccountID)
		if err != nil {
			return MessageView{}, err
		}
	}
	return MessageView{ID: message.ID, ParentMessageID: message.ParentMessageID,
		AuthorAccountID: message.AuthorAccountID, Author: message.From, CreatedAt: message.CreatedAt.Format(time.RFC3339),
		Body: body, Score: score, ViewerVote: viewerVote}, nil
}

func (repository *Repository) SetVote(targetType, targetID, accountID string, value int) (int64, error) {
	if (targetType != "question" && targetType != "answer") || targetID == "" || accountID == "" || value < -1 || value > 1 {
		return 0, errors.New("invalid question vote")
	}
	voteKey := storage.Key("question", "vote", targetType, targetID, accountID)
	scoreKey := storage.Key("question", "score", targetType, targetID)
	var score int64
	err := repository.database.DB.Update(func(transaction *badger.Txn) error {
		oldValue := 0
		if item, err := transaction.Get(voteKey); err == nil {
			data, copyErr := item.ValueCopy(nil)
			if copyErr != nil {
				return copyErr
			}
			var old Vote
			if err := xml.Unmarshal(data, &old); err != nil {
				return err
			}
			oldValue = old.Value
		} else if !errors.Is(err, badger.ErrKeyNotFound) {
			return err
		}
		if item, err := transaction.Get(scoreKey); err == nil {
			data, copyErr := item.ValueCopy(nil)
			if copyErr != nil {
				return copyErr
			}
			if len(data) != 8 {
				return errors.New("invalid question score")
			}
			score = int64(binary.BigEndian.Uint64(data))
		} else if !errors.Is(err, badger.ErrKeyNotFound) {
			return err
		}
		score += int64(value - oldValue)
		encoded := make([]byte, 8)
		binary.BigEndian.PutUint64(encoded, uint64(score))
		if err := transaction.Set(scoreKey, encoded); err != nil {
			return err
		}
		if value == 0 {
			return transaction.Delete(voteKey)
		}
		data, err := xml.Marshal(Vote{TargetType: targetType, TargetID: targetID, AccountID: accountID,
			Value: value, UpdatedAt: time.Now().UTC()})
		if err != nil {
			return err
		}
		return transaction.Set(voteKey, data)
	})
	return score, err
}

func (repository *Repository) Score(targetType, targetID string) (int64, error) {
	data, err := repository.database.Get(storage.Key("question", "score", targetType, targetID))
	if errors.Is(err, storage.ErrNotFound) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	if len(data) != 8 {
		return 0, errors.New("invalid question score")
	}
	return int64(binary.BigEndian.Uint64(data)), nil
}

func (repository *Repository) VoteValue(targetType, targetID, accountID string) (int, error) {
	data, err := repository.database.Get(storage.Key("question", "vote", targetType, targetID, accountID))
	if errors.Is(err, storage.ErrNotFound) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	var vote Vote
	if err := xml.Unmarshal(data, &vote); err != nil {
		return 0, err
	}
	return vote.Value, nil
}

func (repository *Repository) RecordView(questionID string) (uint64, error) {
	key := storage.Key("question", "view", questionID)
	var count uint64
	err := repository.database.DB.Update(func(transaction *badger.Txn) error {
		if item, err := transaction.Get(key); err == nil {
			data, copyErr := item.ValueCopy(nil)
			if copyErr != nil {
				return copyErr
			}
			if len(data) != 8 {
				return errors.New("invalid question view count")
			}
			count = binary.BigEndian.Uint64(data)
		} else if !errors.Is(err, badger.ErrKeyNotFound) {
			return err
		}
		count++
		data := make([]byte, 8)
		binary.BigEndian.PutUint64(data, count)
		return transaction.Set(key, data)
	})
	return count, err
}

func (repository *Repository) ViewCount(questionID string) (uint64, error) {
	data, err := repository.database.Get(storage.Key("question", "view", questionID))
	if errors.Is(err, storage.ErrNotFound) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	if len(data) != 8 {
		return 0, errors.New("invalid question view count")
	}
	return binary.BigEndian.Uint64(data), nil
}

func excerpt(value string, limit int) string {
	value = strings.Join(strings.Fields(value), " ")
	runes := []rune(value)
	if len(runes) <= limit {
		return value
	}
	return string(runes[:limit]) + "…"
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
