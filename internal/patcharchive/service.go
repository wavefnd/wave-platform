package patcharchive

import (
	"errors"
	"net/mail"
	"regexp"
	"sort"
	"strconv"
	"strings"

	maildomain "github.com/wavefnd/wave-platform/internal/mail"
	"github.com/wavefnd/wave-platform/internal/mailbox"
	"github.com/wavefnd/wave-platform/internal/storage"
)

var (
	subjectPattern = regexp.MustCompile(`(?i)^\s*(?:re:\s*)?\[patch([^\]]*)\]\s*(.*)$`)
	versionPattern = regexp.MustCompile(`(?i)(?:^|\s)v([0-9]+)(?:\s|$)`)
	partPattern    = regexp.MustCompile(`(?:^|\s)([0-9]+)/([0-9]+)(?:\s|$)`)
	filePattern    = regexp.MustCompile(`(?m)^diff --git a/(.+?) b/(.+?)\s*$`)
)

type Service struct {
	mailboxes *mailbox.Repository
	messages  *maildomain.Repository
	accountID string
	address   string
}

func NewService(database *storage.Database, accountID, address string) *Service {
	return &Service{mailboxes: mailbox.NewRepository(database), messages: maildomain.NewRepository(database),
		accountID: accountID, address: strings.ToLower(strings.TrimSpace(address))}
}

func Valid(subject, body string) bool {
	metadata := subjectPattern.FindStringSubmatch(subject)
	if len(metadata) != 3 {
		return false
	}
	if len(patchFiles(body)) > 0 {
		return true
	}
	if match := partPattern.FindStringSubmatch(metadata[1]); len(match) == 3 {
		part, _ := strconv.Atoi(match[1])
		total, _ := strconv.Atoi(match[2])
		return total > 1 && part == 0
	}
	return false
}

func (service *Service) Address() string { return service.address }

func (service *Service) List(query string, limit int) ([]Patch, error) {
	box, err := service.mailboxes.MailboxByAccount(service.accountID)
	if err != nil {
		return nil, err
	}
	entries, err := service.mailboxes.Entries(box.ID, "Inbox")
	if err != nil {
		return nil, err
	}
	query = strings.ToLower(strings.TrimSpace(query))
	items := make([]Patch, 0)
	for _, entry := range entries {
		value, valid, parseErr := service.patch(entry.MessageID, false)
		if parseErr != nil {
			return nil, parseErr
		}
		if !valid {
			continue
		}
		if query != "" && !strings.Contains(strings.ToLower(value.Subject+" "+value.AuthorName+" "+value.AuthorEmail+" "+strings.Join(value.Files, " ")+" "+value.Preview), query) {
			continue
		}
		items = append(items, value)
		if limit > 0 && len(items) >= limit {
			break
		}
	}
	sort.SliceStable(items, func(i, j int) bool { return items[i].ReceivedAt.After(items[j].ReceivedAt) })
	return items, nil
}

func (service *Service) Get(id string) (Patch, error) {
	box, err := service.mailboxes.MailboxByAccount(service.accountID)
	if err != nil {
		return Patch{}, err
	}
	entries, err := service.mailboxes.Entries(box.ID, "Inbox")
	if err != nil {
		return Patch{}, err
	}
	found := false
	for _, entry := range entries {
		if entry.MessageID == id {
			found = true
			break
		}
	}
	if !found {
		return Patch{}, storage.ErrNotFound
	}
	value, valid, err := service.patch(id, true)
	if err != nil {
		return Patch{}, err
	}
	if !valid {
		return Patch{}, storage.ErrNotFound
	}
	return value, nil
}

func (service *Service) patch(messageID string, includeBody bool) (Patch, bool, error) {
	message, err := service.messages.Message(messageID)
	if err != nil {
		return Patch{}, false, err
	}
	metadata := subjectPattern.FindStringSubmatch(message.Subject)
	if len(metadata) != 3 {
		return Patch{}, false, nil
	}
	body, err := service.messages.Body(message)
	if err != nil {
		return Patch{}, false, err
	}
	if !Valid(message.Subject, body) {
		return Patch{}, false, nil
	}
	version, part, total := 1, 0, 0
	if match := versionPattern.FindStringSubmatch(metadata[1]); len(match) == 2 {
		version, _ = strconv.Atoi(match[1])
	}
	if match := partPattern.FindStringSubmatch(metadata[1]); len(match) == 3 {
		part, _ = strconv.Atoi(match[1])
		total, _ = strconv.Atoi(match[2])
	}
	files := patchFiles(body)
	authorName, authorEmail := message.From, ""
	if parsed, parseErr := mail.ParseAddress(message.From); parseErr == nil {
		authorName, authorEmail = parsed.Name, strings.ToLower(parsed.Address)
		if authorName == "" {
			authorName = authorEmail
		}
	}
	value := Patch{ID: message.ID, MessageID: message.MessageID, Subject: message.Subject, Title: strings.TrimSpace(metadata[2]),
		AuthorName: authorName, AuthorEmail: authorEmail, Preview: preview(body), Version: version, Part: part, Total: total,
		Files: files, ReceivedAt: message.ReceivedAt}
	if includeBody {
		value.Body = truncate(body, 500000)
	}
	return value, true, nil
}

func patchFiles(body string) []string {
	seen := map[string]bool{}
	files := make([]string, 0)
	for _, match := range filePattern.FindAllStringSubmatch(body, 200) {
		name := strings.TrimSpace(match[2])
		if name == "/dev/null" {
			name = strings.TrimSpace(match[1])
		}
		if name != "" && !seen[name] {
			seen[name] = true
			files = append(files, name)
		}
	}
	return files
}

func preview(body string) string {
	position := strings.Index(body, "\ndiff --git ")
	if position >= 0 {
		body = body[:position]
	}
	lines := strings.Split(body, "\n")
	kept := make([]string, 0, len(lines))
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, ">") || trimmed == "---" || strings.HasPrefix(trimmed, "Signed-off-by:") {
			continue
		}
		kept = append(kept, trimmed)
	}
	return truncate(strings.Join(kept, " "), 220)
}

func truncate(value string, limit int) string {
	characters := []rune(strings.TrimSpace(value))
	if len(characters) > limit {
		return string(characters[:limit]) + "…"
	}
	return string(characters)
}

func IsNotFound(err error) bool { return errors.Is(err, storage.ErrNotFound) }
