package handler

import (
	"encoding/xml"
	"errors"
	"net/http"
	"strconv"
	"strings"

	blogdomain "github.com/wavefnd/wave-platform/internal/blog"
	"github.com/wavefnd/wave-platform/internal/storage"
	"github.com/wavefnd/wave-platform/internal/xmlcodec"
)

type BlogPostsResponse struct {
	XMLName xml.Name             `xml:"https://wave-lang.dev/ns/platform/api/v1 blog-posts"`
	Items   []blogdomain.Summary `xml:"post"`
}

type BlogHandler struct {
	Service *blogdomain.Service
	Auth    *AuthHandler
}

func (handler BlogHandler) List(writer http.ResponseWriter, request *http.Request) {
	if handler.Service == nil {
		writeAPIError(writer, http.StatusServiceUnavailable, "blog-unavailable", "The blog is unavailable.")
		return
	}
	limit, _ := strconv.Atoi(request.URL.Query().Get("limit"))
	category := normalizedBlogCategory(request.URL.Query().Get("category"))
	if request.URL.Query().Get("category") != "" && category == "" {
		writeAPIError(writer, http.StatusBadRequest, "invalid-category", "Blog category must be article, release, or roadmap.")
		return
	}
	items, err := handler.Service.Repository().PostsByCategory(category, false, limit)
	if err != nil {
		writeAPIError(writer, http.StatusInternalServerError, "blog-unavailable", "Blog posts could not be loaded.")
		return
	}
	summaries := make([]blogdomain.Summary, 0, len(items))
	for _, item := range items {
		summaries = append(summaries, blogdomain.SummaryOf(item, false))
	}
	_ = xmlcodec.Write(writer, http.StatusOK, BlogPostsResponse{Items: summaries})
}

func (handler BlogHandler) Get(writer http.ResponseWriter, request *http.Request) {
	if handler.Service == nil {
		writeAPIError(writer, http.StatusServiceUnavailable, "blog-unavailable", "The blog is unavailable.")
		return
	}
	item, err := handler.Service.Repository().Post(request.PathValue("slug"), false)
	if errors.Is(err, storage.ErrNotFound) {
		writeAPIError(writer, http.StatusNotFound, "blog-post-not-found", "The blog post was not found.")
		return
	}
	if err != nil {
		writeAPIError(writer, http.StatusInternalServerError, "blog-unavailable", "The blog post could not be loaded.")
		return
	}
	_ = xmlcodec.Write(writer, http.StatusOK, item)
}

func (handler BlogHandler) EditorList(writer http.ResponseWriter, request *http.Request) {
	if _, ok := handler.authorize(writer, request, false); !ok {
		return
	}
	items, err := handler.Service.Repository().Posts(true, 0)
	if err != nil {
		writeAPIError(writer, http.StatusInternalServerError, "blog-unavailable", "Blog posts could not be loaded.")
		return
	}
	summaries := make([]blogdomain.Summary, 0, len(items))
	for _, item := range items {
		summaries = append(summaries, blogdomain.SummaryOf(item, true))
	}
	_ = xmlcodec.Write(writer, http.StatusOK, BlogPostsResponse{Items: summaries})
}

func (handler BlogHandler) EditorGet(writer http.ResponseWriter, request *http.Request) {
	if _, ok := handler.authorize(writer, request, false); !ok {
		return
	}
	item, err := handler.Service.Repository().Post(request.PathValue("slug"), true)
	if errors.Is(err, storage.ErrNotFound) {
		writeAPIError(writer, http.StatusNotFound, "blog-post-not-found", "The blog post was not found.")
		return
	}
	if err != nil {
		writeAPIError(writer, http.StatusInternalServerError, "blog-unavailable", "The blog post could not be loaded.")
		return
	}
	_ = xmlcodec.Write(writer, http.StatusOK, item)
}

func (handler BlogHandler) Save(writer http.ResponseWriter, request *http.Request) {
	actorID, ok := handler.authorize(writer, request, true)
	if !ok {
		return
	}
	var input blogdomain.Input
	if xmlcodec.Decode(request.Body, &input) != nil {
		writeAPIError(writer, http.StatusBadRequest, "invalid-xml", "The blog post request is not valid XML.")
		return
	}
	item, err := handler.Service.Save(actorID, input)
	if errors.Is(err, blogdomain.ErrInvalidPost) {
		writeAPIError(writer, http.StatusUnprocessableEntity, "invalid-blog-post", strings.TrimPrefix(err.Error(), blogdomain.ErrInvalidPost.Error()+": "))
		return
	}
	if err != nil {
		writeAPIError(writer, http.StatusInternalServerError, "blog-save-failed", "The blog post could not be saved.")
		return
	}
	_ = xmlcodec.Write(writer, http.StatusOK, item)
}

func (handler BlogHandler) authorize(writer http.ResponseWriter, request *http.Request, mutation bool) (string, bool) {
	setPrivateResponseHeaders(writer)
	if handler.Service == nil || handler.Auth == nil {
		writeAPIError(writer, http.StatusServiceUnavailable, "blog-unavailable", "Blog administration is unavailable.")
		return "", false
	}
	actor, authenticated := AuthenticatedAccount(*handler.Auth, request)
	if !authenticated {
		writeAPIError(writer, http.StatusUnauthorized, "not-authenticated", "Authentication is required.")
		return "", false
	}
	if !handler.Auth.Service.IsAdministrator(actor.ID) {
		writeAPIError(writer, http.StatusForbidden, "administrator-required", "Administrator access is required.")
		return "", false
	}
	if mutation && !sameOrigin(request) {
		writeAPIError(writer, http.StatusForbidden, "invalid-origin", "The request origin is not allowed.")
		return "", false
	}
	return actor.ID, true
}

func normalizedBlogCategory(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "article" || value == "release" || value == "roadmap" {
		return value
	}
	return ""
}
