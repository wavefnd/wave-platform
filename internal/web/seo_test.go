package web

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	blogdomain "github.com/wavefnd/wave-platform/internal/blog"
	documentdomain "github.com/wavefnd/wave-platform/internal/document"
	"github.com/wavefnd/wave-platform/internal/storage"
)

func TestBlogMetadataUsesPublishedPostData(t *testing.T) {
	service, closeDatabase := testSEOService(t)
	defer closeDatabase()
	updatedAt := time.Date(2026, time.August, 25, 9, 30, 0, 0, time.UTC)
	if err := service.Repository().Upsert(blogdomain.Post{
		Slug: "compiler-notes", Category: "article", Title: "Compiler notes", Summary: "How Wave lowers source code.",
		Content: "![Compiler pipeline](/media/lunastev/pipeline.webp)\n\nThe compiler pipeline.", Status: "published",
		AuthorAccountID: "founder-account", AuthorName: "LunaStev", PublishedAt: "2026-08-24", CreatedAt: updatedAt.Add(-24 * time.Hour), UpdatedAt: updatedAt,
	}); err != nil {
		t.Fatalf("store blog post: %v", err)
	}

	seo := NewSEOHandler("https://wave.example", nil, service, nil, nil, nil)
	request := httptest.NewRequest(http.MethodGet, "/blog/compiler-notes", nil)
	metadata := seo.HTMLMetadata(request)
	for _, expected := range []string{
		`<meta property="og:type" content="article" />`,
		`<meta property="og:image" content="https://wave.example/media/lunastev/pipeline.webp" />`,
		`<meta name="twitter:card" content="summary_large_image" />`,
		`<meta property="article:published_time" content="2026-08-24" />`,
		`<meta property="article:modified_time" content="2026-08-25T09:30:00Z" />`,
		`<meta property="article:author" content="LunaStev" />`,
		`rel="canonical" href="https://wave.example/blog/compiler-notes"`,
	} {
		if !strings.Contains(metadata, expected) {
			t.Errorf("metadata does not contain %q\n%s", expected, metadata)
		}
	}

	graph := metadataGraph(t, metadata)
	article := graphNode(t, graph, "BlogPosting")
	if article["headline"] != "Compiler notes" || article["datePublished"] != "2026-08-24" || article["articleSection"] != "Blog" {
		t.Fatalf("article schema = %#v", article)
	}
	author, _ := article["author"].(map[string]any)
	if author["@type"] != "Person" || author["name"] != "LunaStev" || author["url"] != "https://wave.example/user/id/founder-account" {
		t.Fatalf("article author = %#v", author)
	}
	graphNode(t, graph, "BreadcrumbList")
	if strings.Contains(metadata, "SearchAction") {
		t.Fatal("obsolete SearchAction must not be emitted")
	}
}

func TestReleaseCanonicalRedirectAndSitemap(t *testing.T) {
	service, closeDatabase := testSEOService(t)
	defer closeDatabase()
	now := time.Date(2026, time.August, 25, 0, 0, 0, 0, time.UTC)
	if err := service.Repository().Upsert(blogdomain.Post{
		Slug: "v0.2.2", Category: "release", Title: "Wave v0.2.2", Summary: "Wave v0.2.2 release.",
		Content: "Release notes.", Status: "published", AuthorName: "Wave Foundation", PublishedAt: "2026-08-25", CreatedAt: now, UpdatedAt: now,
	}); err != nil {
		t.Fatalf("store release: %v", err)
	}
	seo := NewSEOHandler("https://wave.example", nil, service, nil, nil, nil)
	if redirect := seo.CanonicalRedirect("/blog/v0.2.2"); redirect != "/releases/v0.2.2" {
		t.Fatalf("redirect = %q", redirect)
	}

	frontend := t.TempDir()
	if err := os.WriteFile(filepath.Join(frontend, "index.html"), []byte(`<head><!-- wave:seo:start --><!-- wave:seo:end --></head>`), 0o600); err != nil {
		t.Fatalf("write frontend: %v", err)
	}
	request := httptest.NewRequest(http.MethodGet, "/blog/v0.2.2", nil)
	response := httptest.NewRecorder()
	frontendHandler(frontend, seo).ServeHTTP(response, request)
	if response.Code != http.StatusPermanentRedirect || response.Header().Get("Location") != "/releases/v0.2.2" {
		t.Fatalf("redirect status=%d location=%q", response.Code, response.Header().Get("Location"))
	}

	sitemapRequest := httptest.NewRequest(http.MethodGet, "/sitemap.xml", nil)
	sitemapResponse := httptest.NewRecorder()
	seo.Sitemap(sitemapResponse, sitemapRequest)
	body := sitemapResponse.Body.String()
	if contentType := sitemapResponse.Header().Get("Content-Type"); contentType != "application/xml; charset=utf-8" {
		t.Fatalf("sitemap content type = %q", contentType)
	}
	var decoded sitemapURLSet
	if err := xml.Unmarshal(sitemapResponse.Body.Bytes(), &decoded); err != nil {
		t.Fatalf("sitemap is not valid XML: %v\n%s", err, body)
	}
	if decoded.XMLNS != "http://www.sitemaps.org/schemas/sitemap/0.9" || len(decoded.URLs) == 0 {
		t.Fatalf("invalid sitemap root: %#v", decoded)
	}
	if strings.Contains(strings.ToLower(body), "<!doctype html") || strings.Contains(strings.ToLower(body), "<html") {
		t.Fatalf("sitemap fell back to the SPA document: %s", body)
	}
	if !strings.Contains(body, "https://wave.example/releases/v0.2.2") || strings.Contains(body, "https://wave.example/blog/v0.2.2") {
		t.Fatalf("sitemap uses a non-canonical release URL: %s", body)
	}
}

