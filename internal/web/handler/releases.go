package handler

import (
	"encoding/xml"
	"errors"
	"net/http"
	"strconv"

	releasedomain "github.com/wavefnd/wave-platform/internal/release"
	"github.com/wavefnd/wave-platform/internal/storage"
	"github.com/wavefnd/wave-platform/internal/xmlcodec"
)

type ReleasesResponse struct {
	XMLName xml.Name                `xml:"https://wave-lang.dev/ns/platform/api/v1 releases"`
	Items   []releasedomain.Summary `xml:"release"`
}

type ReleasesHandler struct {
	Repository *releasedomain.Repository
}

func (handler ReleasesHandler) List(writer http.ResponseWriter, request *http.Request) {
	if handler.Repository == nil {
		http.Error(writer, "release repository is unavailable", http.StatusServiceUnavailable)
		return
	}
	limit, _ := strconv.Atoi(request.URL.Query().Get("limit"))
	items, err := handler.Repository.Releases(limit)
	if err != nil {
		http.Error(writer, "failed to load releases", http.StatusInternalServerError)
		return
	}
	summaries := make([]releasedomain.Summary, 0, len(items))
	for _, item := range items {
		summaries = append(summaries, releasedomain.Summary{
			Slug: item.Slug, Title: item.Title, PublishedAt: item.PublishedAt, Summary: item.Summary,
		})
	}
	if err := xmlcodec.Write(writer, http.StatusOK, ReleasesResponse{Items: summaries}); err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
	}
}

func (handler ReleasesHandler) Get(writer http.ResponseWriter, request *http.Request) {
	if handler.Repository == nil {
		http.Error(writer, "release repository is unavailable", http.StatusServiceUnavailable)
		return
	}
	item, err := handler.Repository.Release(request.PathValue("slug"))
	if errors.Is(err, storage.ErrNotFound) {
		http.NotFound(writer, request)
		return
	}
	if err != nil {
		http.Error(writer, "failed to load release", http.StatusInternalServerError)
		return
	}
	if err := xmlcodec.Write(writer, http.StatusOK, item); err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
	}
}
