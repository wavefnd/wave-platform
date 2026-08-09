package web

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/wavefnd/wave-platform/internal/web/handler"
)

func TestRouterServesAPIAndSPA(t *testing.T) {
	frontend := t.TempDir()
	index := `<head><!-- wave:seo:start --><title>Default</title><!-- wave:seo:end --></head><main>Wave</main>`
	if err := os.WriteFile(filepath.Join(frontend, "index.html"), []byte(index), 0o600); err != nil {
		t.Fatalf("write index: %v", err)
	}

	router := NewRouter(
		"test",
		frontend,
		"https://wave.example",
		"0.1.0",
		[]handler.ModuleStatus{{Name: "document", Enabled: true, Status: "foundation"}},
		func() error { return nil },
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
	)

	for _, test := range []struct {
		path       string
		status     int
		body       string
		contentXML bool
	}{
		{path: "/api/v1/platform", status: http.StatusOK, body: "<environment>test</environment>", contentXML: true},
		{path: "/api/v1/modules", status: http.StatusUnauthorized, body: `<code>not-authenticated</code>`, contentXML: true},
		{path: "/health/ready", status: http.StatusOK, body: "<ready>true</ready>", contentXML: true},
		{path: "/robots.txt", status: http.StatusOK, body: "Sitemap: https://wave.example/sitemap.xml"},
		{path: "/sitemap.xml", status: http.StatusOK, body: "<loc>https://wave.example/docs</loc>", contentXML: true},
		{path: "/docs/compiler", status: http.StatusOK, body: "<main>Wave</main>"},
	} {
		t.Run(test.path, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, test.path, nil)
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)

			if response.Code != test.status {
				t.Fatalf("status = %d, want %d", response.Code, test.status)
			}
			if !strings.Contains(response.Body.String(), test.body) {
				t.Fatalf("body = %q, want content %q", response.Body.String(), test.body)
			}
			if test.contentXML && !strings.HasPrefix(response.Header().Get("Content-Type"), "application/xml") {
				t.Fatalf("content type = %q", response.Header().Get("Content-Type"))
			}
		})
	}

	metadataRequest := httptest.NewRequest(http.MethodGet, "/docs/compiler", nil)
	metadataResponse := httptest.NewRecorder()
	router.ServeHTTP(metadataResponse, metadataRequest)
	for _, expected := range []string{`<title>Documentation · Wave</title>`, `rel="canonical" href="https://wave.example/docs/compiler"`} {
		if !strings.Contains(metadataResponse.Body.String(), expected) {
			t.Fatalf("rendered HTML does not contain %q", expected)
		}
	}

	request := httptest.NewRequest(http.MethodGet, "/account/security", nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if value := response.Header().Get("X-Robots-Tag"); value != "noindex, nofollow, noarchive" {
		t.Fatalf("X-Robots-Tag = %q", value)
	}
}

func TestRouterReportsUnavailableDatabase(t *testing.T) {
	router := NewRouter("test", t.TempDir(), "https://wave.example", "0.1.0", nil, func() error {
		return errors.New("database unavailable")
	}, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	request := httptest.NewRequest(http.MethodGet, "/health/ready", nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d", response.Code)
	}
	if !strings.Contains(response.Body.String(), "<ready>false</ready>") {
		t.Fatalf("body = %q", response.Body.String())
	}
}
