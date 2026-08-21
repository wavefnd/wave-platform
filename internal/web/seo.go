package web

import (
	"encoding/json"
	"encoding/xml"
	"html"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	blogdomain "github.com/wavefnd/wave-platform/internal/blog"
	communitydomain "github.com/wavefnd/wave-platform/internal/community"
	documentdomain "github.com/wavefnd/wave-platform/internal/document"
	questiondomain "github.com/wavefnd/wave-platform/internal/question"
)

type sitemapURL struct {
	Location     string `xml:"loc"`
	LastModified string `xml:"lastmod,omitempty"`
}

type sitemapURLSet struct {
	XMLName xml.Name     `xml:"urlset"`
	XMLNS   string       `xml:"xmlns,attr"`
	URLs    []sitemapURL `xml:"url"`
}

type SEOHandler struct {
	publicURL string
	documents *documentdomain.Repository
	blog      *blogdomain.Service
	community *communitydomain.Repository
	questions *questiondomain.Repository
}

type pageMetadata struct {
	Title       string
	Description string
	Canonical   string
	Robots      string
	OpenGraph   string
	SchemaType  string
}

func NewSEOHandler(
	publicURL string,
	documents *documentdomain.Repository,
	blog *blogdomain.Service,
	community *communitydomain.Repository,
	questions *questiondomain.Repository,
) SEOHandler {
	return SEOHandler{
		publicURL: strings.TrimRight(publicURL, "/"), documents: documents,
		blog: blog, community: community, questions: questions,
	}
}

func (handler SEOHandler) Robots(writer http.ResponseWriter, request *http.Request) {
	base := handler.baseURL(request)
	writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
	writer.Header().Set("Cache-Control", "public, max-age=3600")
	_, _ = writer.Write([]byte("User-agent: *\nAllow: /\nDisallow: /admin\nDisallow: /account\nDisallow: /login\nDisallow: /register\nDisallow: /mail\nDisallow: /api/\n\nSitemap: " + base + "/sitemap.xml\n"))
}

func (handler SEOHandler) Sitemap(writer http.ResponseWriter, request *http.Request) {
	base := handler.baseURL(request)
	entries := []sitemapURL{
		{Location: base + "/"},
		{Location: base + "/docs"},
		{Location: base + "/blog"},
		{Location: base + "/releases"},
		{Location: base + "/community"},
		{Location: base + "/community/showcase"},
		{Location: base + "/lunastev"},
		{Location: base + "/questions"},
		{Location: base + "/source"},
		{Location: base + "/patches"},
	}

	if handler.documents != nil {
		if documents, err := handler.documents.Summaries("en"); err == nil {
			for _, document := range documents {
				entries = append(entries, sitemapURL{Location: handler.location(base, "docs", document.Path)})
			}
		}
	}
	if handler.blog != nil {
		if posts, err := handler.blog.Repository().Posts(false, 0); err == nil {
			for _, post := range posts {
				section := "blog"
				if post.Category == "release" {
					section = "releases"
				}
				entries = append(entries, sitemapURL{Location: handler.location(base, section, post.Slug), LastModified: dateOnly(post.UpdatedAt.Format(time.RFC3339))})
			}
		}
	}
	if handler.community != nil {
		if threads, err := handler.community.Threads("", 0); err == nil {
			for _, thread := range threads {
				section := "community"
				pathSegments := []string{section, "thread", thread.ID}
				if thread.SpaceID == "founder-notes" || thread.SpaceID == "development-log" {
					section = "lunastev"
					pathSegments = []string{section, "thread", thread.ID}
				} else if thread.SpaceID == "showcase" {
					pathSegments = []string{"community", "showcase", thread.ID}
				}
				entries = append(entries, sitemapURL{Location: handler.location(base, pathSegments...), LastModified: dateOnly(thread.LastActivityAt)})
			}
		}
	}
	if handler.questions != nil {
		if questions, err := handler.questions.Query("", "latest", "", 0, 0, ""); err == nil {
			for _, question := range questions {
				entries = append(entries, sitemapURL{Location: handler.location(base, "questions", question.ID), LastModified: dateOnly(question.LastActivityAt)})
			}
		}
	}

	data, err := xml.MarshalIndent(sitemapURLSet{
		XMLNS: "http://www.sitemaps.org/schemas/sitemap/0.9",
		URLs:  entries,
	}, "", "  ")
	if err != nil {
		http.Error(writer, "cannot create sitemap", http.StatusInternalServerError)
		return
	}
	writer.Header().Set("Content-Type", "application/xml; charset=utf-8")
	writer.Header().Set("Cache-Control", "public, max-age=900")
	_, _ = writer.Write(append([]byte(xml.Header), data...))
}

