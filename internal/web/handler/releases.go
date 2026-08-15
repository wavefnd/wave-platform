package handler

import (
	"encoding/xml"
	"errors"
	"net/http"
	"strconv"

	blogdomain "github.com/wavefnd/wave-platform/internal/blog"
	"github.com/wavefnd/wave-platform/internal/storage"
	"github.com/wavefnd/wave-platform/internal/xmlcodec"
)

type ReleasesResponse struct {
	XMLName xml.Name         `xml:"https://wave-lang.dev/ns/platform/api/v1 releases"`
	Items   []ReleaseSummary `xml:"release"`
}

type ReleaseSummary struct {
	Slug        string `xml:"slug"`
	Title       string `xml:"title"`
	PublishedAt string `xml:"published-at"`
	Summary     string `xml:"summary"`
}

type ReleaseResponse struct {
	XMLName     xml.Name `xml:"https://wave-lang.dev/ns/platform/release/v1 release"`
	Slug        string   `xml:"slug"`
	Title       string   `xml:"title"`
	PublishedAt string   `xml:"published-at"`
	Summary     string   `xml:"summary"`
	Content     string   `xml:"content"`
}

type ReleasesHandler struct {
	Service *blogdomain.Service
}

func (handler ReleasesHandler) List(writer http.ResponseWriter, request *http.Request) {
	if handler.Service == nil {
		http.Error(writer, "release repository is unavailable", http.StatusServiceUnavailable)
		return
	}
	limit, _ := strconv.Atoi(request.URL.Query().Get("limit"))
	items, err := handler.Service.Repository().PostsByCategory("", "release", false, limit)
	if err != nil {
		http.Error(writer, "failed to load releases", http.StatusInternalServerError)
		return
	}
	summaries := make([]ReleaseSummary, 0, len(items))
	for _, item := range items {
		summaries = append(summaries, ReleaseSummary{
			Slug: item.Slug, Title: item.Title, PublishedAt: item.PublishedAt, Summary: item.Summary,
		})
	}
	if err := xmlcodec.Write(writer, http.StatusOK, ReleasesResponse{Items: summaries}); err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
	}
}

func (handler ReleasesHandler) Get(writer http.ResponseWriter, request *http.Request) {
	if handler.Service == nil {
		http.Error(writer, "release repository is unavailable", http.StatusServiceUnavailable)
		return
	}
	post, err := handler.Service.Repository().Post(request.PathValue("slug"), false)
	if errors.Is(err, storage.ErrNotFound) || (err == nil && post.Category != "release") {
		http.NotFound(writer, request)
		return
	}
	if err != nil {
		http.Error(writer, "failed to load release", http.StatusInternalServerError)
		return
	}
	item := ReleaseResponse{Slug: post.Slug, Title: post.Title, PublishedAt: post.PublishedAt,
		Summary: post.Summary, Content: post.Content}
	if err := xmlcodec.Write(writer, http.StatusOK, item); err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
	}
}
