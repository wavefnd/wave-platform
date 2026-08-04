package community

import (
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"net/textproto"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	maildomain "github.com/wavefnd/wave-platform/internal/mail"
	releasedomain "github.com/wavefnd/wave-platform/internal/release"
	"github.com/wavefnd/wave-platform/internal/storage"
)

//go:embed seed/posts/*.md
var releasePosts embed.FS

var releaseTitlePattern = regexp.MustCompile(`(?i)wave(?: language)? v[0-9]`)

func SeedSpaces(database *storage.Database) error {
	repository := NewRepository(database)
	spaces := []Space{
		{ID: "founder-notes", Slug: "founder-notes", Name: "Notes", Visibility: "public", PostingPolicy: "owner"},
		{ID: "development-log", Slug: "development-log", Name: "Development Log", Visibility: "public", PostingPolicy: "owner"},
		{ID: "general", Slug: "general", Name: "General", Visibility: "public", PostingPolicy: "members"},
		{ID: "development", Slug: "development", Name: "Development", Visibility: "public", PostingPolicy: "members"},
		{ID: "showcase", Slug: "showcase", Name: "Showcase", Visibility: "public", PostingPolicy: "members"},
		{ID: "help", Slug: "help", Name: "Help", Visibility: "public", PostingPolicy: "members"},
	}
	for _, space := range spaces {
		if err := repository.UpsertSpace(space); err != nil {
			return err
		}
	}
	return nil
}

func SeedLanguageReleases(database *storage.Database) (int, error) {
	entries, err := releasePosts.ReadDir("seed/posts")
	if err != nil {
		return 0, fmt.Errorf("read embedded release posts: %w", err)
	}

	hash := sha256.New()
	releases := make([]releasedomain.Release, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".md" {
			continue
		}

		source, err := releasePosts.ReadFile("seed/posts/" + entry.Name())
		if err != nil {
			return 0, fmt.Errorf("read embedded release %q: %w", entry.Name(), err)
		}
		_, _ = hash.Write(source)

		item, err := parseReleasePost(entry.Name(), string(source))
		if err != nil {
			return 0, err
		}
		releases = append(releases, item)
	}

	marker := storage.Key(
		"meta", "import", "wave-blog", "language-releases-v2",
		hex.EncodeToString(hash.Sum(nil)),
	)
	if _, err := database.Get(marker); err == nil {
		return 0, nil
	}

	repository := releasedomain.NewRepository(database)
	mailRepository := maildomain.NewRepository(database)
	for _, item := range releases {
		published, err := time.Parse(time.RFC3339, item.PublishedAt)
		if err != nil {
			return 0, fmt.Errorf("parse normalized release time: %w", err)
		}
		raw := releaseMailMessage(item, published)
		if err := mailRepository.UpsertMessage(maildomain.Message{
			ID: item.MessageID, MessageID: "<" + item.Slug + "@wave-lang.dev>",
			ThreadID: item.MessageID, From: "Wave Foundation <release@wave-lang.dev>",
			To: []string{"releases@wave-lang.dev"}, Subject: item.Title,
			ReceivedAt: published, CreatedAt: published,
		}, raw); err != nil {
			return 0, err
		}
		if err := repository.Upsert(item); err != nil {
			return 0, err
		}
	}
	legacyKeys := make([][]byte, 0)
	if err := database.Scan(storage.Prefix("community", "announcement", "object"), func(key, value []byte) error {
		var legacy struct {
			Source string `xml:"migration-source"`
		}
		if err := xml.Unmarshal(value, &legacy); err == nil && legacy.Source == "github.com/LunaStev/wave-blog" {
			legacyKeys = append(legacyKeys, append([]byte(nil), key...))
		}
		return nil
	}); err != nil {
		return 0, err
	}
	for _, key := range legacyKeys {
		if err := database.Delete(key); err != nil {
			return 0, err
		}
	}
	if err := database.Set(marker, []byte("complete")); err != nil {
		return 0, err
	}

	return len(releases), nil
}

