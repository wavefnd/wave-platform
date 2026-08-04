package handler

import (
	"encoding/xml"
	"errors"
	"net/http"
	"strings"

	documentdomain "github.com/wavefnd/wave-platform/internal/document"
	"github.com/wavefnd/wave-platform/internal/storage"
	"github.com/wavefnd/wave-platform/internal/xmlcodec"
)

type DocumentsResponse struct {
	XMLName xml.Name                 `xml:"https://wave-lang.dev/ns/platform/api/v1 documents"`
	Items   []documentdomain.Summary `xml:"document"`
}

type DocumentsHandler struct {
	Repository *documentdomain.Repository
}

func (handler DocumentsHandler) List(writer http.ResponseWriter, request *http.Request) {
	if handler.Repository == nil {
		http.Error(writer, "document repository is unavailable", http.StatusServiceUnavailable)
		return
	}
	locale, ok := documentLocale(writer, request)
	if !ok {
		return
	}
	items, err := handler.Repository.Summaries(locale)
	if err != nil {
		http.Error(writer, "failed to load documents", http.StatusInternalServerError)
		return
	}
	if err := xmlcodec.Write(writer, http.StatusOK, DocumentsResponse{Items: items}); err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
	}
}

func (handler DocumentsHandler) Get(writer http.ResponseWriter, request *http.Request) {
	if handler.Repository == nil {
		http.Error(writer, "document repository is unavailable", http.StatusServiceUnavailable)
		return
	}
	locale, ok := documentLocale(writer, request)
	if !ok {
		return
	}
	path := strings.Trim(request.PathValue("path"), "/")
	if path == "" || strings.Contains(path, "..") {
		writeAPIError(writer, http.StatusBadRequest, "invalid-document-path", "The document path is invalid.")
		return
	}
	item, err := handler.Repository.Published(locale, path)
	if errors.Is(err, storage.ErrNotFound) {
		http.NotFound(writer, request)
		return
	}
	if err != nil {
		http.Error(writer, "failed to load document", http.StatusInternalServerError)
		return
	}
	if err := xmlcodec.Write(writer, http.StatusOK, item); err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
	}
}

func documentLocale(writer http.ResponseWriter, request *http.Request) (string, bool) {
	locale := strings.ToLower(strings.TrimSpace(request.URL.Query().Get("locale")))
	if locale == "" {
		locale = "en"
	}
	if locale != "en" && locale != "ko" {
		writeAPIError(writer, http.StatusBadRequest, "unsupported-locale", "Only English and Korean documents are available.")
		return "", false
	}
	return locale, true
}
