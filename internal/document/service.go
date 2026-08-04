package document

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"encoding/xml"
	"errors"
	"fmt"
	"io/fs"
	"strconv"
	"strings"
	"time"

	"github.com/wavefnd/wave-platform/internal/storage"
)

// Official documents are ordinary Markdown files. The platform turns them into
// XML revisions when it seeds the document repository.
//
//go:embed content
var officialContent embed.FS

type seedDocument struct {
	TranslationSetID string
	Path             string
	Locale           string
	Group            string
	GroupOrder       int
	Order            int
	Title            string
	Summary          string
	Markdown         string
	Headings         []Block
}

func SeedOfficial(database *storage.Database) (int, error) {
	documents, digest, err := readOfficialDocuments()
	if err != nil {
		return 0, err
	}
	marker := storage.Key("meta", "import", "official-documents", hex.EncodeToString(digest))
	if _, err := database.Get(marker); err == nil {
		return 0, nil
	}
	repository := NewRepository(database)
	publishedAt := time.Date(2026, time.August, 5, 0, 0, 0, 0, time.UTC)
	for _, sourceDocument := range documents {
		path := strings.Trim(sourceDocument.Path, "/")
		id := "official/" + sourceDocument.Locale + "/" + path
		revisionID := "v0.2.0-pre-beta"
		contentXML, err := xml.Marshal(Content{Markdown: sourceDocument.Markdown, Blocks: sourceDocument.Headings})
		if err != nil {
			return 0, fmt.Errorf("encode %s document content: %w", path, err)
		}
		contentHash := sha256.Sum256(contentXML)
		if err := repository.PutRevision(Revision{ID: revisionID, DocumentID: id, AuthorID: "wave-foundation",
			ContentHash: hex.EncodeToString(contentHash[:]), ContentXML: contentXML, CreatedAt: publishedAt}); err != nil {
			return 0, err
		}
		if err := repository.UpsertDocument(Document{ID: id, TranslationSetID: sourceDocument.TranslationSetID,
			Path: path, Locale: sourceDocument.Locale, Group: sourceDocument.Group, GroupOrder: sourceDocument.GroupOrder,
			Order: sourceDocument.Order, Title: sourceDocument.Title, Summary: sourceDocument.Summary,
			Version: "0.2.0-pre-beta", SourceRevision: "bd5549b", Status: "published",
			PublishedRevisionID: revisionID, CreatedAt: publishedAt, UpdatedAt: publishedAt}); err != nil {
			return 0, err
		}
	}
	if err := database.Set(marker, []byte("complete")); err != nil {
		return 0, err
	}
	return len(documents), nil
}

func readOfficialDocuments() ([]seedDocument, []byte, error) {
	result := make([]seedDocument, 0, 48)
	hash := sha256.New()
	err := fs.WalkDir(officialContent, "content", func(name string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || !strings.HasSuffix(strings.ToLower(name), ".md") {
			return nil
		}
		source, err := officialContent.ReadFile(name)
		if err != nil {
			return err
		}
		document, err := parseMarkdownDocument(source)
		if err != nil {
			return fmt.Errorf("parse %s: %w", name, err)
		}
		if document.Locale != "en" && document.Locale != "ko" {
			return fmt.Errorf("parse %s: locale must be en or ko", name)
		}
		if document.Path == "" || document.Title == "" || document.Group == "" || document.TranslationSetID == "" {
			return fmt.Errorf("parse %s: translation_set_id, path, locale, group, and title are required", name)
		}
		_, _ = hash.Write([]byte(name))
		_, _ = hash.Write([]byte{0})
		_, _ = hash.Write(source)
		result = append(result, document)
		return nil
	})
	if err != nil {
		return nil, nil, fmt.Errorf("read official documents: %w", err)
	}
	if len(result) == 0 {
		return nil, nil, errors.New("no official Markdown documents found")
	}
	return result, hash.Sum(nil), nil
}

func parseMarkdownDocument(source []byte) (seedDocument, error) {
	const delimiter = "---"
	text := strings.ReplaceAll(string(source), "\r\n", "\n")
	if !strings.HasPrefix(text, delimiter+"\n") {
		return seedDocument{}, errors.New("document must start with Markdown front matter")
	}
	parts := strings.SplitN(text[len(delimiter)+1:], "\n"+delimiter+"\n", 2)
	if len(parts) != 2 {
		return seedDocument{}, errors.New("front matter closing delimiter is missing")
	}
	metadata := make(map[string]string)
	for _, line := range strings.Split(parts[0], "\n") {
		key, value, found := strings.Cut(line, ":")
		if !found || strings.TrimSpace(key) == "" {
			return seedDocument{}, fmt.Errorf("invalid front matter line %q", line)
		}
		metadata[strings.TrimSpace(key)] = strings.TrimSpace(value)
	}
	groupOrder, err := integerMetadata(metadata, "group_order")
	if err != nil {
		return seedDocument{}, err
	}
	order, err := integerMetadata(metadata, "order")
	if err != nil {
		return seedDocument{}, err
	}
	markdown := strings.TrimSpace(parts[1]) + "\n"
	return seedDocument{
		TranslationSetID: metadata["translation_set_id"], Path: strings.Trim(metadata["path"], "/"),
		Locale: metadata["locale"], Group: metadata["group"], GroupOrder: groupOrder, Order: order,
		Title: metadata["title"], Summary: metadata["summary"], Markdown: markdown,
		Headings: markdownHeadings(markdown),
	}, nil
}

func integerMetadata(metadata map[string]string, key string) (int, error) {
	value, err := strconv.Atoi(metadata[key])
	if err != nil {
		return 0, fmt.Errorf("%s must be an integer", key)
	}
	return value, nil
}

func markdownHeadings(markdown string) []Block {
	scanner := bufio.NewScanner(bytes.NewBufferString(markdown))
	result := make([]Block, 0, 8)
	inFence := false
	occurrences := make(map[string]int)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(strings.TrimSpace(line), "```") {
			inFence = !inFence
			continue
		}
		if inFence {
			continue
		}
		marks := 0
		for marks < len(line) && line[marks] == '#' {
			marks++
		}
		if marks < 2 || marks > 3 || marks >= len(line) || line[marks] != ' ' {
			continue
		}
		text := strings.TrimSpace(line[marks+1:])
		anchor := markdownAnchor(text)
		if count := occurrences[anchor]; count > 0 {
			anchor += "-" + strconv.Itoa(count)
		}
		occurrences[markdownAnchor(text)]++
		result = append(result, Block{Kind: "heading", Anchor: anchor, Level: marks, Text: text})
	}
	return result
}

func markdownAnchor(value string) string {
	var builder strings.Builder
	for _, character := range strings.ToLower(value) {
		switch {
		case character == ' ':
			builder.WriteRune('-')
		case character == '-' || character == '_' || character >= '0' && character <= '9' ||
			character >= 'a' && character <= 'z' || character >= 0x80:
			builder.WriteRune(character)
		}
	}
	return builder.String()
}