func (handler SEOHandler) HTMLMetadata(request *http.Request) string {
	metadata := handler.metadata(request.URL.Path, handler.baseURL(request))
	schema, _ := json.Marshal(map[string]any{
		"@context":    "https://schema.org",
		"@type":       metadata.SchemaType,
		"name":        metadata.Title,
		"description": metadata.Description,
		"url":         metadata.Canonical,
		"isPartOf": map[string]any{
			"@type": "WebSite", "name": "Wave Programming Language", "url": handler.baseURL(request),
		},
	})
	escape := html.EscapeString
	return `<meta name="description" content="` + escape(metadata.Description) + `" />
    <meta name="robots" content="` + escape(metadata.Robots) + `" />
    <meta property="og:title" content="` + escape(metadata.Title) + `" />
    <meta property="og:description" content="` + escape(metadata.Description) + `" />
    <meta property="og:type" content="` + escape(metadata.OpenGraph) + `" />
    <meta property="og:url" content="` + escape(metadata.Canonical) + `" />
    <meta property="og:site_name" content="Wave" />
    <meta name="twitter:card" content="summary" />
    <meta name="twitter:title" content="` + escape(metadata.Title) + `" />
    <meta name="twitter:description" content="` + escape(metadata.Description) + `" />
    <link rel="canonical" href="` + escape(metadata.Canonical) + `" />
    <script type="application/ld+json" data-wave-schema="true">` + string(schema) + `</script>
    <title>` + escape(metadata.Title) + `</title>`
}

