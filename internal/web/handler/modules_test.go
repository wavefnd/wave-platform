package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/wavefnd/wave-platform/internal/storage"
	"github.com/wavefnd/wave-platform/internal/testsupport"
)

func TestModulesRequireAdministratorOrOwner(t *testing.T) {
	database, err := storage.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = database.Close() })
	service, err := testsupport.NewIdentity(database)
	if err != nil {
		t.Fatal(err)
	}
	owner, _, err := service.BootstrapTOTPAdmin("Wave Owner", "wave-owner", "owner@example.net", testsupport.TOTPSecret)
	if err != nil {
		t.Fatal(err)
	}
	_, token, _, err := testsupport.Authenticate(service, owner.Email)
	if err != nil {
		t.Fatal(err)
	}
	handler := ModulesHandler{Modules: []ModuleStatus{{Name: "document", Enabled: true, Status: "ready"}}, Auth: &AuthHandler{Service: service}}

	unauthenticated := httptest.NewRecorder()
	handler.Status(unauthenticated, httptest.NewRequest(http.MethodGet, "/api/v1/modules", nil))
	if unauthenticated.Code != http.StatusUnauthorized {
		t.Fatalf("unauthenticated status=%d", unauthenticated.Code)
	}

	authorizedRequest := httptest.NewRequest(http.MethodGet, "/api/v1/modules", nil)
	authorizedRequest.AddCookie(&http.Cookie{Name: SessionCookieName, Value: token})
	authorized := httptest.NewRecorder()
	handler.Status(authorized, authorizedRequest)
	if authorized.Code != http.StatusOK {
		t.Fatalf("owner status=%d body=%s", authorized.Code, authorized.Body.String())
	}
}
