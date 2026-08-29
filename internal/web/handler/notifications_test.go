package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	notificationdomain "github.com/wavefnd/wave-platform/internal/notification"
	"github.com/wavefnd/wave-platform/internal/storage"
	"github.com/wavefnd/wave-platform/internal/testsupport"
)

func TestNotificationsRequireAuthenticationAndRemainAccountScoped(t *testing.T) {
	database, err := storage.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = database.Close() })
	identityService, err := testsupport.NewIdentity(database)
	if err != nil {
		t.Fatal(err)
	}
	member, err := testsupport.Register(identityService, "Notification Member")
	if err != nil {
		t.Fatal(err)
	}
	other, err := testsupport.Register(identityService, "Other Member")
	if err != nil {
		t.Fatal(err)
	}
	_, token, _, err := testsupport.Authenticate(identityService, member.Email)
	if err != nil {
		t.Fatal(err)
	}
	service := notificationdomain.NewService(database)
	created, err := service.Notify(notificationdomain.Input{RecipientAccountID: member.ID, ActorAccountID: other.ID,
		ActorName: other.DisplayName, Kind: "question.answer", Subject: "How do modules work?", URL: "/questions/question-1"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.Notify(notificationdomain.Input{RecipientAccountID: other.ID, Kind: "rfc.status",
		Subject: "RFC 4", URL: "/rfcs/4"}); err != nil {
		t.Fatal(err)
	}
	handler := NotificationHandler{Service: service, Auth: &AuthHandler{Service: identityService}}

	unauthenticated := httptest.NewRequest(http.MethodGet, "http://wave.test/api/v1/notifications", nil)
	unauthenticatedResponse := httptest.NewRecorder()
	handler.List(unauthenticatedResponse, unauthenticated)
	if unauthenticatedResponse.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d body=%s", unauthenticatedResponse.Code, unauthenticatedResponse.Body.String())
	}
	assertPrivateNotificationResponse(t, unauthenticatedResponse)

	request := notificationHandlerRequest(http.MethodGet, "http://wave.test/api/v1/notifications", token)
	response := httptest.NewRecorder()
	handler.List(response, request)
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), "How do modules work?") ||
		strings.Contains(response.Body.String(), "RFC 4") || !strings.Contains(response.Body.String(), "<unread-count>1</unread-count>") {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
	assertPrivateNotificationResponse(t, response)

	crossOrigin := notificationHandlerRequest(http.MethodPost, "http://wave.test/api/v1/notifications/"+created.ID+"/read", token)
	crossOrigin.Header.Set("Origin", "https://outside.example")
	crossOrigin.SetPathValue("notification", created.ID)
	crossOriginResponse := httptest.NewRecorder()
	handler.MarkRead(crossOriginResponse, crossOrigin)
	if crossOriginResponse.Code != http.StatusForbidden {
		t.Fatalf("status=%d body=%s", crossOriginResponse.Code, crossOriginResponse.Body.String())
	}

	markRead := notificationHandlerRequest(http.MethodPost, "http://wave.test/api/v1/notifications/"+created.ID+"/read", token)
	markRead.SetPathValue("notification", created.ID)
	markReadResponse := httptest.NewRecorder()
	handler.MarkRead(markReadResponse, markRead)
	if markReadResponse.Code != http.StatusOK || !strings.Contains(markReadResponse.Body.String(), "<read-at>") {
		t.Fatalf("status=%d body=%s", markReadResponse.Code, markReadResponse.Body.String())
	}
}

func notificationHandlerRequest(method, target, token string) *http.Request {
	request := httptest.NewRequest(method, target, nil)
	request.Header.Set("Origin", "http://wave.test")
	request.AddCookie(&http.Cookie{Name: SessionCookieName, Value: token})
	return request
}

func assertPrivateNotificationResponse(t *testing.T, response *httptest.ResponseRecorder) {
	t.Helper()
	if response.Header().Get("Cache-Control") != "no-store" || response.Header().Get("Pragma") != "no-cache" ||
		response.Header().Get("X-Robots-Tag") != "noindex, nofollow, noarchive" {
		t.Fatalf("private response headers=%v", response.Header())
	}
}
