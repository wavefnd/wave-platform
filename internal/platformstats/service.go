package platformstats

import (
	"context"
	"encoding/xml"
	"fmt"
	"time"

	"github.com/wavefnd/wave-platform/internal/gitmirror"
	maildomain "github.com/wavefnd/wave-platform/internal/mail"
	"github.com/wavefnd/wave-platform/internal/storage"
)

type Snapshot struct {
	XMLName       xml.Name `xml:"https://wave-lang.dev/ns/platform/api/v1 stats"`
	Accounts      int      `xml:"accounts"`
	MessagesToday int      `xml:"messages-today"`
	GitMirrors    int      `xml:"git-mirrors"`
}

type Service struct {
	database  *storage.Database
	gitMirror *gitmirror.Service
	now       func() time.Time
}

func NewService(database *storage.Database, gitMirror *gitmirror.Service) *Service {
	return &Service{database: database, gitMirror: gitMirror, now: time.Now}
}

func (service *Service) Snapshot(ctx context.Context) (Snapshot, error) {
	var snapshot Snapshot
	if err := service.database.Scan(storage.Prefix("account", "object"), func(_, _ []byte) error {
		snapshot.Accounts++
		return nil
	}); err != nil {
		return Snapshot{}, fmt.Errorf("count accounts: %w", err)
	}

	today := service.now().In(time.Local)
	if err := service.database.Scan(storage.Prefix("mail", "message"), func(_, value []byte) error {
		var message maildomain.Message
		if err := xml.Unmarshal(value, &message); err != nil {
			return fmt.Errorf("decode mail message: %w", err)
		}
		occurredAt := message.ReceivedAt
		if occurredAt.IsZero() {
			occurredAt = message.CreatedAt
		}
		local := occurredAt.In(time.Local)
		if local.Year() == today.Year() && local.YearDay() == today.YearDay() {
			snapshot.MessagesToday++
		}
		return nil
	}); err != nil {
		return Snapshot{}, fmt.Errorf("count today's messages: %w", err)
	}

	if service.gitMirror != nil {
		repositories, err := service.gitMirror.Repositories(ctx)
		if err != nil {
			return Snapshot{}, fmt.Errorf("count Git mirrors: %w", err)
		}
		for _, repository := range repositories {
			if repository.Status == "ready" {
				snapshot.GitMirrors++
			}
		}
	}
	return snapshot, nil
}
