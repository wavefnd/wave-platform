package handler

import (
	"encoding/xml"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/wavefnd/wave-platform/internal/identity"
	mailingdomain "github.com/wavefnd/wave-platform/internal/mailinglist"
	"github.com/wavefnd/wave-platform/internal/storage"
	"github.com/wavefnd/wave-platform/internal/testsupport"
)

func TestMailingListHandlerIsPrivateAndRequiresWaveAuthentication(t *testing.T) {
	database, err := storage.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = database.Close() })
	identityService, err := testsupport.NewIdentity(database)
	if err != nil {
		t.Fatal(err)
	}
	member, _ := testsupport.Register(identityService, "List Member")
	_, token, _, _ := testsupport.Authenticate(identityService, member.Email)
	service := mailingdomain.NewService(database, "wave-lang.dev", identity.PatchMailboxAccountID)
	if err := service.EnsureDefaults(); err != nil {
		t.Fatal(err)
	}
	handler := MailingListHandler{Service: service, Auth: &AuthHandler{Service: identityService}}

	unauthorized := httptest.NewRequest(http.MethodGet, "http://wave.test/api/v1/mailing-lists", nil)
	unauthorizedResponse := httptest.NewRecorder()
	handler.Lists(unauthorizedResponse, unauthorized)
	if unauthorizedResponse.Code != http.StatusUnauthorized || unauthorizedResponse.Header().Get("X-Robots-Tag") != "noindex, nofollow, noarchive" {
		t.Fatalf("unauthorized status=%d headers=%v", unauthorizedResponse.Code, unauthorizedResponse.Header())
	}

	listRequest := mailingListRequest(http.MethodGet, "http://wave.test/api/v1/mailing-lists", "", token)
	listResponse := httptest.NewRecorder()
	handler.Lists(listResponse, listRequest)
	if listResponse.Code != http.StatusOK || !strings.Contains(listResponse.Body.String(), "development@wave-lang.dev") ||
		!strings.Contains(listResponse.Body.String(), "<webhook-policy>summary</webhook-policy>") {
		t.Fatalf("list status=%d body=%s", listResponse.Code, listResponse.Body.String())
	}

	crossOrigin := mailingListRequest(http.MethodPost, "http://wave.test/api/v1/mailing-lists/development/subscription",
		`<mailing-list-subscription xmlns="https://wave-lang.dev/ns/platform/api/v1"><subscribed>true</subscribed></mailing-list-subscription>`, token)
	crossOrigin.Header.Set("Origin", "https://attacker.test")
	crossOrigin.SetPathValue("list", "development")
	crossOriginResponse := httptest.NewRecorder()
	handler.Subscription(crossOriginResponse, crossOrigin)
	if crossOriginResponse.Code != http.StatusForbidden {
		t.Fatalf("cross-origin status=%d body=%s", crossOriginResponse.Code, crossOriginResponse.Body.String())
	}

	subscribe := mailingListRequest(http.MethodPost, "http://wave.test/api/v1/mailing-lists/development/subscription",
		`<mailing-list-subscription xmlns="https://wave-lang.dev/ns/platform/api/v1"><subscribed>true</subscribed></mailing-list-subscription>`, token)
	subscribe.SetPathValue("list", "development")
	subscribeResponse := httptest.NewRecorder()
	handler.Subscription(subscribeResponse, subscribe)
	if subscribeResponse.Code != http.StatusNoContent {
		t.Fatalf("subscribe status=%d body=%s", subscribeResponse.Code, subscribeResponse.Body.String())
	}

	post := mailingListRequest(http.MethodPost, "http://wave.test/api/v1/mailing-lists/development/threads",
		`<mailing-list-post xmlns="https://wave-lang.dev/ns/platform/api/v1"><subject>Wave package ABI</subject><body>Define a stable package ABI for internal Wave modules.</body></mailing-list-post>`, token)
	post.SetPathValue("list", "development")
	postResponse := httptest.NewRecorder()
	handler.Post(postResponse, post)
	if postResponse.Code != http.StatusCreated {
		t.Fatalf("post status=%d body=%s", postResponse.Code, postResponse.Body.String())
	}
	var thread mailingdomain.ThreadView
	if err := xml.Unmarshal(postResponse.Body.Bytes(), &thread); err != nil || thread.ID == "" || len(thread.Messages) != 1 {
		t.Fatalf("thread=%#v err=%v body=%s", thread, err, postResponse.Body.String())
	}

	detail := mailingListRequest(http.MethodGet, "http://wave.test/api/v1/mailing-lists/development/threads/"+thread.ID, "", token)
	detail.SetPathValue("list", "development")
	detail.SetPathValue("thread", thread.ID)
	detailResponse := httptest.NewRecorder()
	handler.Thread(detailResponse, detail)
	if detailResponse.Code != http.StatusOK || !strings.Contains(detailResponse.Body.String(), "Wave package ABI") {
		t.Fatalf("detail status=%d body=%s", detailResponse.Code, detailResponse.Body.String())
	}
}

func mailingListRequest(method, target, body, token string) *http.Request {
	request := httptest.NewRequest(method, target, strings.NewReader(body))
	request.Header.Set("Origin", "http://wave.test")
	if token != "" {
		request.AddCookie(&http.Cookie{Name: SessionCookieName, Value: token})
	}
	return request
}
