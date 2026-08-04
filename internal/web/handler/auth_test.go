package handler

import (
	"encoding/base64"
	"encoding/xml"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/pquerna/otp/totp"
	"github.com/wavefnd/wave-platform/internal/identity"
	"github.com/wavefnd/wave-platform/internal/storage"
)

func TestTOTPRegisterCurrentAndLogout(t *testing.T) {
	database, err := storage.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	key := base64.StdEncoding.EncodeToString([]byte("0123456789abcdef0123456789abcdef"))
	service, err := identity.NewServiceWithTOTP(database, "wave-lang.dev", true, time.Hour, key, "Wave Test", "http://wave.test")
	if err != nil {
		t.Fatal(err)
	}
	handler := AuthHandler{Service: service, MailDomain: "wave-lang.dev", RegistrationOpen: true}
	configResponse := httptest.NewRecorder()
	handler.Config(configResponse, httptest.NewRequest(http.MethodGet, "http://wave.test/api/v1/auth/config", nil))
	if configResponse.Code != http.StatusOK || !strings.Contains(configResponse.Body.String(), "<totp-configured>true</totp-configured>") {
		t.Fatalf("config status=%d body=%s", configResponse.Code, configResponse.Body.String())
	}

	beginBody := `<?xml version="1.0"?><registration xmlns="https://wave-lang.dev/ns/platform/api/v1"><display-name>John Mark</display-name><recovery-email>john@example.com</recovery-email></registration>`
	begin := httptest.NewRequest(http.MethodPost, "http://wave.test/api/v1/auth/register/begin", strings.NewReader(beginBody))
	begin.Header.Set("Origin", "http://wave.test")
	beginResponse := httptest.NewRecorder()
	handler.BeginRegistration(beginResponse, begin)
	if beginResponse.Code != http.StatusOK {
		t.Fatalf("begin status=%d body=%s", beginResponse.Code, beginResponse.Body.String())
	}
	var enrollment EnrollmentResponse
	if err := xml.Unmarshal(beginResponse.Body.Bytes(), &enrollment); err != nil {
		t.Fatal(err)
	}
	code, err := totp.GenerateCode(enrollment.Secret, time.Now().UTC())
	if err != nil {
		t.Fatal(err)
	}
	finishBody := `<?xml version="1.0"?><enrollment xmlns="https://wave-lang.dev/ns/platform/api/v1"><token>` + enrollment.Token + `</token><code>` + code + `</code></enrollment>`
	finish := httptest.NewRequest(http.MethodPost, "http://wave.test/api/v1/auth/register/finish", strings.NewReader(finishBody))
	finish.Header.Set("Origin", "http://wave.test")
	finishResponse := httptest.NewRecorder()
	handler.FinishRegistration(finishResponse, finish)
	if finishResponse.Code != http.StatusCreated || !strings.Contains(finishResponse.Body.String(), "john-mark@wave-lang.dev") {
		t.Fatalf("finish status=%d body=%s", finishResponse.Code, finishResponse.Body.String())
	}
	cookies := finishResponse.Result().Cookies()
	if len(cookies) != 1 || cookies[0].Name != SessionCookieName || !cookies[0].HttpOnly {
		t.Fatalf("cookies=%#v", cookies)
	}

	currentRequest := httptest.NewRequest(http.MethodGet, "http://wave.test/api/v1/auth/session", nil)
	currentRequest.AddCookie(cookies[0])
	currentResponse := httptest.NewRecorder()
	handler.Current(currentResponse, currentRequest)
	if currentResponse.Code != http.StatusOK {
		t.Fatalf("current status=%d", currentResponse.Code)
	}

	logoutRequest := httptest.NewRequest(http.MethodPost, "http://wave.test/api/v1/auth/logout", nil)
	logoutRequest.Header.Set("Origin", "http://wave.test")
	logoutRequest.AddCookie(cookies[0])
	logoutResponse := httptest.NewRecorder()
	handler.Logout(logoutResponse, logoutRequest)
	if logoutResponse.Code != http.StatusNoContent {
		t.Fatalf("logout status=%d", logoutResponse.Code)
	}
}

func TestSessionTokenWithoutCookie(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "http://wave.test/", nil)
	if token, ok := sessionToken(request); ok || token != "" {
		t.Fatalf("sessionToken() = %q, %v", token, ok)
	}
}
