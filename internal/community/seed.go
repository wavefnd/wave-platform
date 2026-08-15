package community

import (
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"encoding/xml"
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	blogdomain "github.com/wavefnd/wave-platform/internal/blog"
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
		{ID: "operating-systems", Slug: "operating-systems", Name: "Operating Systems", Visibility: "public", PostingPolicy: "members"},
		{ID: "web", Slug: "web", Name: "Web", Visibility: "public", PostingPolicy: "members"},
		{ID: "compiler", Slug: "compiler", Name: "Compiler", Visibility: "public", PostingPolicy: "members"},
		{ID: "audio", Slug: "audio", Name: "Audio", Visibility: "public", PostingPolicy: "members"},
		{ID: "gui", Slug: "gui", Name: "GUI", Visibility: "public", PostingPolicy: "members"},
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

func SeedReleaseBlogPosts(database *storage.Database) (int, error) {
	entries, err := releasePosts.ReadDir("seed/posts")
	if err != nil {
		return 0, fmt.Errorf("read embedded release posts: %w", err)
	}

	hash := sha256.New()
	posts := make([]blogdomain.Post, 0, len(entries))
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
		posts = append(posts, item)
	}

	marker := storage.Key(
		"meta", "import", "wave-blog", "release-blog-posts-v1",
		hex.EncodeToString(hash.Sum(nil)),
	)
	if _, err := database.Get(marker); err == nil {
		return 0, nil
	}

	repository := blogdomain.NewRepository(database)
	imported := 0
	for _, item := range posts {
		if _, err := repository.Post(item.Slug, true); err == nil {
			continue
		} else if !errors.Is(err, storage.ErrNotFound) {
			return 0, err
		}
		if err := repository.Upsert(item); err != nil {
			return 0, err
		}
		imported++
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

	return imported, nil
}

func parseReleasePost(filename, source string) (blogdomain.Post, error) {
	if strings.Contains(strings.ToLower(filename), "patch") {
		return blogdomain.Post{}, fmt.Errorf("patch post is not a release: %s", filename)
	}

	normalized := strings.ReplaceAll(source, "\r\n", "\n")
	if !strings.HasPrefix(normalized, "---\n") {
		return blogdomain.Post{}, fmt.Errorf("invalid front matter: %s", filename)
	}
	remainder := strings.TrimPrefix(normalized, "---\n")
	separator := strings.Index(remainder, "\n---\n")
	if separator < 0 {
		return blogdomain.Post{}, fmt.Errorf("invalid front matter: %s", filename)
	}

	metadata := remainder[:separator]
	body := strings.TrimSpace(remainder[separator+5:])
	title := frontMatterValue(metadata, "title")
	if !releaseTitlePattern.MatchString(title) {
		return blogdomain.Post{}, fmt.Errorf("post is not a version release: %s", filename)
	}

	published, err := time.ParseInLocation(
		"2006-01-02 15:04:05",
		frontMatterValue(metadata, "date"),
		time.FixedZone("Asia/Seoul", 9*60*60),
	)
	if err != nil {
		return blogdomain.Post{}, fmt.Errorf("parse release date %q: %w", filename, err)
	}

	slug := strings.TrimSuffix(filename, filepath.Ext(filename))
	if len(slug) > 11 && slug[4] == '-' && slug[7] == '-' && slug[10] == '-' {
		slug = slug[11:]
	}

	summary := frontMatterValue(metadata, "description")
	if summary == "" {
		summary = firstParagraph(body)
	}
	return blogdomain.Post{
		Slug:        slug,
		Category:    "release",
		Title:       title,
		Summary:     summary,
		Content:     body,
		Status:      "published",
		AuthorName:  "Wave Foundation",
		PublishedAt: published.Format(time.RFC3339),
		CreatedAt:   published.UTC(),
		UpdatedAt:   published.UTC(),
	}, nil
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