func TestBlogMetadataIncludesVisibleComments(t *testing.T) {
	service, closeDatabase := testSEOService(t)
	defer closeDatabase()
	now := time.Date(2026, time.August, 26, 4, 0, 0, 0, time.UTC)
	if err := service.Repository().Upsert(blogdomain.Post{Slug: "comments", Category: "article", CommentPolicy: "open",
		Title: "Comments", Summary: "Discussion", Content: "Article", Status: "published", AuthorName: "Wave Foundation",
		PublishedAt: "2026-08-26", CreatedAt: now, UpdatedAt: now}); err != nil {
		t.Fatal(err)
	}
	if err := service.Repository().AddComment(blogdomain.Comment{ID: "visible", PostSlug: "comments", AuthorAccountID: "member",
		AuthorName: "Member", Body: "Useful compiler note.", Status: "visible", CreatedAt: now, UpdatedAt: now}); err != nil {
		t.Fatal(err)
	}
	if err := service.Repository().AddComment(blogdomain.Comment{ID: "hidden", PostSlug: "comments", AuthorAccountID: "member",
		AuthorName: "Member", Body: "Hidden note.", Status: "hidden", CreatedAt: now, UpdatedAt: now}); err != nil {
		t.Fatal(err)
	}
	graph := metadataGraph(t, NewSEOHandler("https://wave.example", nil, service, nil, nil, nil).HTMLMetadata(
		httptest.NewRequest(http.MethodGet, "/blog/comments", nil)))
	article := graphNode(t, graph, "BlogPosting")
	if article["commentCount"] != float64(1) {
		t.Fatalf("commentCount=%#v", article["commentCount"])
	}
	comments, _ := article["comment"].([]any)
	if len(comments) != 1 || strings.Contains(strings.ToLower(fmt.Sprint(comments)), "hidden note") {
		t.Fatalf("comments=%#v", comments)
	}
}

func TestMissingBlogPostIsNotIndexable(t *testing.T) {
	service, closeDatabase := testSEOService(t)
	defer closeDatabase()
	seo := NewSEOHandler("https://wave.example", nil, service, nil, nil, nil)
	metadata := seo.HTMLMetadata(httptest.NewRequest(http.MethodGet, "/blog/missing", nil))
	if !strings.Contains(metadata, `content="noindex, nofollow, noarchive"`) || !strings.Contains(metadata, `<title>Page not found · Wave</title>`) {
		t.Fatalf("missing page metadata = %s", metadata)
	}
	if strings.Contains(metadata, `application/ld+json`) {
		t.Fatal("non-indexable missing pages must not emit structured data")
	}
	request := httptest.NewRequest(http.MethodGet, "/blog/missing", nil)
	if status := seo.StatusCode(request); status != http.StatusNotFound {
		t.Fatalf("missing post status = %d", status)
	}
}

