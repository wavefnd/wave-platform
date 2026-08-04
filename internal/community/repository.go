package community

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

func (repository *Repository) UpsertSpace(space Space) error {
	if space.ID == "" || space.Slug == "" {
		return errors.New("community space id and slug are required")
	}
	data, err := xml.Marshal(space)
	if err != nil {
		return fmt.Errorf("encode community space: %w", err)
	}
	return repository.database.Set(storage.Key("community", "space", "object", space.ID), data)
}

func (repository *Repository) Spaces() ([]Space, error) {
	spaces := make([]Space, 0)
	err := repository.database.Scan(storage.Prefix("community", "space", "object"), func(_, value []byte) error {
		var space Space
		if err := xml.Unmarshal(value, &space); err != nil {
			return fmt.Errorf("decode community space: %w", err)
		}
		spaces = append(spaces, space)
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(spaces, func(left, right int) bool { return spaces[left].Name < spaces[right].Name })
	return spaces, nil
}

func (repository *Repository) Space(id string) (Space, error) {
	data, err := repository.database.Get(storage.Key("community", "space", "object", id))
	if err != nil {
		return Space{}, err
	}
	var space Space
	if err := xml.Unmarshal(data, &space); err != nil {
		return Space{}, fmt.Errorf("decode community space: %w", err)
	}
	return space, nil
}

func (repository *Repository) UpsertThread(thread Thread) error {
	if thread.ID == "" || thread.SpaceID == "" || thread.RootMessageID == "" {
		return errors.New("community thread id, space id, and root message id are required")
	}
	if _, err := repository.mail.Message(thread.RootMessageID); err != nil {
		return fmt.Errorf("community root message: %w", err)
	}
	data, err := xml.Marshal(thread)
	if err != nil {
		return fmt.Errorf("encode community thread: %w", err)
	}
	return repository.database.Set(storage.Key("community", "thread", "object", thread.ID), data)
}

func (repository *Repository) Thread(id string) (Thread, error) {
	data, err := repository.database.Get(storage.Key("community", "thread", "object", id))
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return Thread{}, storage.ErrNotFound
		}
		return Thread{}, err
	}
	var thread Thread
	if err := xml.Unmarshal(data, &thread); err != nil {
		return Thread{}, fmt.Errorf("decode community thread: %w", err)
	}
	return thread, nil
}

func (repository *Repository) Threads(spaceID string, limit int) ([]ThreadSummary, error) {
	return repository.QueryThreads(spaceID, "", "latest", limit, 0)
}

