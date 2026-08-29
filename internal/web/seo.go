package web

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"html"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"

	blogdomain "github.com/wavefnd/wave-platform/internal/blog"
	communitydomain "github.com/wavefnd/wave-platform/internal/community"
	documentdomain "github.com/wavefnd/wave-platform/internal/document"
	questiondomain "github.com/wavefnd/wave-platform/internal/question"
	rfcdomain "github.com/wavefnd/wave-platform/internal/rfc"
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
	rfcs      *rfcdomain.Service
}

type pageMetadata struct {
	Title          string
	Headline       string
	Description    string
	Canonical      string
	Robots         string
	OpenGraph      string
	SchemaType     string
	AuthorName     string
	AuthorURL      string
	PublishedAt    string
	ModifiedAt     string
	ArticleSection string
	Image          string
	ImageAlt       string
	Breadcrumbs    []seoBreadcrumb
	Items          []seoItem
	Comments       []seoComment
	CommentCount   int
	Language       string
}

type seoBreadcrumb struct {
	Name string
	URL  string
}

type seoItem struct {
	Name string
	URL  string
}

type seoComment struct {
	AuthorName string
	AuthorURL  string
	Text       string
	CreatedAt  string
}

var markdownImagePattern = regexp.MustCompile(`!\[([^\]]*)\]\((?:<([^>]+)>|([^\s)]+))(?:\s+["'][^)]*["'])?\)`)

func NewSEOHandler(
	publicURL string,
	documents *documentdomain.Repository,
	blog *blogdomain.Service,
	community *communitydomain.Repository,
	questions *questiondomain.Repository,
	rfcs *rfcdomain.Service,
) SEOHandler {
	return SEOHandler{
		publicURL: strings.TrimRight(publicURL, "/"), documents: documents,
		blog: blog, community: community, questions: questions, rfcs: rfcs,
	}
}