func (handler SEOHandler) metadata(requestPath, base string) pageMetadata {
	metadata := pageMetadata{
		Title: "Wave Programming Language", Description: "The official home for the Wave programming language, documentation, releases, community, questions, source, and mail.",
		Canonical: canonicalLocation(base, requestPath), Robots: "index, follow, max-image-preview:large", OpenGraph: "website", SchemaType: "WebPage",
	}
	segments := strings.Split(strings.Trim(requestPath, "/"), "/")
	first := ""
	if len(segments) > 0 {
		first = segments[0]
	}
	switch first {
	case "":
		metadata.SchemaType = "WebSite"
	case "docs":
		metadata.Title = "Documentation · Wave"
		metadata.Description = "Official Wave programming language guides and reference documentation."
		metadata.SchemaType = "CollectionPage"
		if len(segments) > 1 && handler.documents != nil {
			if document, err := handler.documents.Published("en", strings.Join(segments[1:], "/")); err == nil {
				metadata.Title = document.Title + " · Wave Documentation"
				metadata.Description = document.Summary.Summary
				metadata.OpenGraph = "article"
				metadata.SchemaType = "TechArticle"
			}
		}
	case "blog":
		metadata.Title = "Blog · Wave"
		metadata.Description = "Official Wave programming language news, engineering updates, and release articles."
		metadata.SchemaType = "CollectionPage"
		if len(segments) > 1 && handler.blog != nil {
			if post, err := handler.blog.Repository().Post(segments[1], false); err == nil {
				metadata.Title = post.Title + " · Wave Blog"
				metadata.Description = post.Summary
				metadata.OpenGraph = "article"
				metadata.SchemaType = "Article"
				if post.Category == "release" {
					metadata.Canonical = handler.location(base, "releases", post.Slug)
				}
			}
		}
	case "releases":
		metadata.Title = "Wave Releases"
		metadata.Description = "Published Wave versions, release dates, and the changes included in each version."
		metadata.SchemaType = "CollectionPage"
		if len(segments) > 1 && handler.blog != nil {
			if post, err := handler.blog.Repository().Post(segments[1], false); err == nil && post.Category == "release" {
				metadata.Title = post.Title + " · Wave Releases"
				metadata.Description = post.Summary
				metadata.OpenGraph = "article"
				metadata.SchemaType = "TechArticle"
			}
		}
	case "community", "lunastev":
		personal := first == "lunastev"
		showcase := !personal && len(segments) > 1 && segments[1] == "showcase"
		metadata.Title = map[bool]string{true: "LunaStev · Wave", false: "Community · Wave"}[personal]
		metadata.Description = "Wave programming language community posts and technical discussions."
		metadata.SchemaType = "CollectionPage"
		if showcase {
			metadata.Title = "Showcase · Wave"
			metadata.Description = "Tools, applications, experiments, and systems built by the Wave community."
		}
		if len(segments) == 3 && (segments[1] == "thread" || segments[1] == "showcase") && handler.community != nil {
			if threads, err := handler.community.Threads("", 0); err == nil {
				for _, thread := range threads {
					if thread.ID != segments[2] {
						continue
					}
					suffix := map[bool]string{true: " · LunaStev", false: " · Wave Community"}[personal]
					if showcase {
						suffix = " · Wave Showcase"
					}
					metadata.Title = thread.Title + suffix
					metadata.Description = thread.Excerpt
					metadata.OpenGraph = "article"
					metadata.SchemaType = "WebPage"
					break
				}
			}
		}
	case "questions":
		metadata.Title = "Questions · Wave"
		metadata.Description = "Technical questions and answers about the Wave programming language."
		metadata.SchemaType = "CollectionPage"
		if len(segments) == 2 && handler.questions != nil {
			if questions, err := handler.questions.Query("", "latest", "", 0, 0, ""); err == nil {
				for _, question := range questions {
					if question.ID != segments[1] {
						continue
					}
					metadata.Title = question.Title + " · Wave Questions"
					metadata.Description = question.Excerpt
					metadata.SchemaType = "WebPage"
					break
				}
			}
		}
	case "source":
		metadata.Title = "Source · Wave"
		metadata.Description = "Read-only source browser for official Wave Git mirrors."
		metadata.SchemaType = "CollectionPage"
		if len(segments) > 1 {
			metadata.Title = segments[1] + " · Wave Source"
			metadata.SchemaType = "SoftwareSourceCode"
		}
	case "mail", "account", "login", "register", "admin", "search":
		metadata.Title = "Wave"
		metadata.Description = "Wave Platform."
		metadata.Robots = "noindex, nofollow, noarchive"
		metadata.SchemaType = "WebPage"
	default:
		metadata.Title = "Page not found · Wave"
		metadata.Description = "The requested Wave page was not found."
		metadata.Robots = "noindex, nofollow, noarchive"
	}
	if strings.HasSuffix(requestPath, "/new") {
		metadata.Robots = "noindex, nofollow, noarchive"
	}
	return metadata
}

func (handler SEOHandler) baseURL(request *http.Request) string {
	if parsed, err := url.Parse(handler.publicURL); err == nil && parsed.Scheme != "" && parsed.Host != "" {
		return strings.TrimRight(parsed.String(), "/")
	}
	scheme := request.Header.Get("X-Forwarded-Proto")
	if scheme != "http" && scheme != "https" {
		scheme = "https"
	}
	return scheme + "://" + request.Host
}

func (handler SEOHandler) location(base string, elements ...string) string {
	parsed, err := url.Parse(base)
	if err != nil {
		return base
	}
	parts := []string{parsed.Path}
	parts = append(parts, elements...)
	parsed.Path = path.Join(parts...)
	return parsed.String()
}

func canonicalLocation(base, requestPath string) string {
	if requestPath == "" || requestPath == "/" {
		return strings.TrimRight(base, "/") + "/"
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return base
	}
	parsed.Path = path.Join(parsed.Path, requestPath)
	parsed.RawQuery = ""
	parsed.Fragment = ""
	return parsed.String()
}

func dateOnly(value string) string {
	value = strings.TrimSpace(value)
	if len(value) >= len("2006-01-02") {
		return value[:len("2006-01-02")]
	}
	return ""
}