func (repository *Repository) QueryThreads(spaceID, query, order string, limit, offset int) ([]ThreadSummary, error) {
	threads := make([]ThreadSummary, 0)
	query = strings.ToLower(strings.TrimSpace(query))
	err := repository.database.Scan(storage.Prefix("community", "thread", "object"), func(_, value []byte) error {
		var thread Thread
		if err := xml.Unmarshal(value, &thread); err != nil {
			return fmt.Errorf("decode community thread: %w", err)
		}
		if spaceID != "" && thread.SpaceID != spaceID {
			return nil
		}
		summary, err := repository.summary(thread)
		if err != nil {
			return err
		}
		if query != "" {
			haystack := strings.ToLower(summary.Title + " " + summary.Excerpt + " " + strings.Join(summary.Tags, " "))
			if !strings.Contains(haystack, query) {
				return nil
			}
		}
		threads = append(threads, summary)
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.SliceStable(threads, func(left, right int) bool {
		if threads[left].Pinned != threads[right].Pinned {
			return threads[left].Pinned
		}
		switch order {
		case "active":
			return threads[left].LastActivityAt > threads[right].LastActivityAt
		case "top":
			if threads[left].Score != threads[right].Score {
				return threads[left].Score > threads[right].Score
			}
			return threads[left].LastActivityAt > threads[right].LastActivityAt
		default:
			return threads[left].CreatedAt > threads[right].CreatedAt
		}
	})
	if offset > 0 {
		if offset >= len(threads) {
			return []ThreadSummary{}, nil
		}
		threads = threads[offset:]
	}
	if limit > 0 && len(threads) > limit {
		threads = threads[:limit]
	}
	return threads, nil
}

func (repository *Repository) View(id string) (ThreadView, error) {
	return repository.ViewFor(id, "")
}

func (repository *Repository) ViewFor(id, viewerAccountID string) (ThreadView, error) {
	thread, err := repository.Thread(id)
	if err != nil {
		return ThreadView{}, err
	}
	root, err := repository.mail.Message(thread.RootMessageID)
	if err != nil {
		return ThreadView{}, err
	}
	messages, err := repository.mail.MessagesByThread(root.ThreadID)
	if err != nil {
		return ThreadView{}, err
	}
	rootView, err := repository.messageView(root)
	if err != nil {
		return ThreadView{}, err
	}
	score, err := repository.Score("thread", thread.ID)
	if err != nil {
		return ThreadView{}, err
	}
	views, err := repository.ViewCount(thread.ID)
	if err != nil {
		return ThreadView{}, err
	}
	viewerVote, subscribed := 0, false
	if viewerAccountID != "" {
		viewerVote, err = repository.VoteValue("thread", thread.ID, viewerAccountID)
		if err != nil {
			return ThreadView{}, err
		}
		subscribed, err = repository.IsSubscribed(thread.ID, viewerAccountID)
		if err != nil {
			return ThreadView{}, err
		}
	}
	rootView.Score, rootView.ViewerVote = score, viewerVote
	view := ThreadView{Thread: thread, Title: root.Subject, Root: rootView, Score: score, ViewCount: views,
		ViewerVote: viewerVote, Subscribed: subscribed}
	for _, message := range messages {
		if message.ID == thread.RootMessageID {
			continue
		}
		item, err := repository.messageView(message)
		if err != nil {
			return ThreadView{}, err
		}
		item.Score, err = repository.Score("message", message.ID)
		if err != nil {
			return ThreadView{}, err
		}
		if viewerAccountID != "" {
			item.ViewerVote, err = repository.VoteValue("message", message.ID, viewerAccountID)
			if err != nil {
				return ThreadView{}, err
			}
		}
		view.Replies = append(view.Replies, item)
	}
	return view, nil
}

func (repository *Repository) summary(thread Thread) (ThreadSummary, error) {
	root, err := repository.mail.Message(thread.RootMessageID)
	if err != nil {
		return ThreadSummary{}, err
	}
	body, err := repository.mail.Body(root)
	if err != nil {
		return ThreadSummary{}, err
	}
	messages, err := repository.mail.MessagesByThread(root.ThreadID)
	if err != nil {
		return ThreadSummary{}, err
	}
	replyCount := 0
	lastActivity := root.CreatedAt
	for _, message := range messages {
		if message.ID != thread.RootMessageID {
			replyCount++
		}
		if message.CreatedAt.After(lastActivity) {
			lastActivity = message.CreatedAt
		}
	}
	score, err := repository.Score("thread", thread.ID)
	if err != nil {
		return ThreadSummary{}, err
	}
	viewCount, err := repository.ViewCount(thread.ID)
	if err != nil {
		return ThreadSummary{}, err
	}
	return ThreadSummary{
		ID: thread.ID, SpaceID: thread.SpaceID, Title: root.Subject, Author: root.From,
		Excerpt: excerpt(body, 220), CreatedAt: root.CreatedAt.Format(time.RFC3339),
		ReplyCount: replyCount, ViewCount: viewCount, Score: score, LastActivityAt: lastActivity.Format(time.RFC3339),
		Tags: thread.Tags, Pinned: thread.Pinned, Locked: thread.Locked,
	}, nil
}

func (repository *Repository) messageView(message maildomain.Message) (MessageView, error) {
	body, err := repository.mail.Body(message)
	if err != nil {
		return MessageView{}, err
	}
	return MessageView{
		ID: message.ID, ParentMessageID: message.ParentMessageID, AuthorAccountID: message.AuthorAccountID, Author: message.From,
		CreatedAt: message.CreatedAt.Format(time.RFC3339), Body: body,
	}, nil
}

func (repository *Repository) SetVote(targetType, targetID, accountID string, value int) (int64, error) {
	if (targetType != "thread" && targetType != "message") || targetID == "" || accountID == "" || value < -1 || value > 1 {
		return 0, errors.New("invalid community vote")
	}
	voteKey := storage.Key("community", "vote", targetType, targetID, accountID)
	scoreKey := storage.Key("community", "score", targetType, targetID)
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
				return errors.New("invalid community score")
			}
			score = int64(binary.BigEndian.Uint64(data))
		} else if !errors.Is(err, badger.ErrKeyNotFound) {
			return err
		}
		score += int64(value - oldValue)
		encodedScore := make([]byte, 8)
		binary.BigEndian.PutUint64(encodedScore, uint64(score))
		if err := transaction.Set(scoreKey, encodedScore); err != nil {
			return err
		}
		if value == 0 {
			return transaction.Delete(voteKey)
		}
		data, err := xml.Marshal(Vote{TargetType: targetType, TargetID: targetID, AccountID: accountID, Value: value, UpdatedAt: time.Now().UTC()})
		if err != nil {
			return err
		}
		return transaction.Set(voteKey, data)
	})
	return score, err
}

