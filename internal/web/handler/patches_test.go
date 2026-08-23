package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/wavefnd/wave-platform/internal/identifier"
	"github.com/wavefnd/wave-platform/internal/identity"
	maildomain "github.com/wavefnd/wave-platform/internal/mail"
	"github.com/wavefnd/wave-platform/internal/mailbox"
	patchdomain "github.com/wavefnd/wave-platform/internal/patcharchive"
	"github.com/wavefnd/wave-platform/internal/permission"
	"github.com/wavefnd/wave-platform/internal/storage"
	"github.com/wavefnd/wave-platform/internal/testsupport"
)

type patchHandlerFixture struct {
	handler         PatchesHandler
	patchID         string
	memberToken     string
	maintainerToken string
}

func TestPatchArchiveListAndGetRequireActiveWaveMember(t *testing.T) {
	fixture := newPatchHandlerFixture(t)

	for _, test := range []struct {
		name string
		call func(http.ResponseWriter, *http.Request)
		path string
	}{
		{name: "list", call: fixture.handler.List, path: "http://wave.test/api/v1/patches"},
		{name: "get", call: fixture.handler.Get, path: "http://wave.test/api/v1/patches/" + fixture.patchID},
	} {
		t.Run(test.name+" anonymous", func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, test.path, nil)
			request.SetPathValue("patch", fixture.patchID)
			response := httptest.NewRecorder()
			test.call(response, request)
			if response.Code != http.StatusUnauthorized {
				t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
			}
			assertPrivatePatchResponse(t, response)
		})

		t.Run(test.name+" member", func(t *testing.T) {
			request := patchHandlerRequest(http.MethodGet, test.path, "", fixture.memberToken)
			request.SetPathValue("patch", fixture.patchID)
			response := httptest.NewRecorder()
			test.call(response, request)
			if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), fixture.patchID) {
				t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
			}
			assertPrivatePatchResponse(t, response)
		})
	}
}

func TestPatchDownloadAndReviewRequireSourceMaintainer(t *testing.T) {
	fixture := newPatchHandlerFixture(t)

	for _, test := range []struct {
		name   string
		token  string
		status int
	}{
		{name: "anonymous", status: http.StatusUnauthorized},
		{name: "member", token: fixture.memberToken, status: http.StatusForbidden},
	} {
		t.Run("download "+test.name, func(t *testing.T) {
			request := patchHandlerRequest(http.MethodGet, "http://wave.test/api/v1/patches/"+fixture.patchID+"/download", "", test.token)
			request.SetPathValue("patch", fixture.patchID)
			response := httptest.NewRecorder()
			fixture.handler.Download(response, request)
			if response.Code != test.status {
				t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
			}
			assertPrivatePatchResponse(t, response)
		})

		t.Run("review "+test.name, func(t *testing.T) {
			request := patchHandlerRequest(http.MethodPost, "http://wave.test/api/v1/patches/"+fixture.patchID+"/review",
				`<patch-review><status>reviewing</status></patch-review>`, test.token)
			request.SetPathValue("patch", fixture.patchID)
			response := httptest.NewRecorder()
			fixture.handler.Review(response, request)
			if response.Code != test.status {
				t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
			}
			assertPrivatePatchResponse(t, response)
		})
	}

	download := patchHandlerRequest(http.MethodGet, "http://wave.test/api/v1/patches/"+fixture.patchID+"/download", "", fixture.maintainerToken)
	download.SetPathValue("patch", fixture.patchID)
	downloadResponse := httptest.NewRecorder()
	fixture.handler.Download(downloadResponse, download)
	if downloadResponse.Code != http.StatusOK || downloadResponse.Header().Get("Content-Type") != "application/mbox" ||
		!strings.Contains(downloadResponse.Body.String(), "diff --git a/compiler.go b/compiler.go") {
		t.Fatalf("download status=%d headers=%v body=%s", downloadResponse.Code, downloadResponse.Header(), downloadResponse.Body.String())
	}
	assertPrivatePatchResponse(t, downloadResponse)

	review := patchHandlerRequest(http.MethodPost, "http://wave.test/api/v1/patches/"+fixture.patchID+"/review",
		`<patch-review><status>reviewing</status><target-repository>wavefnd/Wave</target-repository></patch-review>`, fixture.maintainerToken)
	review.SetPathValue("patch", fixture.patchID)
	reviewResponse := httptest.NewRecorder()
	fixture.handler.Review(reviewResponse, review)
	if reviewResponse.Code != http.StatusOK || !strings.Contains(reviewResponse.Body.String(), "<review-status>reviewing</review-status>") {
		t.Fatalf("review status=%d body=%s", reviewResponse.Code, reviewResponse.Body.String())
	}
	assertPrivatePatchResponse(t, reviewResponse)
}

