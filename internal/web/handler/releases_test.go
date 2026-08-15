package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	blogdomain "github.com/wavefnd/wave-platform/internal/blog"
	"github.com/wavefnd/wave-platform/internal/storage"
)

func TestLegacyReleaseAPIReadsReleaseBlogPosts(t *testing.T) {
	database, err := storage.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = database.Close() })
	repository := blogdomain.NewRepository(database)
	now := time.Date(2026, 8, 15, 0, 0, 0, 0, time.UTC)
	for _, item := range []blogdomain.Post{
		{Slug: "wave-release", Locale: "en", Category: "release", Title: "Wave release", Summary: "Release summary", Content: "## Changes", Status: "published", PublishedAt: now.Format(time.RFC3339), CreatedAt: now, UpdatedAt: now},
		{Slug: "engineering-note", Locale: "en", Category: "article", Title: "Engineering note", Summary: "Article summary", Content: "Details", Status: "published", PublishedAt: now.Add(time.Hour).Format(time.RFC3339), CreatedAt: now, UpdatedAt: now},
	} {
		if err := repository.Upsert(item); err != nil {
			t.Fatal(err)
		}
	}
	handler := ReleasesHandler{Service: blogdomain.NewService(database)}

	listRequest := httptest.NewRequest(http.MethodGet, "/api/v1/releases", nil)
	listResponse := httptest.NewRecorder()
	handler.List(listResponse, listRequest)
	if listResponse.Code != http.StatusOK || !strings.Contains(listResponse.Body.String(), "Wave release") || strings.Contains(listResponse.Body.String(), "Engineering note") {
		t.Fatalf("release list status=%d body=%q", listResponse.Code, listResponse.Body.String())
	}

	getRequest := httptest.NewRequest(http.MethodGet, "/api/v1/releases/wave-release", nil)
	getRequest.SetPathValue("slug", "wave-release")
	getResponse := httptest.NewRecorder()
	handler.Get(getResponse, getRequest)
	if getResponse.Code != http.StatusOK || !strings.Contains(getResponse.Body.String(), "## Changes") {
		t.Fatalf("release detail status=%d body=%q", getResponse.Code, getResponse.Body.String())
	}

	articleRequest := httptest.NewRequest(http.MethodGet, "/api/v1/releases/engineering-note", nil)
	articleRequest.SetPathValue("slug", "engineering-note")
	articleResponse := httptest.NewRecorder()
	handler.Get(articleResponse, articleRequest)
	if articleResponse.Code != http.StatusNotFound {
		t.Fatalf("article release status=%d", articleResponse.Code)
	}
}