func (handler SEOHandler) Robots(writer http.ResponseWriter, request *http.Request) {
	base := handler.baseURL(request)
	writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
	writer.Header().Set("Cache-Control", "public, max-age=3600")
	_, _ = writer.Write([]byte("User-agent: *\nAllow: /\nDisallow: /admin\nDisallow: /account\nDisallow: /login\nDisallow: /register\nDisallow: /mail\nDisallow: /blog/editor\nDisallow: /api/\n\nSitemap: " + base + "/sitemap.xml\n"))
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
		{Location: base + "/rfcs"},
		{Location: base + "/source"},
	}

	if handler.documents != nil {
		if documents, err := handler.documents.Summaries("en"); err == nil {
			for _, document := range documents {
				entries = append(entries, sitemapURL{Location: handler.location(base, "docs", "en", document.Path)})
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
				lastModified := ""
				if !post.UpdatedAt.IsZero() {
					lastModified = dateOnly(post.UpdatedAt.Format(time.RFC3339))
				}
				entries = append(entries, sitemapURL{Location: handler.location(base, section, post.Slug), LastModified: lastModified})
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
	if handler.rfcs != nil {
		if proposals, err := handler.rfcs.Repository().Proposals("", ""); err == nil {
			for _, proposal := range proposals {
				entries = append(entries, sitemapURL{Location: handler.location(base, "rfcs", strconv.FormatUint(proposal.Number, 10)), LastModified: dateOnly(proposal.UpdatedAt.Format(time.RFC3339))})
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
	base := strings.TrimRight(handler.baseURL(request), "/")
	organizationID := base + "/#organization"
	websiteID := base + "/#website"
	pageID := metadata.Canonical + "#webpage"
	organization := map[string]any{
		"@type": "Organization", "@id": organizationID, "name": "Wave Foundation", "url": base + "/",
		"logo":   map[string]any{"@type": "ImageObject", "url": base + "/img/wave-logo.ico", "width": 256, "height": 256},
		"sameAs": []string{"https://github.com/wavefnd", "https://discord.gg/3nev5nHqq9", "https://opencollective.com/wave-lang"},
	}
	website := map[string]any{
		"@type": "WebSite", "@id": websiteID, "name": "Wave Programming Language", "url": base + "/",
		"publisher": map[string]any{"@id": organizationID}, "inLanguage": []string{"en", "ko", "ja", "zh", "es", "de", "ru", "id", "vi"},
	}
	pageType := metadata.SchemaType
	if isArticleSchema(pageType) || pageType == "WebSite" {
		pageType = "WebPage"
	}
	pageSchema := map[string]any{
		"@type": pageType, "@id": pageID, "url": metadata.Canonical, "name": metadata.Title,
		"description": metadata.Description, "isPartOf": map[string]any{"@id": websiteID}, "inLanguage": metadata.Language,
	}
	graph := []any{organization, website, pageSchema}
	if len(metadata.Breadcrumbs) >= 2 {
		breadcrumbID := metadata.Canonical + "#breadcrumb"
		items := make([]map[string]any, 0, len(metadata.Breadcrumbs))
		for index, breadcrumb := range metadata.Breadcrumbs {
			items = append(items, map[string]any{
				"@type": "ListItem", "position": index + 1, "name": breadcrumb.Name, "item": breadcrumb.URL,
			})
		}
		graph = append(graph, map[string]any{"@type": "BreadcrumbList", "@id": breadcrumbID, "itemListElement": items})
		pageSchema["breadcrumb"] = map[string]any{"@id": breadcrumbID}
	}
	if len(metadata.Items) > 0 {
		listID := metadata.Canonical + "#items"
		items := make([]map[string]any, 0, len(metadata.Items))
		for index, item := range metadata.Items {
			items = append(items, map[string]any{
				"@type": "ListItem", "position": index + 1, "name": item.Name, "url": item.URL,
			})
		}
		graph = append(graph, map[string]any{"@type": "ItemList", "@id": listID, "itemListElement": items})
		pageSchema["mainEntity"] = map[string]any{"@id": listID}
	}
	if isArticleSchema(metadata.SchemaType) {
		articleID := metadata.Canonical + "#article"
		headline := metadata.Headline
		if headline == "" {
			headline = metadata.Title
		}
		article := map[string]any{
			"@type": metadata.SchemaType, "@id": articleID, "url": metadata.Canonical, "headline": headline,
			"description": metadata.Description, "mainEntityOfPage": map[string]any{"@id": pageID},
			"publisher": map[string]any{"@id": organizationID}, "isAccessibleForFree": true, "inLanguage": metadata.Language,
		}
		if metadata.AuthorName != "" {
			authorType := "Person"
			if metadata.AuthorName == "Wave Foundation" {
				authorType = "Organization"
			}
			author := map[string]any{"@type": authorType, "name": metadata.AuthorName}
			if metadata.AuthorURL != "" {
				author["url"] = metadata.AuthorURL
			}
			article["author"] = author
		}
		if metadata.PublishedAt != "" {
			article["datePublished"] = metadata.PublishedAt
		}
		if metadata.ModifiedAt != "" {
			article["dateModified"] = metadata.ModifiedAt
		}
		if metadata.ArticleSection != "" {
			article["articleSection"] = metadata.ArticleSection
		}
		if metadata.Image != "" {
			article["image"] = []string{metadata.Image}
		}
		if metadata.SchemaType == "BlogPosting" {
			article["commentCount"] = metadata.CommentCount
			comments := make([]map[string]any, 0, len(metadata.Comments))
			for _, comment := range metadata.Comments {
				author := map[string]any{"@type": "Person", "name": comment.AuthorName}
				if comment.AuthorURL != "" {
					author["url"] = comment.AuthorURL
				}
				comments = append(comments, map[string]any{"@type": "Comment", "text": comment.Text,
					"dateCreated": comment.CreatedAt, "author": author})
			}
			if len(comments) > 0 {
				article["comment"] = comments
			}
		}
		graph = append(graph, article)
		pageSchema["mainEntity"] = map[string]any{"@id": articleID}
	}
	schema, _ := json.Marshal(map[string]any{"@context": "https://schema.org", "@graph": graph})
	escape := html.EscapeString
	result := `<meta name="description" content="` + escape(metadata.Description) + `" />
    <meta name="robots" content="` + escape(metadata.Robots) + `" />
    <meta property="og:title" content="` + escape(metadata.Title) + `" />
    <meta property="og:description" content="` + escape(metadata.Description) + `" />
    <meta property="og:type" content="` + escape(metadata.OpenGraph) + `" />
    <meta property="og:url" content="` + escape(metadata.Canonical) + `" />
    <meta property="og:site_name" content="Wave" />
    <meta property="og:locale" content="en_US" />
    <meta name="twitter:card" content="` + map[bool]string{true: "summary_large_image", false: "summary"}[metadata.Image != ""] + `" />
    <meta name="twitter:title" content="` + escape(metadata.Title) + `" />
    <meta name="twitter:description" content="` + escape(metadata.Description) + `" />
`
	if metadata.Image != "" {
		result += `    <meta property="og:image" content="` + escape(metadata.Image) + `" />
    <meta property="og:image:alt" content="` + escape(metadata.ImageAlt) + `" />
    <meta name="twitter:image" content="` + escape(metadata.Image) + `" />
    <meta name="twitter:image:alt" content="` + escape(metadata.ImageAlt) + `" />
`
	}
	if metadata.PublishedAt != "" {
		result += `    <meta property="article:published_time" content="` + escape(metadata.PublishedAt) + `" />
`
	}
	if metadata.ModifiedAt != "" {
		result += `    <meta property="article:modified_time" content="` + escape(metadata.ModifiedAt) + `" />
`
	}
	if metadata.ArticleSection != "" {
		result += `    <meta property="article:section" content="` + escape(metadata.ArticleSection) + `" />
`
	}
	if metadata.AuthorName != "" {
		result += `    <meta property="article:author" content="` + escape(metadata.AuthorName) + `" />
`
	}
	schemaMarkup := ""
	if !strings.HasPrefix(metadata.Robots, "noindex") {
		schemaMarkup = `    <script type="application/ld+json" data-wave-schema="true">` + string(schema) + `</script>
`
	}
	return result + `    <link rel="canonical" href="` + escape(metadata.Canonical) + `" />
` + schemaMarkup + `    <title>` + escape(metadata.Title) + `</title>`
}

func (handler SEOHandler) StatusCode(request *http.Request) int {
	segments := strings.Split(strings.Trim(request.URL.Path, "/"), "/")
	if len(segments) == 0 || segments[0] == "" {
		return http.StatusOK
	}
	switch segments[0] {
	case "blog":
		if len(segments) == 1 || len(segments) >= 2 && segments[1] == "editor" || handler.blog == nil {
			return http.StatusOK
		}
		if len(segments) != 2 {
			return http.StatusNotFound
		}
		if _, err := handler.blog.Repository().Post(segments[1], false); err != nil {
			return http.StatusNotFound
		}
	case "releases":
		if len(segments) == 1 || handler.blog == nil {
			return http.StatusOK
		}
		if len(segments) != 2 {
			return http.StatusNotFound
		}
		post, err := handler.blog.Repository().Post(segments[1], false)
		if err != nil || post.Category != "release" {
			return http.StatusNotFound
		}
	case "docs", "community", "lunastev", "questions", "rfcs", "source", "mail", "account", "login", "register", "admin", "search", "user", "patches":
		return http.StatusOK
	default:
		return http.StatusNotFound
	}
	return http.StatusOK
}

func (handler SEOHandler) metadata(requestPath, base string) pageMetadata {
	home := seoBreadcrumb{Name: "Home", URL: strings.TrimRight(base, "/") + "/"}
	metadata := pageMetadata{
		Title: "Wave Programming Language", Description: "The official home for the Wave programming language, documentation, releases, community, questions, source, and mail.",
		Canonical: canonicalLocation(base, requestPath), Robots: "index, follow, max-image-preview:large", OpenGraph: "website", SchemaType: "WebPage",
		Language: languageForPath(requestPath),
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
		documentLocale := "en"
		documentStart := 1
		documentRoot := handler.location(base, "docs", documentLocale)
		if len(segments) > 1 && supportedDocumentLocale(segments[1]) {
			documentLocale = segments[1]
			documentStart = 2
			documentRoot = handler.location(base, "docs", documentLocale)
		}
		metadata.Language = documentLocale
		metadata.Breadcrumbs = []seoBreadcrumb{home, {Name: "Documentation", URL: documentRoot}}
		if len(segments) > documentStart && handler.documents != nil {
			documentPath := strings.Join(segments[documentStart:], "/")
			document, err := handler.documents.Published(documentLocale, documentPath)
			if err != nil && documentLocale != "en" {
				if fallback, fallbackErr := handler.documents.Published("en", documentPath); fallbackErr == nil {
					document = fallback
					err = nil
					metadata.Language = "en"
					metadata.Canonical = handler.location(base, "docs", "en", documentPath)
					metadata.Breadcrumbs = []seoBreadcrumb{home, {Name: "Documentation", URL: handler.location(base, "docs", "en")}}
				}
			}
			if err == nil {
				metadata.Title = document.Title + " · Wave Documentation"
				metadata.Headline = document.Title
				metadata.Description = seoDescription(document.Summary.Summary, document.Title)
				metadata.OpenGraph = "article"
				metadata.SchemaType = "TechArticle"
				metadata.AuthorName = "Wave Foundation"
				metadata.ModifiedAt = document.UpdatedAt
				metadata.ArticleSection = "Documentation"
				metadata.Breadcrumbs = append(metadata.Breadcrumbs, seoBreadcrumb{Name: document.Title, URL: metadata.Canonical})
			} else {
				metadata = notFoundMetadata(metadata)
			}
		}
	case "blog":
		metadata.Title = "Blog · Wave"
		metadata.Description = "Official Wave programming language news, engineering updates, and release articles."
		metadata.SchemaType = "CollectionPage"
		metadata.Breadcrumbs = []seoBreadcrumb{home, {Name: "Blog", URL: handler.location(base, "blog")}}
		metadata.Items = handler.blogItems(base, "blog")
		if len(segments) > 1 && handler.blog != nil {
			if segments[1] == "editor" {
				metadata.Title = "WaveEditor · Wave Blog"
				metadata.Description = "Wave Blog editor."
				metadata.Robots = "noindex, nofollow, noarchive"
				metadata.Items = nil
			} else if len(segments) == 2 {
				if post, err := handler.blog.Repository().Post(segments[1], false); err == nil {
					metadata = handler.blogPostMetadata(metadata, post, base)
				} else {
					metadata = notFoundMetadata(metadata)
				}
			} else {
				metadata = notFoundMetadata(metadata)
			}
		}
	case "releases":
		metadata.Title = "Wave Releases"
		metadata.Description = "Published Wave versions, release dates, and the changes included in each version."
		metadata.SchemaType = "CollectionPage"
		metadata.Breadcrumbs = []seoBreadcrumb{home, {Name: "Releases", URL: handler.location(base, "releases")}}
		metadata.Items = handler.blogItems(base, "releases")
		if len(segments) > 1 && handler.blog != nil {
			if post, err := handler.blog.Repository().Post(segments[1], false); len(segments) == 2 && err == nil && post.Category == "release" {
				metadata = handler.blogPostMetadata(metadata, post, base)
			} else {
				metadata = notFoundMetadata(metadata)
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
	case "rfcs":
		metadata.Title = "Request for Comments · Wave"
		metadata.Description = "Public design proposals and decisions for significant changes to the Wave language and platform."
		metadata.SchemaType = "CollectionPage"
		metadata.Breadcrumbs = []seoBreadcrumb{home, {Name: "RFCs", URL: handler.location(base, "rfcs")}}
		if len(segments) == 2 && handler.rfcs != nil {
			if number, err := strconv.ParseUint(segments[1], 10, 64); err == nil {
				if proposal, proposalErr := handler.rfcs.Repository().Proposal(number); proposalErr == nil {
					metadata.Title = fmt.Sprintf("RFC-%04d: %s · Wave", proposal.Number, proposal.Title)
					metadata.Headline = fmt.Sprintf("RFC-%04d: %s", proposal.Number, proposal.Title)
					metadata.Description = proposal.Summary
					metadata.OpenGraph = "article"
					metadata.SchemaType = "TechArticle"
					metadata.ArticleSection = "RFC"
					metadata.Breadcrumbs = append(metadata.Breadcrumbs, seoBreadcrumb{Name: fmt.Sprintf("RFC-%04d", proposal.Number), URL: metadata.Canonical})
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
	if strings.HasSuffix(requestPath, "/new") || strings.HasSuffix(requestPath, "/edit") {
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

func isArticleSchema(schemaType string) bool {
	switch schemaType {
	case "Article", "BlogPosting", "TechArticle":
		return true
	default:
		return false
	}
}

func notFoundMetadata(metadata pageMetadata) pageMetadata {
	metadata.Title = "Page not found · Wave"
	metadata.Headline = ""
	metadata.Description = "The requested Wave page was not found."
	metadata.Robots = "noindex, nofollow, noarchive"
	metadata.OpenGraph = "website"
	metadata.SchemaType = "WebPage"
	metadata.AuthorName = ""
	metadata.AuthorURL = ""
	metadata.PublishedAt = ""
	metadata.ModifiedAt = ""
	metadata.ArticleSection = ""
	metadata.Image = ""
	metadata.ImageAlt = ""
	metadata.Items = nil
	metadata.Comments = nil
	metadata.CommentCount = 0
	return metadata
}

func (handler SEOHandler) blogPostMetadata(metadata pageMetadata, post blogdomain.Post, base string) pageMetadata {
	sectionName := "Blog"
	sectionPath := "blog"
	metadata.SchemaType = "BlogPosting"
	metadata.Title = post.Title + " · Wave Blog"
	if post.Category == "release" {
		sectionName = "Releases"
		sectionPath = "releases"
		metadata.SchemaType = "TechArticle"
		metadata.Title = post.Title + " · Wave Releases"
	} else if post.Category == "roadmap" {
		sectionName = "Roadmap"
		metadata.SchemaType = "TechArticle"
	}
	metadata.Headline = post.Title
	metadata.Description = seoDescription(post.Summary, post.Title)
	metadata.Canonical = handler.location(base, sectionPath, post.Slug)
	metadata.OpenGraph = "article"
	metadata.AuthorName = strings.TrimSpace(post.AuthorName)
	if metadata.AuthorName == "" {
		metadata.AuthorName = "Wave Foundation"
	}
	if strings.TrimSpace(post.AuthorAccountID) != "" {
		metadata.AuthorURL = handler.location(base, "user", "id", post.AuthorAccountID)
	}
	metadata.PublishedAt = strings.TrimSpace(post.PublishedAt)
	if !post.UpdatedAt.IsZero() {
		metadata.ModifiedAt = post.UpdatedAt.UTC().Format(time.RFC3339)
	}
	metadata.ArticleSection = sectionName
	metadata.Items = nil
	metadata.Comments = nil
	metadata.CommentCount = 0
	metadata.Breadcrumbs = []seoBreadcrumb{
		{Name: "Home", URL: strings.TrimRight(base, "/") + "/"},
		{Name: sectionName, URL: handler.location(base, sectionPath)},
		{Name: post.Title, URL: metadata.Canonical},
	}
	metadata.Image, metadata.ImageAlt = markdownImage(post.Content, metadata.Canonical)
	if metadata.Image != "" && metadata.ImageAlt == "" {
		metadata.ImageAlt = post.Title
	}
	if metadata.SchemaType == "BlogPosting" && handler.blog != nil {
		if comments, err := handler.blog.Comments(post.Slug, false); err == nil {
			metadata.CommentCount = len(comments)
			for _, comment := range comments {
				metadata.Comments = append(metadata.Comments, seoComment{AuthorName: comment.AuthorName,
					AuthorURL: handler.location(base, "user", "id", comment.AuthorAccountID), Text: seoDescription(comment.Body, "Comment"),
					CreatedAt: comment.CreatedAt.UTC().Format(time.RFC3339)})
				if len(metadata.Comments) == 20 {
					break
				}
			}
		}
	}
	return metadata
}

func (handler SEOHandler) blogItems(base, section string) []seoItem {
	if handler.blog == nil {
		return nil
	}
	posts, err := handler.blog.Repository().Posts(false, 0)
	if err != nil {
		return nil
	}
	items := make([]seoItem, 0, len(posts))
	for _, post := range posts {
		postSection := "blog"
		if post.Category == "release" {
			postSection = "releases"
		}
		if section == "releases" && post.Category != "release" {
			continue
		}
		if section == "blog" && post.Category == "release" {
			continue
		}
		items = append(items, seoItem{Name: post.Title, URL: handler.location(base, postSection, post.Slug)})
		if len(items) == 20 {
			break
		}
	}
	return items
}

func markdownImage(markdown, canonical string) (string, string) {
	match := markdownImagePattern.FindStringSubmatch(markdown)
	if len(match) == 0 {
		return "", ""
	}
	rawURL := match[2]
	if rawURL == "" {
		rawURL = match[3]
	}
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil || (parsed.Scheme != "" && parsed.Scheme != "http" && parsed.Scheme != "https") {
		return "", ""
	}
	base, err := url.Parse(canonical)
	if err != nil {
		return "", ""
	}
	resolved := base.ResolveReference(parsed)
	if resolved.Scheme != "http" && resolved.Scheme != "https" || resolved.Host == "" {
		return "", ""
	}
	return resolved.String(), strings.TrimSpace(match[1])
}

func seoDescription(value, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		value = strings.TrimSpace(fallback)
	}
	value = strings.Join(strings.Fields(value), " ")
	runes := []rune(value)
	if len(runes) > 180 {
		return strings.TrimSpace(string(runes[:177])) + "..."
	}
	return value
}

func (handler SEOHandler) CanonicalRedirect(requestPath string) string {
	segments := strings.Split(strings.Trim(requestPath, "/"), "/")
	if len(segments) >= 2 && segments[0] == "docs" && !supportedDocumentLocale(segments[1]) && handler.documents != nil {
		documentPath := strings.Join(segments[1:], "/")
		if _, err := handler.documents.Published("en", documentPath); err == nil {
			return "/docs/en/" + strings.Join(segments[1:], "/")
		}
	}
	if len(segments) == 2 && segments[0] == "blog" && handler.blog != nil {
		post, err := handler.blog.Repository().Post(segments[1], false)
		if err == nil && post.Category == "release" {
			return "/releases/" + url.PathEscape(post.Slug)
		}
	}
	return ""
}

func languageForPath(requestPath string) string {
	segments := strings.Split(strings.Trim(requestPath, "/"), "/")
	if len(segments) >= 2 && segments[0] == "docs" && supportedDocumentLocale(segments[1]) {
		return segments[1]
	}
	return "en"
}

func supportedDocumentLocale(value string) bool {
	switch value {
	case "en", "ko", "ja", "zh", "es", "de", "ru", "id", "vi":
		return true
	default:
		return false
	}
}
