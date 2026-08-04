package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	questiondomain "github.com/wavefnd/wave-platform/internal/question"
	"github.com/wavefnd/wave-platform/internal/storage"
	"github.com/wavefnd/wave-platform/internal/testsupport"
)

func TestQuestionsCreateAnswerAcceptAndList(t *testing.T) {
	database, err := storage.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()

	identityService, err := testsupport.NewIdentity(database)
	if err != nil {
		t.Fatal(err)
	}
	asker, err := testsupport.Register(identityService, "Ada Lovelace")
	if err != nil {
		t.Fatal(err)
	}
	answerer, err := testsupport.Register(identityService, "Grace Hopper")
	if err != nil {
		t.Fatal(err)
	}
	_, askerToken, _, err := testsupport.Authenticate(identityService, asker.Email)
	if err != nil {
		t.Fatal(err)
	}
	_, answererToken, _, err := testsupport.Authenticate(identityService, answerer.Email)
	if err != nil {
		t.Fatal(err)
	}

	auth := AuthHandler{Service: identityService}
	repository := questiondomain.NewRepository(database)
	handler := QuestionsHandler{Repository: repository, Service: questiondomain.NewService(database, "wave-lang.dev"), Auth: &auth}

	createBody := `<question-create xmlns="https://wave-lang.dev/ns/platform/api/v1"><title>Why does generic inference fail here?</title><body>The compiler reports a type mismatch for this generic function call.</body><tags><tag>compiler</tag><tag>generics</tag></tags><wave-version>0.2.0-pre-beta</wave-version><platform>Linux x86_64</platform></question-create>`
	createRequest := questionMutationRequest(http.MethodPost, "http://wave.test/api/v1/questions", createBody, askerToken)
	createResponse := httptest.NewRecorder()
	handler.Create(createResponse, createRequest)
	if createResponse.Code != http.StatusCreated {
		t.Fatalf("create status=%d body=%s", createResponse.Code, createResponse.Body.String())
	}
	items, err := repository.Query("generic", "newest", "compiler", 30, 0, asker.ID)
	if err != nil || len(items) != 1 {
		t.Fatalf("questions=%#v err=%v", items, err)
	}
	questionID := items[0].ID

	answerBody := `<question-answer xmlns="https://wave-lang.dev/ns/platform/api/v1"><body>Add an explicit type argument until inference is fixed.</body></question-answer>`
	answerRequest := questionMutationRequest(http.MethodPost, "http://wave.test/api/v1/questions/"+questionID+"/answers", answerBody, answererToken)
	answerRequest.SetPathValue("question", questionID)
	answerResponse := httptest.NewRecorder()
	handler.Answer(answerResponse, answerRequest)
	if answerResponse.Code != http.StatusCreated {
		t.Fatalf("answer status=%d body=%s", answerResponse.Code, answerResponse.Body.String())
	}
	view, err := repository.View(questionID, asker.ID)
	if err != nil || len(view.Answers) != 1 {
		t.Fatalf("view=%#v err=%v", view, err)
	}

	acceptBody := `<question-accept xmlns="https://wave-lang.dev/ns/platform/api/v1"><answer-id>` + view.Answers[0].ID + `</answer-id></question-accept>`
	acceptRequest := questionMutationRequest(http.MethodPost, "http://wave.test/api/v1/questions/"+questionID+"/accept", acceptBody, askerToken)
	acceptRequest.SetPathValue("question", questionID)
	acceptResponse := httptest.NewRecorder()
	handler.Accept(acceptResponse, acceptRequest)
	if acceptResponse.Code != http.StatusOK || !strings.Contains(acceptResponse.Body.String(), "resolved") {
		t.Fatalf("accept status=%d body=%s", acceptResponse.Code, acceptResponse.Body.String())
	}

	listRequest := httptest.NewRequest(http.MethodGet, "http://wave.test/api/v1/questions?q=generic&sort=active", nil)
	listResponse := httptest.NewRecorder()
	handler.List(listResponse, listRequest)
	if listResponse.Code != http.StatusOK || !strings.Contains(listResponse.Body.String(), "Why does generic inference fail here?") {
		t.Fatalf("list status=%d body=%s", listResponse.Code, listResponse.Body.String())
	}

	unauthorized := questionMutationRequest(http.MethodPost, "http://wave.test/api/v1/questions", createBody, "")
	unauthorizedResponse := httptest.NewRecorder()
	handler.Create(unauthorizedResponse, unauthorized)
	if unauthorizedResponse.Code != http.StatusUnauthorized {
		t.Fatalf("unauthorized status=%d body=%s", unauthorizedResponse.Code, unauthorizedResponse.Body.String())
	}
}

func questionMutationRequest(method, target, body, token string) *http.Request {
	request := httptest.NewRequest(method, target, strings.NewReader(body))
	request.Header.Set("Origin", "http://wave.test")
	if token != "" {
		request.AddCookie(&http.Cookie{Name: SessionCookieName, Value: token})
	}
	return request
}