func parseReleasePost(filename, source string) (releasedomain.Release, error) {
	if strings.Contains(strings.ToLower(filename), "patch") {
		return releasedomain.Release{}, fmt.Errorf("patch post is not a release: %s", filename)
	}

	normalized := strings.ReplaceAll(source, "\r\n", "\n")
	if !strings.HasPrefix(normalized, "---\n") {
		return releasedomain.Release{}, fmt.Errorf("invalid front matter: %s", filename)
	}
	remainder := strings.TrimPrefix(normalized, "---\n")
	separator := strings.Index(remainder, "\n---\n")
	if separator < 0 {
		return releasedomain.Release{}, fmt.Errorf("invalid front matter: %s", filename)
	}

	metadata := remainder[:separator]
	body := strings.TrimSpace(remainder[separator+5:])
	title := frontMatterValue(metadata, "title")
	if !releaseTitlePattern.MatchString(title) {
		return releasedomain.Release{}, fmt.Errorf("post is not a version release: %s", filename)
	}

	published, err := time.ParseInLocation(
		"2006-01-02 15:04:05",
		frontMatterValue(metadata, "date"),
		time.FixedZone("Asia/Seoul", 9*60*60),
	)
	if err != nil {
		return releasedomain.Release{}, fmt.Errorf("parse release date %q: %w", filename, err)
	}

	slug := strings.TrimSuffix(filename, filepath.Ext(filename))
	if len(slug) > 11 && slug[4] == '-' && slug[7] == '-' && slug[10] == '-' {
		slug = slug[11:]
	}

	summary := frontMatterValue(metadata, "description")
	if summary == "" {
		summary = firstParagraph(body)
	}
	sourceName := frontMatterValue(metadata, "source")
	idPrefix := "wave-blog"
	if sourceName == "" {
		sourceName = "github.com/LunaStev/wave-blog"
	} else {
		idPrefix = "wave-release"
	}

	return releasedomain.Release{
		ID:          idPrefix + "/" + slug,
		Slug:        slug,
		Title:       title,
		PublishedAt: published.Format(time.RFC3339),
		Summary:     summary,
		MessageID:   idPrefix + "/" + slug,
		Content:     body,
		Source:      sourceName,
	}, nil
}

func releaseMailMessage(item releasedomain.Release, published time.Time) []byte {
	headers := textproto.MIMEHeader{}
	headers.Set("From", "Wave Foundation <release@wave-lang.dev>")
	headers.Set("To", "releases@wave-lang.dev")
	headers.Set("Subject", item.Title)
	headers.Set("Date", published.Format(time.RFC1123Z))
	headers.Set("Message-ID", "<"+item.Slug+"@wave-lang.dev>")
	headers.Set("MIME-Version", "1.0")
	headers.Set("Content-Type", "text/markdown; charset=utf-8")
	headers.Set("Content-Transfer-Encoding", "8bit")

	order := []string{"From", "To", "Subject", "Date", "Message-ID", "MIME-Version", "Content-Type", "Content-Transfer-Encoding"}
	var message strings.Builder
	for _, name := range order {
		message.WriteString(name)
		message.WriteString(": ")
		message.WriteString(headers.Get(name))
		message.WriteString("\r\n")
	}
	message.WriteString("\r\n")
	message.WriteString(item.Content)
	return []byte(message.String())
}

func frontMatterValue(metadata, key string) string {
	lines := strings.Split(metadata, "\n")
	prefix := key + ":"
	for index, line := range lines {
		if !strings.HasPrefix(line, prefix) {
			continue
		}

		value := strings.TrimSpace(strings.TrimPrefix(line, prefix))
		if strings.HasPrefix(value, "\"") && !strings.HasSuffix(value, "\"") {
			for index++; index < len(lines); index++ {
				value += " " + strings.TrimSpace(lines[index])
				if strings.HasSuffix(strings.TrimSpace(lines[index]), "\"") {
					break
				}
			}
		}
		return strings.Trim(strings.TrimSpace(value), "\"")
	}
	return ""
}

func firstParagraph(body string) string {
	for _, paragraph := range strings.Split(body, "\n\n") {
		paragraph = strings.TrimSpace(paragraph)
		if paragraph == "" || strings.HasPrefix(paragraph, "#") || strings.HasPrefix(paragraph, "```") {
			continue
		}
		return strings.Join(strings.Fields(paragraph), " ")
	}
	return ""
}
