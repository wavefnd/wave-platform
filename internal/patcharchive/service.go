package patcharchive

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/xml"
	"errors"
	"fmt"
	"net/mail"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/wavefnd/wave-platform/internal/account"
	"github.com/wavefnd/wave-platform/internal/audit"
	"github.com/wavefnd/wave-platform/internal/identifier"
	maildomain "github.com/wavefnd/wave-platform/internal/mail"
	"github.com/wavefnd/wave-platform/internal/mailbox"
	"github.com/wavefnd/wave-platform/internal/permission"
	"github.com/wavefnd/wave-platform/internal/storage"
)

var (
	subjectPattern          = regexp.MustCompile(`(?i)^\s*(?:re:\s*)?\[patch([^\]]*)\]\s*(.*)$`)
	versionPattern          = regexp.MustCompile(`(?i)(?:^|\s)v([0-9]+)(?:\s|$)`)
	partPattern             = regexp.MustCompile(`(?:^|\s)([0-9]+)/([0-9]+)(?:\s|$)`)
	filePattern             = regexp.MustCompile(`(?m)^diff --git a/(.+?) b/(.+?)\s*$`)
	messageIDPattern        = regexp.MustCompile(`<[^<>\s]+>`)
	targetRepositoryPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,49}/[A-Za-z0-9][A-Za-z0-9._-]{0,49}$`)
)

var (
	ErrForbidden      = errors.New("source maintainer access is required")
	ErrInvalidReview  = errors.New("invalid patch review")
	ErrInvalidComment = errors.New("invalid patch review comment")
)

type Service struct {
	database    *storage.Database
	mailboxes   *mailbox.Repository
	messages    *maildomain.Repository
	accounts    *account.Repository
	permissions *permission.Repository
	audit       *audit.Repository
	accountID   string
	address     string
	now         func() time.Time
}

func NewService(database *storage.Database, accountID, address string) *Service {
	return &Service{database: database, mailboxes: mailbox.NewRepository(database), messages: maildomain.NewRepository(database),
		accounts: account.NewRepository(database), permissions: permission.NewRepository(database), audit: audit.NewRepository(database),
		accountID: accountID, address: strings.ToLower(strings.TrimSpace(address)), now: time.Now}
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
	members, err := service.seriesMembers(id)
	if err != nil {
		return Patch{}, err
	}
	value.SeriesCount = len(members)
	comments, err := service.ReviewComments(id)
	if err != nil {
		return Patch{}, err
	}
	value.ReviewComments = comments
	value.ReviewCommentCount = len(comments)
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
		Files: files, ReviewStatus: "received", ReceivedAt: message.ReceivedAt}
	if review, reviewErr := service.review(message.ID); reviewErr == nil {
		value.ReviewStatus, value.TargetRepository = review.Status, review.TargetRepository
		value.AssigneeAccountID, value.ReviewUpdatedAt = review.AssigneeAccountID, review.UpdatedAt
		if assignee, accountErr := service.accounts.Account(review.AssigneeAccountID); accountErr == nil {
			value.AssigneeName = assignee.DisplayName
		}
	} else if !errors.Is(reviewErr, storage.ErrNotFound) {
		return Patch{}, false, reviewErr
	}
	comments, commentErr := service.ReviewComments(message.ID)
	if commentErr != nil {
		return Patch{}, false, commentErr
	}
	value.ReviewCommentCount = len(comments)
	if includeBody {
		value.Body = truncate(body, 500000)
		raw, rawErr := service.messages.RawMessage(message)
		if rawErr != nil {
			return Patch{}, false, rawErr
		}
		digest := sha256.Sum256(raw)
		value.SHA256 = hex.EncodeToString(digest[:])
	}
	return value, true, nil
}

func (service *Service) ReviewComments(patchID string) ([]ReviewComment, error) {
	items := make([]ReviewComment, 0)
	err := service.database.Scan(storage.Prefix("patch", "comment", patchID), func(_, data []byte) error {
		var item ReviewComment
		if err := xml.Unmarshal(data, &item); err != nil {
			return fmt.Errorf("decode patch review comment: %w", err)
		}
		if author, accountErr := service.accounts.Account(item.AuthorAccountID); accountErr == nil {
			item.AuthorName = author.DisplayName
		}
		items = append(items, item)
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.SliceStable(items, func(left, right int) bool {
		if items[left].Line == items[right].Line {
			return items[left].CreatedAt.Before(items[right].CreatedAt)
		}
		return items[left].Line < items[right].Line
	})
	return items, nil
}

func (service *Service) AddReviewComment(actorID, patchID string, input ReviewCommentInput) (ReviewComment, error) {
	allowed, err := service.CanMaintain(actorID)
	if err != nil {
		return ReviewComment{}, err
	}
	if !allowed {
		return ReviewComment{}, ErrForbidden
	}
	body := strings.TrimSpace(input.Body)
	if body == "" || len([]rune(body)) > 4000 {
		return ReviewComment{}, fmt.Errorf("%w: comment body must contain 1 to 4000 characters", ErrInvalidComment)
	}
	patch, err := service.Get(patchID)
	if err != nil {
		return ReviewComment{}, err
	}
	path, lineText := "", ""
	if input.Line < 0 {
		return ReviewComment{}, fmt.Errorf("%w: line cannot be negative", ErrInvalidComment)
	}
	if input.Line > 0 {
		path, lineText = patchLine(patch.Body, input.Line)
		if path == "" {
			return ReviewComment{}, fmt.Errorf("%w: line must identify a line inside a patch diff", ErrInvalidComment)
		}
	}
	id, err := identifier.New("patch-comment")
	if err != nil {
		return ReviewComment{}, err
	}
	now := service.now().UTC()
	item := ReviewComment{ID: id, PatchID: patchID, AuthorAccountID: actorID, Path: path, Line: input.Line,
		LineText: truncate(lineText, 1000), Body: body, CreatedAt: now, UpdatedAt: now}
	data, err := xml.Marshal(item)
	if err != nil {
		return ReviewComment{}, err
	}
	if err := service.database.Set(storage.Key("patch", "comment", patchID, id), data); err != nil {
		return ReviewComment{}, err
	}
	if err := service.appendAudit(actorID, "patch/"+patchID+"/comment/"+id, "patch.review-comment.create"); err != nil {
		return ReviewComment{}, err
	}
	if author, accountErr := service.accounts.Account(actorID); accountErr == nil {
		item.AuthorName = author.DisplayName
	}
	return item, nil
}

func (service *Service) ResolveReviewComment(actorID, patchID, commentID string, resolved bool) (ReviewComment, error) {
	allowed, err := service.CanMaintain(actorID)
	if err != nil {
		return ReviewComment{}, err
	}
	if !allowed {
		return ReviewComment{}, ErrForbidden
	}
	if _, err := service.Get(patchID); err != nil {
		return ReviewComment{}, err
	}
	key := storage.Key("patch", "comment", patchID, commentID)
	data, err := service.database.Get(key)
	if err != nil {
		return ReviewComment{}, err
	}
	var item ReviewComment
	if err := xml.Unmarshal(data, &item); err != nil {
		return ReviewComment{}, fmt.Errorf("decode patch review comment: %w", err)
	}
	item.Resolved = resolved
	item.UpdatedAt = service.now().UTC()
	data, err = xml.Marshal(item)
	if err != nil {
		return ReviewComment{}, err
	}
	if err := service.database.Set(key, data); err != nil {
		return ReviewComment{}, err
	}
	action := "patch.review-comment.reopen"
	if resolved {
		action = "patch.review-comment.resolve"
	}
	if err := service.appendAudit(actorID, "patch/"+patchID+"/comment/"+commentID, action); err != nil {
		return ReviewComment{}, err
	}
	if author, accountErr := service.accounts.Account(item.AuthorAccountID); accountErr == nil {
		item.AuthorName = author.DisplayName
	}
	return item, nil
}

func (service *Service) CanMaintain(accountID string) (bool, error) {
	owner, err := service.permissions.HasRole(accountID, "platform-owner")
	if err != nil || owner {
		return owner, err
	}
	return service.permissions.HasRole(accountID, "source-maintainer")
}

func (service *Service) UpdateReview(actorID, patchID string, input ReviewInput) (Patch, error) {
	allowed, err := service.CanMaintain(actorID)
	if err != nil {
		return Patch{}, err
	}
	if !allowed {
		return Patch{}, ErrForbidden
	}
	status := strings.ToLower(strings.TrimSpace(input.Status))
	if !map[string]bool{"received": true, "reviewing": true, "accepted": true, "rejected": true, "applied": true}[status] {
		return Patch{}, fmt.Errorf("%w: unsupported status", ErrInvalidReview)
	}
	target := strings.TrimSpace(input.TargetRepository)
	if target != "" && !targetRepositoryPattern.MatchString(target) {
		return Patch{}, fmt.Errorf("%w: target repository must use owner/name", ErrInvalidReview)
	}
	if _, err := service.Get(patchID); err != nil {
		return Patch{}, err
	}
	now := service.now().UTC()
	review := Review{PatchID: patchID, Status: status, TargetRepository: target, AssigneeAccountID: actorID, UpdatedBy: actorID, UpdatedAt: now}
	data, err := xml.Marshal(review)
	if err != nil {
		return Patch{}, err
	}
	if err := service.database.Set(storage.Key("patch", "review", patchID), data); err != nil {
		return Patch{}, err
	}
	if err := service.appendAudit(actorID, "patch/"+patchID, "patch.review."+status); err != nil {
		return Patch{}, err
	}
	return service.Get(patchID)
}

func (service *Service) DownloadMbox(actorID, patchID string, includeSeries bool) ([]byte, string, error) {
	allowed, err := service.CanMaintain(actorID)
	if err != nil {
		return nil, "", err
	}
	if !allowed {
		return nil, "", ErrForbidden
	}
	members, err := service.seriesMembers(patchID)
	if err != nil {
		return nil, "", err
	}
	if !includeSeries || len(members) < 2 {
		requested := members[0]
		for _, item := range members {
			if item.ID == patchID {
				requested = item
				break
			}
		}
		members = []Patch{requested}
	}
	var output bytes.Buffer
	for _, item := range members {
		message, messageErr := service.messages.Message(item.ID)
		if messageErr != nil {
			return nil, "", messageErr
		}
		raw, rawErr := service.messages.RawMessage(message)
		if rawErr != nil {
			return nil, "", rawErr
		}
		output.Write(mboxMessage(raw, item.AuthorEmail, item.ReceivedAt))
	}
	action, prefix := "patch.download", "patch-"
	if includeSeries && len(members) > 1 {
		action, prefix = "patch.series.download", "patch-series-"
	}
	if err := service.appendAudit(actorID, "patch/"+patchID, action); err != nil {
		return nil, "", err
	}
	return output.Bytes(), prefix + filenameToken(patchID) + ".mbox", nil
}

type seriesNode struct {
	patch  Patch
	parent string
}

func (service *Service) seriesMembers(patchID string) ([]Patch, error) {
	box, err := service.mailboxes.MailboxByAccount(service.accountID)
	if err != nil {
		return nil, err
	}
	entries, err := service.mailboxes.Entries(box.ID, "Inbox")
	if err != nil {
		return nil, err
	}
	nodes := make(map[string]seriesNode)
	byMessageID := make(map[string]string)
	for _, entry := range entries {
		item, valid, patchErr := service.patch(entry.MessageID, false)
		if patchErr != nil {
			return nil, patchErr
		}
		if !valid {
			continue
		}
		message, messageErr := service.messages.Message(entry.MessageID)
		if messageErr != nil {
			return nil, messageErr
		}
		raw, rawErr := service.messages.RawMessage(message)
		if rawErr != nil {
			return nil, rawErr
		}
		parent := ""
		if parsed, parseErr := mail.ReadMessage(bytes.NewReader(raw)); parseErr == nil {
			parent = firstMessageID(parsed.Header.Get("In-Reply-To"))
		}
		nodes[item.ID] = seriesNode{patch: item, parent: parent}
		byMessageID[firstMessageID(item.MessageID)] = item.ID
	}
	target, found := nodes[patchID]
	if !found {
		return nil, storage.ErrNotFound
	}
	if target.patch.Total < 2 {
		return []Patch{target.patch}, nil
	}
	rootOf := func(id string) string {
		seen := map[string]bool{}
		for id != "" && !seen[id] {
			seen[id] = true
			node := nodes[id]
			parentID := byMessageID[node.parent]
			if parentID == "" {
				return id
			}
			id = parentID
		}
		return id
	}
	root := rootOf(patchID)
	members := make([]Patch, 0)
	for id, node := range nodes {
		if rootOf(id) == root && node.patch.Version == target.patch.Version && node.patch.Total == target.patch.Total && strings.EqualFold(node.patch.AuthorEmail, target.patch.AuthorEmail) {
			members = append(members, node.patch)
		}
	}
	if len(members) == 0 {
		members = append(members, target.patch)
	}
	sort.SliceStable(members, func(i, j int) bool {
		if members[i].Part == members[j].Part {
			return members[i].ReceivedAt.Before(members[j].ReceivedAt)
		}
		return members[i].Part < members[j].Part
	})
	return members, nil
}

func (service *Service) review(patchID string) (Review, error) {
	data, err := service.database.Get(storage.Key("patch", "review", patchID))
	if err != nil {
		return Review{}, err
	}
	var review Review
	if err := xml.Unmarshal(data, &review); err != nil {
		return Review{}, fmt.Errorf("decode patch review: %w", err)
	}
	return review, nil
}

func (service *Service) appendAudit(actorID, resourceID, action string) error {
	id, err := identifier.New("audit")
	if err != nil {
		return err
	}
	return service.audit.Append(audit.Event{ID: id, ActorID: "account/" + actorID, ResourceID: resourceID,
		Action: action, Result: "success", OccurredAt: service.now().UTC()})
}

func firstMessageID(value string) string {
	return messageIDPattern.FindString(value)
}

func mboxMessage(raw []byte, sender string, receivedAt time.Time) []byte {
	if strings.TrimSpace(sender) == "" {
		sender = "unknown@localhost"
	}
	content := strings.ReplaceAll(string(raw), "\r\n", "\n")
	content = strings.ReplaceAll(content, "\nFrom ", "\n>From ")
	content = strings.TrimRight(content, "\n") + "\n"
	separator := "From " + strings.Fields(sender)[0] + " " + receivedAt.UTC().Format("Mon Jan _2 15:04:05 2006") + "\n"
	return []byte(separator + content + "\n")
}

func filenameToken(value string) string {
	var result strings.Builder
	for _, character := range value {
		if character >= 'a' && character <= 'z' || character >= 'A' && character <= 'Z' || character >= '0' && character <= '9' || character == '-' || character == '_' {
			result.WriteRune(character)
		}
	}
	if result.Len() == 0 {
		return "download"
	}
	return result.String()
}

func patchLine(body string, number int) (string, string) {
	if number < 1 {
		return "", ""
	}
	path := ""
	for index, line := range strings.Split(body, "\n") {
		if match := filePattern.FindStringSubmatch(line); len(match) == 3 {
			path = strings.TrimSpace(match[2])
			if path == "/dev/null" {
				path = strings.TrimSpace(match[1])
			}
		}
		if index+1 == number {
			return path, line
		}
	}
	return "", ""
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
