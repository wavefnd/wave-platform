package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/wavefnd/wave-platform/internal/permission"
	rfcdomain "github.com/wavefnd/wave-platform/internal/rfc"
	"github.com/wavefnd/wave-platform/internal/storage"
	"github.com/wavefnd/wave-platform/internal/testsupport"
)

func TestRFCHandlerRequiresAuthenticationAndSeparatesAdministratorDecisions(t *testing.T) {
	database, err := storage.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = database.Close() })
	identityService, err := testsupport.NewIdentity(database)
	if err != nil {
		t.Fatal(err)
	}
	author, _ := testsupport.Register(identityService, "Proposal Author")
	admin, _ := testsupport.Register(identityService, "Platform Admin")
	owner, _ := testsupport.Register(identityService, "Project Owner")
	_, authorToken, _, _ := testsupport.Authenticate(identityService, author.Email)
	_, adminToken, _, _ := testsupport.Authenticate(identityService, admin.Email)
	_, ownerToken, _, _ := testsupport.Authenticate(identityService, owner.Email)
	permissions := permission.NewRepository(database)
	_ = permissions.Assign(permission.Assignment{AccountID: admin.ID, RoleID: "platform-admin", Scope: "platform"})
	_ = permissions.Assign(permission.Assignment{AccountID: owner.ID, RoleID: "platform-owner", Scope: "platform"})
	handler := RFCHandler{Service: rfcdomain.NewService(database), Auth: &AuthHandler{Service: identityService}}

	body := `<rfc xmlns="https://wave-lang.dev/ns/platform/api/v1"><title>Stable WebAssembly target</title><content>## Motivation&#10;&#10;Define an official WebAssembly target and stable host ABI.</content></rfc>`
	unauthorized := rfcMutation(http.MethodPost, "http://wave.test/api/v1/rfcs", body, "")
	unauthorizedResponse := httptest.NewRecorder()
	handler.Create(unauthorizedResponse, unauthorized)
	if unauthorizedResponse.Code != http.StatusUnauthorized {
		t.Fatalf("unauthorized status=%d body=%s", unauthorizedResponse.Code, unauthorizedResponse.Body.String())
	}

	create := rfcMutation(http.MethodPost, "http://wave.test/api/v1/rfcs", body, authorToken)
	createResponse := httptest.NewRecorder()
	handler.Create(createResponse, create)
	if createResponse.Code != http.StatusCreated || !strings.Contains(createResponse.Body.String(), `number="1"`) {
		t.Fatalf("create status=%d body=%s", createResponse.Code, createResponse.Body.String())
	}

	adminStatus := rfcMutation(http.MethodPost, "http://wave.test/api/v1/rfcs/1/status", `<rfc-status xmlns="https://wave-lang.dev/ns/platform/api/v1"><status>accepted</status></rfc-status>`, adminToken)
	adminStatus.SetPathValue("number", "1")
	adminResponse := httptest.NewRecorder()
	handler.UpdateStatus(adminResponse, adminStatus)
	if adminResponse.Code != http.StatusForbidden {
		t.Fatalf("administrator status=%d body=%s", adminResponse.Code, adminResponse.Body.String())
	}

	ownerStatus := rfcMutation(http.MethodPost, "http://wave.test/api/v1/rfcs/1/status", `<rfc-status xmlns="https://wave-lang.dev/ns/platform/api/v1"><status>discussion</status></rfc-status>`, ownerToken)
	ownerStatus.SetPathValue("number", "1")
	ownerResponse := httptest.NewRecorder()
	handler.UpdateStatus(ownerResponse, ownerStatus)
	if ownerResponse.Code != http.StatusOK || !strings.Contains(ownerResponse.Body.String(), "discussion") {
		t.Fatalf("owner status=%d body=%s", ownerResponse.Code, ownerResponse.Body.String())
	}

	comment := rfcMutation(http.MethodPost, "http://wave.test/api/v1/rfcs/1/comments", `<rfc-comment xmlns="https://wave-lang.dev/ns/platform/api/v1"><body>Please document WASI compatibility.</body></rfc-comment>`, authorToken)
	comment.SetPathValue("number", "1")
	commentResponse := httptest.NewRecorder()
	handler.Comment(commentResponse, comment)
	if commentResponse.Code != http.StatusCreated {
		t.Fatalf("comment status=%d body=%s", commentResponse.Code, commentResponse.Body.String())
	}
}

func rfcMutation(method, target, body, token string) *http.Request {
	request := httptest.NewRequest(method, target, strings.NewReader(body))
	request.Header.Set("Origin", "http://wave.test")
	if token != "" {
		request.AddCookie(&http.Cookie{Name: SessionCookieName, Value: token})
	}
	return request
}
