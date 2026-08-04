package handler

import (
	"encoding/xml"
	"errors"
	"mime"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/wavefnd/wave-platform/internal/gitmirror"
	"github.com/wavefnd/wave-platform/internal/storage"
	"github.com/wavefnd/wave-platform/internal/xmlcodec"
)

type RepositoriesResponse struct {
	XMLName      xml.Name               `xml:"https://wave-lang.dev/ns/platform/api/v1 repositories"`
	Repositories []gitmirror.Repository `xml:"repository"`
}

type CommitsResponse struct {
	XMLName xml.Name           `xml:"https://wave-lang.dev/ns/platform/api/v1 commits"`
	Commits []gitmirror.Commit `xml:"commit"`
}

type SourceHandler struct {
	Service *gitmirror.Service
}

func (handler SourceHandler) Repositories(writer http.ResponseWriter, request *http.Request) {
	if handler.Service == nil {
		http.Error(writer, "source service is unavailable", http.StatusServiceUnavailable)
		return
	}
	repositories, err := handler.Service.Repositories(request.Context())
	if err != nil {
		http.Error(writer, "failed to load repositories", http.StatusInternalServerError)
		return
	}
	handler.write(writer, RepositoriesResponse{Repositories: repositories})
}

func (handler SourceHandler) Tree(writer http.ResponseWriter, request *http.Request) {
	if handler.Service == nil {
		http.Error(writer, "source service is unavailable", http.StatusServiceUnavailable)
		return
	}
	tree, err := handler.Service.Tree(request.Context(), request.PathValue("repository"), request.URL.Query().Get("ref"), request.URL.Query().Get("path"))
	if err != nil {
		handler.writeError(writer, request, err)
		return
	}
	handler.write(writer, tree)
}

func (handler SourceHandler) Blob(writer http.ResponseWriter, request *http.Request) {
	if handler.Service == nil {
		http.Error(writer, "source service is unavailable", http.StatusServiceUnavailable)
		return
	}
	blob, err := handler.Service.Blob(request.Context(), request.PathValue("repository"), request.URL.Query().Get("ref"), request.URL.Query().Get("path"))
	if err != nil {
		handler.writeError(writer, request, err)
		return
	}
	handler.write(writer, blob)
}

func (handler SourceHandler) RawBlob(writer http.ResponseWriter, request *http.Request) {
	if handler.Service == nil {
		http.Error(writer, "source service is unavailable", http.StatusServiceUnavailable)
		return
	}
	path := request.URL.Query().Get("path")
	content, oid, err := handler.Service.RawBlob(request.Context(), request.PathValue("repository"), request.URL.Query().Get("ref"), path)
	if err != nil {
		handler.writeError(writer, request, err)
		return
	}
	contentType := mime.TypeByExtension(filepath.Ext(path))
	if contentType == "" {
		contentType = http.DetectContentType(content)
	}
	writer.Header().Set("Content-Type", contentType)
	writer.Header().Set("ETag", `"`+oid+`"`)
	writer.Header().Set("Cache-Control", "public, max-age=300")
	writer.Header().Set("X-Content-Type-Options", "nosniff")
	writer.WriteHeader(http.StatusOK)
	_, _ = writer.Write(content)
}

func (handler SourceHandler) Commits(writer http.ResponseWriter, request *http.Request) {
	if handler.Service == nil {
		http.Error(writer, "source service is unavailable", http.StatusServiceUnavailable)
		return
	}
	commits, err := handler.Service.Commits(request.Context(), request.PathValue("repository"), request.URL.Query().Get("ref"), request.URL.Query().Get("path"))
	if err != nil {
		handler.writeError(writer, request, err)
		return
	}
	handler.write(writer, CommitsResponse{Commits: commits})
}

func (handler SourceHandler) CommitDetail(writer http.ResponseWriter, request *http.Request) {
	if handler.Service == nil {
		http.Error(writer, "source service is unavailable", http.StatusServiceUnavailable)
		return
	}
	detail, err := handler.Service.CommitDetail(request.Context(), request.PathValue("repository"), request.PathValue("oid"))
	if err != nil {
		handler.writeError(writer, request, err)
		return
	}
	handler.write(writer, detail)
}

func (handler SourceHandler) Refs(writer http.ResponseWriter, request *http.Request) {
	if handler.Service == nil {
		http.Error(writer, "source service is unavailable", http.StatusServiceUnavailable)
		return
	}
	refs, err := handler.Service.Refs(request.Context(), request.PathValue("repository"))
	if err != nil {
		handler.writeError(writer, request, err)
		return
	}
	handler.write(writer, refs)
}

func (handler SourceHandler) write(writer http.ResponseWriter, value any) {
	if err := xmlcodec.Write(writer, http.StatusOK, value); err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
	}
}

func (handler SourceHandler) writeError(writer http.ResponseWriter, request *http.Request, err error) {
	switch {
	case errors.Is(err, storage.ErrNotFound):
		http.NotFound(writer, request)
	case strings.Contains(err.Error(), "invalid"):
		http.Error(writer, "invalid source request", http.StatusBadRequest)
	case strings.Contains(err.Error(), "not mirrored yet"):
		http.Error(writer, "repository is not available yet", http.StatusServiceUnavailable)
	default:
		http.Error(writer, "failed to read repository", http.StatusInternalServerError)
	}
}
