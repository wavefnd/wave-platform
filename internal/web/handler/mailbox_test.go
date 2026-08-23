package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/wavefnd/wave-platform/internal/storage"
	"github.com/wavefnd/wave-platform/internal/testsupport"
)

func TestMailboxSendReadSearchAndArchive(t *testing.T) {
	database, err := storage.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	service, err := testsupport.NewIdentity(database)
	if err != nil {
		t.Fatal(err)
	}
	sender, err := testsupport.Register(service, "Ada Lovelace")
	if err != nil {
		t.Fatal(err)
	}
	recipient, err := testsupport.Register(service, "Grace Hopper")
	if err != nil {
		t.Fatal(err)
	}
	_, senderToken, _, err := testsupport.Authenticate(service, sender.Email)
	if err != nil {
		t.Fatal(err)
	}
	_, recipientToken, _, err := testsupport.Authenticate(service, recipient.Email)
	if err != nil {
		t.Fatal(err)
	}
	auth := AuthHandler{Service: service}
	handler := MailboxHandler{Auth: auth}

	sendBody := `<send-mail xmlns="https://wave-lang.dev/ns/platform/api/v1"><to>` + recipient.Email + `</to><subject>Compiler status</subject><body>The build is ready.</body></send-mail>`
	sendRequest := httptest.NewRequest(http.MethodPost, "http://wave.test/api/v1/mailbox/messages", strings.NewReader(sendBody))
	sendRequest.Header.Set("Origin", "http://wave.test")
	sendRequest.AddCookie(&http.Cookie{Name: SessionCookieName, Value: senderToken})
	sendResponse := httptest.NewRecorder()
	handler.Send(sendResponse, sendRequest)
	if sendResponse.Code != http.StatusCreated {
		t.Fatalf("send status=%d body=%s", sendResponse.Code, sendResponse.Body.String())
	}

	listRequest := httptest.NewRequest(http.MethodGet, "http://wave.test/api/v1/mailbox?folder=Inbox&q=build", nil)
	listRequest.AddCookie(&http.Cookie{Name: SessionCookieName, Value: recipientToken})
	listResponse := httptest.NewRecorder()
	handler.List(listResponse, listRequest)
	if listResponse.Code != http.StatusOK || !strings.Contains(listResponse.Body.String(), "The build is ready.") {
		t.Fatalf("list status=%d body=%s", listResponse.Code, listResponse.Body.String())
	}
	items, err := service.MailboxItems(recipient.ID, "Inbox")
	if err != nil || len(items) != 1 {
		t.Fatalf("items=%d err=%v", len(items), err)
	}
	entryID := items[0].Entry.ID

	messageRequest := httptest.NewRequest(http.MethodGet, "http://wave.test/api/v1/mailbox/messages/"+entryID, nil)
	messageRequest.SetPathValue("entry", entryID)
	messageRequest.AddCookie(&http.Cookie{Name: SessionCookieName, Value: recipientToken})
	messageResponse := httptest.NewRecorder()
	handler.Message(messageResponse, messageRequest)
	if messageResponse.Code != http.StatusOK || !strings.Contains(messageResponse.Body.String(), "Compiler status") {
		t.Fatalf("message status=%d body=%s", messageResponse.Code, messageResponse.Body.String())
	}

	actionBody := `<mailbox-action xmlns="https://wave-lang.dev/ns/platform/api/v1"><action>archive</action></mailbox-action>`
	actionRequest := httptest.NewRequest(http.MethodPost, "http://wave.test/api/v1/mailbox/messages/"+entryID+"/action", strings.NewReader(actionBody))
	actionRequest.SetPathValue("entry", entryID)
	actionRequest.Header.Set("Origin", "http://wave.test")
	actionRequest.AddCookie(&http.Cookie{Name: SessionCookieName, Value: recipientToken})
	actionResponse := httptest.NewRecorder()
	handler.Action(actionResponse, actionRequest)
	if actionResponse.Code != http.StatusOK {
		t.Fatalf("action status=%d body=%s", actionResponse.Code, actionResponse.Body.String())
	}
	archived, err := service.MailboxItems(recipient.ID, "Archive")
	if err != nil || len(archived) != 1 {
		t.Fatalf("archived=%d err=%v", len(archived), err)
	}
}