func (repository *Repository) Score(targetType, targetID string) (int64, error) {
	data, err := repository.database.Get(storage.Key("community", "score", targetType, targetID))
	if errors.Is(err, storage.ErrNotFound) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	if len(data) != 8 {
		return 0, errors.New("invalid community score")
	}
	return int64(binary.BigEndian.Uint64(data)), nil
}

func (repository *Repository) VoteValue(targetType, targetID, accountID string) (int, error) {
	data, err := repository.database.Get(storage.Key("community", "vote", targetType, targetID, accountID))
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

func (repository *Repository) RecordView(threadID string) (uint64, error) {
	key := storage.Key("community", "view", threadID)
	var count uint64
	err := repository.database.DB.Update(func(transaction *badger.Txn) error {
		if item, err := transaction.Get(key); err == nil {
			data, copyErr := item.ValueCopy(nil)
			if copyErr != nil {
				return copyErr
			}
			if len(data) != 8 {
				return errors.New("invalid community view count")
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

func (repository *Repository) ViewCount(threadID string) (uint64, error) {
	data, err := repository.database.Get(storage.Key("community", "view", threadID))
	if errors.Is(err, storage.ErrNotFound) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	if len(data) != 8 {
		return 0, errors.New("invalid community view count")
	}
	return binary.BigEndian.Uint64(data), nil
}

func (repository *Repository) SetSubscribed(threadID, accountID string, subscribed bool) error {
	key := storage.Key("community", "subscription", threadID, accountID)
	if !subscribed {
		return repository.database.Delete(key)
	}
	data, err := xml.Marshal(Subscription{ThreadID: threadID, AccountID: accountID, CreatedAt: time.Now().UTC()})
	if err != nil {
		return err
	}
	return repository.database.Set(key, data)
}

func (repository *Repository) IsSubscribed(threadID, accountID string) (bool, error) {
	_, err := repository.database.Get(storage.Key("community", "subscription", threadID, accountID))
	if errors.Is(err, storage.ErrNotFound) {
		return false, nil
	}
	return err == nil, err
}

func (repository *Repository) Subscribers(threadID string) ([]string, error) {
	accounts := make([]string, 0)
	err := repository.database.Scan(storage.Prefix("community", "subscription", threadID), func(_, value []byte) error {
		var item Subscription
		if err := xml.Unmarshal(value, &item); err != nil {
			return err
		}
		accounts = append(accounts, item.AccountID)
		return nil
	})
	return accounts, err
}

func excerpt(value string, limit int) string {
	value = strings.Join(strings.Fields(value), " ")
	if len([]rune(value)) <= limit {
		return value
	}
	runes := []rune(value)
	return strings.TrimSpace(string(runes[:limit])) + "…"
}
