package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	blogdomain "github.com/wavefnd/wave-platform/internal/blog"
	"github.com/wavefnd/wave-platform/internal/storage"
	"github.com/wavefnd/wave-platform/internal/testsupport"
)

func TestBlogAdministrationRequiresAdministratorAndPublishes(t *testing.T) {
	database, err := storage.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = database.Close() })
	identityService, err := testsupport.NewIdentity(database)
	if err != nil {
		t.Fatal(err)
	}
	owner, _, err := identityService.BootstrapTOTPAdmin("Wave Owner", "wave-owner", "owner@example.net", testsupport.TOTPSecret)
	if err != nil {
		t.Fatal(err)
	}
	member, err := testsupport.Register(identityService, "Member")
	if err != nil {
		t.Fatal(err)
	}
	_, ownerToken, _, err := testsupport.Authenticate(identityService, owner.Email)
	if err != nil {
		t.Fatal(err)
	}
	_, memberToken, _, err := testsupport.Authenticate(identityService, member.Email)
	if err != nil {
		t.Fatal(err)
	}
	handler := BlogHandler{Service: blogdomain.NewService(database), Auth: &AuthHandler{Service: identityService}}
	body := `<blog-post xmlns="https://wave-lang.dev/ns/platform/api/v1"><slug>wave-release</slug><category>release</category><title>Wave release</title><summary>Release summary</summary><content>## Ready</content><status>published</status></blog-post>`
	for _, test := range []struct {
		name, token string
		status      int
	}{
		{name: "anonymous", status: http.StatusUnauthorized},
		{name: "member", token: memberToken, status: http.StatusForbidden},
		{name: "owner", token: ownerToken, status: http.StatusOK},
	} {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "http://wave.test/api/v1/blog/editor/posts", strings.NewReader(body))
			request.Header.Set("Origin", "http://wave.test")
			if test.token != "" {
				request.AddCookie(&http.Cookie{Name: SessionCookieName, Value: test.token})
			}
			response := httptest.NewRecorder()
			handler.Save(response, request)
			if response.Code != test.status {
				t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
			}
		})
	}
	request := httptest.NewRequest(http.MethodGet, "http://wave.test/api/v1/blog/posts/wave-release", nil)
	request.SetPathValue("slug", "wave-release")
	response := httptest.NewRecorder()
	handler.Get(response, request)
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), "Wave release") {
		t.Fatalf("public status=%d body=%s", response.Code, response.Body.String())
	}
}

func TestBlogArticleCommentsRequireAuthenticationForWrites(t *testing.T) {
	database, err := storage.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = database.Close() })
	identityService, err := testsupport.NewIdentity(database)
	if err != nil {
		t.Fatal(err)
	}
	owner, _, err := identityService.BootstrapTOTPAdmin("Wave Owner", "wave-owner", "owner@example.net", testsupport.TOTPSecret)
	if err != nil {
		t.Fatal(err)
	}
	member, err := testsupport.Register(identityService, "Member")
	if err != nil {
		t.Fatal(err)
	}
	_, memberToken, _, err := testsupport.Authenticate(identityService, member.Email)
	if err != nil {
		t.Fatal(err)
	}
	service := blogdomain.NewService(database)
	if _, err := service.Save(owner.ID, blogdomain.Input{Slug: "engineering-note", Category: "article", Title: "Engineering note",
		Summary: "A compiler note", Content: "## Parser", Status: "published", CommentPolicy: "open"}); err != nil {
		t.Fatal(err)
	}
	handler := BlogHandler{Service: service, Auth: &AuthHandler{Service: identityService}}
	body := `<blog-comment xmlns="https://wave-lang.dev/ns/platform/api/v1"><body>Clear and useful update.</body></blog-comment>`
	for _, test := range []struct {
		name, token string
		status      int
	}{
		{name: "anonymous", status: http.StatusUnauthorized},
		{name: "member", token: memberToken, status: http.StatusCreated},
	} {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "http://wave.test/api/v1/blog/posts/engineering-note/comments", strings.NewReader(body))
			request.SetPathValue("slug", "engineering-note")
			request.Header.Set("Origin", "http://wave.test")
			if test.token != "" {
				request.AddCookie(&http.Cookie{Name: SessionCookieName, Value: test.token})
			}
			response := httptest.NewRecorder()
			handler.AddComment(response, request)
			if response.Code != test.status {
				t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
			}
		})
	}
	request := httptest.NewRequest(http.MethodGet, "http://wave.test/api/v1/blog/posts/engineering-note/comments", nil)
	request.SetPathValue("slug", "engineering-note")
	response := httptest.NewRecorder()
	handler.Comments(response, request)
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), "Clear and useful update") {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
}