func TestManagementMailboxRequiresAdministrator(t *testing.T) {
	database, err := storage.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	service, err := testsupport.NewIdentity(database)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.EnsureManagementMailbox(); err != nil {
		t.Fatal(err)
	}
	owner, _, err := service.BootstrapTOTPAdmin("Wave Owner", "wave-owner", "owner@example.net", testsupport.TOTPSecret)
	if err != nil {
		t.Fatal(err)
	}
	member, err := testsupport.Register(service, "Mailbox Member")
	if err != nil {
		t.Fatal(err)
	}
	_, ownerToken, _, err := testsupport.Authenticate(service, owner.Email)
	if err != nil {
		t.Fatal(err)
	}
	_, memberToken, _, err := testsupport.Authenticate(service, member.Email)
	if err != nil {
		t.Fatal(err)
	}
	handler := MailboxHandler{Auth: AuthHandler{Service: service}}
	for _, test := range []struct {
		name   string
		token  string
		status int
	}{
		{name: "anonymous", status: http.StatusUnauthorized},
		{name: "member", token: memberToken, status: http.StatusForbidden},
		{name: "owner", token: ownerToken, status: http.StatusOK},
	} {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, "http://wave.test/api/v1/admin/mailbox", nil)
			if test.token != "" {
				request.AddCookie(&http.Cookie{Name: SessionCookieName, Value: test.token})
			}
			response := httptest.NewRecorder()
			handler.ManagementList(response, request)
			if response.Code != test.status {
				t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
			}
		})
	}

	sendBody := `<send-management-mail xmlns="https://wave-lang.dev/ns/platform/api/v1"><from>help@wave-lang.dev</from><to>` + member.Email + `</to><subject>Support reply</subject><body>Your request has been resolved.</body></send-management-mail>`
	memberSendRequest := httptest.NewRequest(http.MethodPost, "http://wave.test/api/v1/admin/mailbox/messages", strings.NewReader(sendBody))
	memberSendRequest.Header.Set("Origin", "http://wave.test")
	memberSendRequest.AddCookie(&http.Cookie{Name: SessionCookieName, Value: memberToken})
	memberSendResponse := httptest.NewRecorder()
	handler.ManagementSend(memberSendResponse, memberSendRequest)
	if memberSendResponse.Code != http.StatusForbidden {
		t.Fatalf("member send status=%d body=%s", memberSendResponse.Code, memberSendResponse.Body.String())
	}

	ownerSendRequest := httptest.NewRequest(http.MethodPost, "http://wave.test/api/v1/admin/mailbox/messages", strings.NewReader(sendBody))
	ownerSendRequest.Header.Set("Origin", "http://wave.test")
	ownerSendRequest.AddCookie(&http.Cookie{Name: SessionCookieName, Value: ownerToken})
	ownerSendResponse := httptest.NewRecorder()
	handler.ManagementSend(ownerSendResponse, ownerSendRequest)
	if ownerSendResponse.Code != http.StatusCreated || !strings.Contains(ownerSendResponse.Body.String(), "Wave Support") {
		t.Fatalf("owner send status=%d body=%s", ownerSendResponse.Code, ownerSendResponse.Body.String())
	}
	managementSent, err := service.ManagementMailboxItems("Sent")
	if err != nil || len(managementSent) != 1 {
		t.Fatalf("management sent=%d err=%v", len(managementSent), err)
	}
	memberInbox, err := service.MailboxItems(member.ID, "Inbox")
	if err != nil || len(memberInbox) != 1 || memberInbox[0].Message.Subject != "Support reply" {
		t.Fatalf("member inbox=%#v err=%v", memberInbox, err)
	}
}
