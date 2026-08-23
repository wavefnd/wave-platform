package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	editor "github.com/wavefnd/wave-platform/internal/editor"
	"github.com/wavefnd/wave-platform/internal/storage"
	"github.com/wavefnd/wave-platform/internal/testsupport"
)

func TestEditorTransformRequiresAuthenticationAndUsesUnicodeOffsets(t *testing.T) {
	database, err := storage.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = database.Close() })
	identityService, err := testsupport.NewIdentity(database)
	if err != nil {
		t.Fatal(err)
	}
	member, err := testsupport.Register(identityService, "Editor Member")
	if err != nil {
		t.Fatal(err)
	}
	_, token, _, err := testsupport.Authenticate(identityService, member.Email)
	if err != nil {
		t.Fatal(err)
	}
	handler := EditorHandler{Engine: editor.ReferenceEngine{}, Auth: &AuthHandler{Service: identityService}}
	body := `<editor-transform xmlns="https://wave-lang.dev/ns/platform/api/v1"><content>Wave 언어</content><selection-start>5</selection-start><selection-end>7</selection-end><command>bold</command></editor-transform>`

	unauthenticated := httptest.NewRequest(http.MethodPost, "http://wave.test/api/v1/editor/transform", strings.NewReader(body))
	unauthenticated.Header.Set("Origin", "http://wave.test")
	unauthenticatedResponse := httptest.NewRecorder()
	handler.Transform(unauthenticatedResponse, unauthenticated)
	if unauthenticatedResponse.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d", unauthenticatedResponse.Code)
	}

	request := httptest.NewRequest(http.MethodPost, "http://wave.test/api/v1/editor/transform", strings.NewReader(body))
	request.Header.Set("Origin", "http://wave.test")
	request.AddCookie(&http.Cookie{Name: SessionCookieName, Value: token})
	response := httptest.NewRecorder()
	handler.Transform(response, request)
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), "Wave **언어**") || !strings.Contains(response.Body.String(), "<engine>go</engine>") {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
	if response.Header().Get("X-Robots-Tag") != "noindex, nofollow, noarchive" {
		t.Fatalf("headers=%v", response.Header())
	}
}
