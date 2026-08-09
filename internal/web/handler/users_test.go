package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/wavefnd/wave-platform/internal/storage"
	"github.com/wavefnd/wave-platform/internal/testsupport"
)

func TestUserDirectoryIsPublicAndProfileUpdateRequiresAuthentication(t *testing.T) {
	database, err := storage.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = database.Close() })
	identities, err := testsupport.NewIdentity(database)
	if err != nil {
		t.Fatal(err)
	}
	member, err := testsupport.Register(identities, "Profile Member")
	if err != nil {
		t.Fatal(err)
	}
	_, token, _, err := testsupport.Authenticate(identities, member.Email)
	if err != nil {
		t.Fatal(err)
	}
	auth := AuthHandler{Service: identities, MailDomain: "wave-lang.dev"}
	handler := UsersHandler{Auth: &auth}

	directory := httptest.NewRecorder()
	handler.Directory(directory, httptest.NewRequest(http.MethodGet, "http://wave.test/api/v1/users", nil))
	if directory.Code != http.StatusOK || !strings.Contains(directory.Body.String(), member.Email) {
		t.Fatalf("directory status=%d body=%s", directory.Code, directory.Body.String())
	}

	profileRequest := httptest.NewRequest(http.MethodGet, "http://wave.test/api/v1/users/"+member.Username, nil)
	profileRequest.SetPathValue("user", member.Username)
	profile := httptest.NewRecorder()
	handler.Profile(profile, profileRequest)
	if profile.Code != http.StatusOK || !strings.Contains(profile.Body.String(), member.DisplayName) {
		t.Fatalf("profile status=%d body=%s", profile.Code, profile.Body.String())
	}

	body := `<user-profile-update xmlns="https://wave-lang.dev/ns/platform/api/v1"><display-name>Shared Nickname</display-name><bio>Wave contributor.</bio><time-zone>Asia/Seoul</time-zone></user-profile-update>`
	unauthorized := httptest.NewRecorder()
	handler.UpdateProfile(unauthorized, httptest.NewRequest(http.MethodPost, "http://wave.test/api/v1/users/me/profile", strings.NewReader(body)))
	if unauthorized.Code != http.StatusUnauthorized {
		t.Fatalf("unauthorized status=%d body=%s", unauthorized.Code, unauthorized.Body.String())
	}

	updateRequest := httptest.NewRequest(http.MethodPost, "http://wave.test/api/v1/users/me/profile", strings.NewReader(body))
	updateRequest.Header.Set("Origin", "http://wave.test")
	updateRequest.AddCookie(&http.Cookie{Name: SessionCookieName, Value: token})
	updated := httptest.NewRecorder()
	handler.UpdateProfile(updated, updateRequest)
	if updated.Code != http.StatusOK || !strings.Contains(updated.Body.String(), "Asia/Seoul") || !strings.Contains(updated.Body.String(), "Shared Nickname") {
		t.Fatalf("update status=%d body=%s", updated.Code, updated.Body.String())
	}
}
