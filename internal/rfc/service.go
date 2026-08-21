package rfc

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/wavefnd/wave-platform/internal/account"
	"github.com/wavefnd/wave-platform/internal/audit"
	"github.com/wavefnd/wave-platform/internal/identifier"
	"github.com/wavefnd/wave-platform/internal/permission"
	"github.com/wavefnd/wave-platform/internal/storage"
)

var (
	ErrForbidden       = errors.New("RFC access is forbidden")
	ErrInvalidProposal = errors.New("invalid RFC proposal")
	ErrInvalidStatus   = errors.New("invalid RFC status")
	ErrInvalidComment  = errors.New("invalid RFC comment")
)

var validStatuses = map[string]bool{
	"draft": true, "discussion": true, "accepted": true, "rejected": true,
	"implementing": true, "completed": true, "withdrawn": true,
}

type Service struct {
	repository  *Repository
	accounts    *account.Repository
	permissions *permission.Repository
	audit       *audit.Repository
	now         func() time.Time
}

func NewService(database *storage.Database) *Service {
	return &Service{repository: NewRepository(database), accounts: account.NewRepository(database),
		permissions: permission.NewRepository(database), audit: audit.NewRepository(database), now: time.Now}
}

func (service *Service) Repository() *Repository { return service.repository }

func (service *Service) CanMaintain(accountID string) (bool, error) {
	owner, err := service.permissions.HasRole(accountID, "platform-owner")
	if err != nil || owner {
		return owner, err
	}
	return service.permissions.HasRole(accountID, "rfc-maintainer")
}

func (service *Service) Create(actorID string, input ProposalInput) (Proposal, error) {
	author, err := service.accounts.Account(actorID)
	if err != nil {
		return Proposal{}, err
	}
	title, content, err := validateProposal(input)
	if err != nil {
		return Proposal{}, err
	}
	number, err := service.repository.NextNumber()
	if err != nil {
		return Proposal{}, err
	}
	now := service.now().UTC()
	item := Proposal{Number: number, Title: title, Summary: summarize(content, 180), Content: content, Status: "draft",
		AuthorAccountID: author.ID, AuthorName: author.DisplayName, CreatedAt: now, UpdatedAt: now}
	if err := service.repository.Upsert(item); err != nil {
		return Proposal{}, err
	}
	if err := service.appendAudit(actorID, number, "rfc.create"); err != nil {
		return Proposal{}, err
	}
	return item, nil
}

func (service *Service) Update(actorID string, number uint64, input ProposalInput) (Proposal, error) {
	item, err := service.repository.Proposal(number)
	if err != nil {
		return Proposal{}, err
	}
	if item.AuthorAccountID != actorID || item.Status != "draft" {
		return Proposal{}, ErrForbidden
	}
	title, content, err := validateProposal(input)
	if err != nil {
		return Proposal{}, err
	}
	item.Title, item.Content, item.Summary, item.UpdatedAt = title, content, summarize(content, 180), service.now().UTC()
	item.Comments = nil
	if err := service.repository.Upsert(item); err != nil {
		return Proposal{}, err
	}
	if err := service.appendAudit(actorID, number, "rfc.update"); err != nil {
		return Proposal{}, err
	}
	return service.repository.Proposal(number)
}

func (service *Service) UpdateStatus(actorID string, number uint64, status string) (Proposal, error) {
	allowed, err := service.CanMaintain(actorID)
	if err != nil {
		return Proposal{}, err
	}
	if !allowed {
		return Proposal{}, ErrForbidden
	}
	status = strings.ToLower(strings.TrimSpace(status))
	if !validStatuses[status] {
		return Proposal{}, ErrInvalidStatus
	}
	item, err := service.repository.Proposal(number)
	if err != nil {
		return Proposal{}, err
	}
	item.Status, item.UpdatedAt, item.Comments = status, service.now().UTC(), nil
	if err := service.repository.Upsert(item); err != nil {
		return Proposal{}, err
	}
	if err := service.appendAudit(actorID, number, "rfc.status."+status); err != nil {
		return Proposal{}, err
	}
	return service.repository.Proposal(number)
}

func (service *Service) AddComment(actorID string, number uint64, input CommentInput) (Comment, error) {
	author, err := service.accounts.Account(actorID)
	if err != nil {
		return Comment{}, err
	}
	if _, err := service.repository.Proposal(number); err != nil {
		return Comment{}, err
	}
	body := strings.TrimSpace(strings.ReplaceAll(input.Body, "\r\n", "\n"))
	if len([]rune(body)) < 1 || len([]rune(body)) > 10000 {
		return Comment{}, fmt.Errorf("%w: body must contain between 1 and 10000 characters", ErrInvalidComment)
	}
	id, err := identifier.New("rfc-comment")
	if err != nil {
		return Comment{}, err
	}
	item := Comment{ID: id, ProposalNumber: number, AuthorAccountID: author.ID, AuthorName: author.DisplayName,
		Body: body, CreatedAt: service.now().UTC()}
	if err := service.repository.AddComment(item); err != nil {
		return Comment{}, err
	}
	if err := service.appendAudit(actorID, number, "rfc.comment"); err != nil {
		return Comment{}, err
	}
	return item, nil
}

func validateProposal(input ProposalInput) (string, string, error) {
	title := strings.TrimSpace(input.Title)
	content := strings.TrimSpace(strings.ReplaceAll(input.Content, "\r\n", "\n"))
	if len([]rune(title)) < 5 || len([]rune(title)) > 180 {
		return "", "", fmt.Errorf("%w: title must contain between 5 and 180 characters", ErrInvalidProposal)
	}
	if len([]rune(content)) < 20 || len([]rune(content)) > 200000 {
		return "", "", fmt.Errorf("%w: content must contain between 20 and 200000 characters", ErrInvalidProposal)
	}
	return title, content, nil
}

func summarize(content string, limit int) string {
	characters := []rune(strings.Join(strings.Fields(content), " "))
	if len(characters) > limit {
		return string(characters[:limit]) + "…"
	}
	return string(characters)
}

func (service *Service) appendAudit(actorID string, number uint64, action string) error {
	id, err := identifier.New("audit")
	if err != nil {
		return err
	}
	return service.audit.Append(audit.Event{ID: id, ActorID: "account/" + actorID,
		ResourceID: fmt.Sprintf("rfc/%d", number), Action: action, Result: "success", OccurredAt: service.now().UTC()})
}