func TestMarkdownImageOnlyAcceptsWebURLs(t *testing.T) {
	if image, _ := markdownImage(`![unsafe](javascript:alert(1))`, "https://wave.example/blog/post"); image != "" {
		t.Fatalf("unsafe image URL = %q", image)
	}
	image, alt := markdownImage(`![Diagram](../media/diagram.webp)`, "https://wave.example/blog/post")
	if image != "https://wave.example/media/diagram.webp" || alt != "Diagram" {
		t.Fatalf("resolved image=%q alt=%q", image, alt)
	}
}

func TestDocumentationSEOUsesCanonicalLocalePaths(t *testing.T) {
	database, err := storage.Open(t.TempDir())
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	defer func() { _ = database.Close() }()
	if _, err := documentdomain.SeedOfficial(database); err != nil {
		t.Fatalf("seed documentation: %v", err)
	}
	seo := NewSEOHandler("https://wave.example", documentdomain.NewRepository(database), nil, nil, nil, nil)

	metadata := seo.HTMLMetadata(httptest.NewRequest(http.MethodGet, "/docs/ko/language/explicit-memory-type-model", nil))
	if !strings.Contains(metadata, `rel="canonical" href="https://wave.example/docs/ko/language/explicit-memory-type-model"`) {
		t.Fatalf("localized canonical is missing: %s", metadata)
	}
	article := graphNode(t, metadataGraph(t, metadata), "TechArticle")
	if article["inLanguage"] != "ko" || article["dateModified"] == "" {
		t.Fatalf("localized article schema = %#v", article)
	}
	japanese := seo.HTMLMetadata(httptest.NewRequest(http.MethodGet, "/docs/ja/language/explicit-memory-type-model", nil))
	if !strings.Contains(japanese, `rel="canonical" href="https://wave.example/docs/ja/language/explicit-memory-type-model"`) {
		t.Fatalf("Japanese localized canonical is missing: %s", japanese)
	}
	if graphNode(t, metadataGraph(t, japanese), "TechArticle")["inLanguage"] != "ja" {
		t.Fatal("Japanese article schema must identify its content language")
	}

	fallback := seo.HTMLMetadata(httptest.NewRequest(http.MethodGet, "/docs/zh/language/explicit-memory-type-model", nil))
	if !strings.Contains(fallback, `rel="canonical" href="https://wave.example/docs/en/language/explicit-memory-type-model"`) {
		t.Fatalf("English fallback canonical is missing: %s", fallback)
	}
	if graphNode(t, metadataGraph(t, fallback), "TechArticle")["inLanguage"] != "en" {
		t.Fatal("English fallback must identify its content language")
	}
	if redirect := seo.CanonicalRedirect("/docs/language/explicit-memory-type-model"); redirect != "/docs/en/language/explicit-memory-type-model" {
		t.Fatalf("legacy documentation redirect = %q", redirect)
	}

	sitemapRequest := httptest.NewRequest(http.MethodGet, "/sitemap.xml", nil)
	sitemapResponse := httptest.NewRecorder()
	seo.Sitemap(sitemapResponse, sitemapRequest)
	if !strings.Contains(sitemapResponse.Body.String(), "https://wave.example/docs/en/language/explicit-memory-type-model") {
		t.Fatalf("sitemap does not use localized canonical paths: %s", sitemapResponse.Body.String())
	}
}

func testSEOService(t *testing.T) (*blogdomain.Service, func()) {
	t.Helper()
	database, err := storage.Open(t.TempDir())
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	return blogdomain.NewService(database), func() { _ = database.Close() }
}

func metadataGraph(t *testing.T, metadata string) []any {
	t.Helper()
	start := strings.Index(metadata, `data-wave-schema="true">`)
	if start < 0 {
		t.Fatal("schema script is missing")
	}
	start += len(`data-wave-schema="true">`)
	end := strings.Index(metadata[start:], "</script>")
	if end < 0 {
		t.Fatal("schema script is not closed")
	}
	var document struct {
		Graph []any `json:"@graph"`
	}
	if err := json.Unmarshal([]byte(metadata[start:start+end]), &document); err != nil {
		t.Fatalf("decode schema: %v", err)
	}
	return document.Graph
}

func graphNode(t *testing.T, graph []any, schemaType string) map[string]any {
	t.Helper()
	for _, value := range graph {
		node, _ := value.(map[string]any)
		if node["@type"] == schemaType {
			return node
		}
	}
	t.Fatalf("schema node %q not found in %#v", schemaType, graph)
	return nil
}