func newPatchHandlerFixture(t *testing.T) patchHandlerFixture {
	t.Helper()
	database, err := storage.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = database.Close() })

	identityService, err := testsupport.NewIdentity(database)
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := identityService.BootstrapTOTPAdmin("Wave Owner", "wave-owner", "owner@example.net", testsupport.TOTPSecret); err != nil {
		t.Fatal(err)
	}
	member, err := testsupport.Register(identityService, "Patch Reader")
	if err != nil {
		t.Fatal(err)
	}
	maintainer, err := testsupport.Register(identityService, "Patch Maintainer")
	if err != nil {
		t.Fatal(err)
	}
	if err := permission.NewRepository(database).Assign(permission.Assignment{AccountID: maintainer.ID, RoleID: "source-maintainer", Scope: "source"}); err != nil {
		t.Fatal(err)
	}
	_, memberToken, _, err := testsupport.Authenticate(identityService, member.Email)
	if err != nil {
		t.Fatal(err)
	}
	_, maintainerToken, _, err := testsupport.Authenticate(identityService, maintainer.Email)
	if err != nil {
		t.Fatal(err)
	}

	box, err := identityService.EnsurePatchMailbox()
	if err != nil {
		t.Fatal(err)
	}
	patchID, err := identifier.New("message")
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 8, 23, 12, 0, 0, 0, time.UTC)
	subject := "[PATCH] compiler: preserve diagnostic ranges"
	body := "Keep diagnostic ranges stable.\n\ndiff --git a/compiler.go b/compiler.go\n--- a/compiler.go\n+++ b/compiler.go\n@@ -1 +1 @@\n-old\n+new\n"
	message := maildomain.Message{ID: patchID, MessageID: "<patch-handler-test@wave-lang.dev>", ThreadID: patchID,
		AuthorAccountID: member.ID, From: "Patch Reader <" + member.Email + ">", To: []string{box.Address}, Subject: subject,
		ReceivedAt: now, CreatedAt: now}
	raw := "From: " + message.From + "\r\nTo: " + box.Address + "\r\nSubject: " + subject + "\r\nMessage-ID: " + message.MessageID +
		"\r\nContent-Type: text/plain; charset=utf-8\r\n\r\n" + body
	if err := maildomain.NewRepository(database).UpsertMessage(message, []byte(raw)); err != nil {
		t.Fatal(err)
	}
	entryID, err := identifier.New("mailbox-entry")
	if err != nil {
		t.Fatal(err)
	}
	if err := mailbox.NewRepository(database).AddEntry(mailbox.Entry{ID: entryID, MailboxID: box.ID, MessageID: patchID, Folder: "Inbox", CreatedAt: now}); err != nil {
		t.Fatal(err)
	}

	service := patchdomain.NewService(database, identity.PatchMailboxAccountID, box.Address)
	return patchHandlerFixture{handler: PatchesHandler{Service: service, Auth: &AuthHandler{Service: identityService}}, patchID: patchID,
		memberToken: memberToken, maintainerToken: maintainerToken}
}

func patchHandlerRequest(method, target, body, token string) *http.Request {
	request := httptest.NewRequest(method, target, strings.NewReader(body))
	request.Header.Set("Origin", "http://wave.test")
	if token != "" {
		request.AddCookie(&http.Cookie{Name: SessionCookieName, Value: token})
	}
	return request
}

func assertPrivatePatchResponse(t *testing.T, response *httptest.ResponseRecorder) {
	t.Helper()
	if response.Header().Get("Cache-Control") != "no-store" || response.Header().Get("Pragma") != "no-cache" ||
		response.Header().Get("X-Robots-Tag") != "noindex, nofollow, noarchive" {
		t.Fatalf("private response headers=%v", response.Header())
	}
}
