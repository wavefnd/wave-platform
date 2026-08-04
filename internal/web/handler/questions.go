package handler

import (
	"encoding/xml"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/wavefnd/wave-platform/internal/account"
	questiondomain "github.com/wavefnd/wave-platform/internal/question"
	"github.com/wavefnd/wave-platform/internal/storage"
	"github.com/wavefnd/wave-platform/internal/xmlcodec"
)

type QuestionsHandler struct {
	Repository *questiondomain.Repository
	Service    *questiondomain.Service
	Auth       *AuthHandler
}

type QuestionsResponse struct {
	XMLName xml.Name                 `xml:"https://wave-lang.dev/ns/platform/api/v1 questions"`
	Items   []questiondomain.Summary `xml:"question"`
}

type CreateQuestionRequest struct {
	XMLName     xml.Name `xml:"https://wave-lang.dev/ns/platform/api/v1 question-create"`
	Title       string   `xml:"title"`
	Body        string   `xml:"body"`
	Tags        []string `xml:"tags>tag"`
	WaveVersion string   `xml:"wave-version,omitempty"`
	Platform    string   `xml:"platform,omitempty"`
}

type CreateAnswerRequest struct {
	XMLName xml.Name `xml:"https://wave-lang.dev/ns/platform/api/v1 question-answer"`
	Body    string   `xml:"body"`
}

type QuestionVoteRequest struct {
	XMLName    xml.Name `xml:"https://wave-lang.dev/ns/platform/api/v1 question-vote"`
	TargetType string   `xml:"target-type"`
	TargetID   string   `xml:"target-id"`
	Value      int      `xml:"value"`
}

type QuestionVoteResponse struct {
	XMLName    xml.Name `xml:"https://wave-lang.dev/ns/platform/api/v1 question-vote-result"`
	Score      int64    `xml:"score"`
	ViewerVote int      `xml:"viewer-vote"`
}

type AcceptAnswerRequest struct {
	XMLName  xml.Name `xml:"https://wave-lang.dev/ns/platform/api/v1 question-accept"`
	AnswerID string   `xml:"answer-id,omitempty"`
}

func (handler QuestionsHandler) List(writer http.ResponseWriter, request *http.Request) {
	if handler.Repository == nil {
		writeAPIError(writer, http.StatusServiceUnavailable, "questions-unavailable", "The questions service is unavailable.")
		return
	}
	limit, _ := strconv.Atoi(request.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(request.URL.Query().Get("offset"))
	query := request.URL.Query().Get("q")
	if len([]rune(query)) > 200 {
		writeAPIError(writer, http.StatusBadRequest, "invalid-query", "The question search is too long.")
		return
	}
	if limit <= 0 || limit > 100 {
		limit = 30
	}
	if offset < 0 {
		offset = 0
	}
	viewerID := ""
	if handler.Auth != nil {
		if viewer, ok := AuthenticatedAccount(*handler.Auth, request); ok {
			viewerID = viewer.ID
		}
	}
	items, err := handler.Repository.Query(query, request.URL.Query().Get("sort"), request.URL.Query().Get("tag"), limit, offset, viewerID)
	if err != nil {
		writeAPIError(writer, http.StatusInternalServerError, "questions-failed", "The questions could not be loaded.")
		return
	}
	_ = xmlcodec.Write(writer, http.StatusOK, QuestionsResponse{Items: items})
}

func (handler QuestionsHandler) Get(writer http.ResponseWriter, request *http.Request) {
	if handler.Repository == nil {
		writeAPIError(writer, http.StatusServiceUnavailable, "questions-unavailable", "The questions service is unavailable.")
		return
	}
	viewerID := ""
	if handler.Auth != nil {
		if viewer, ok := AuthenticatedAccount(*handler.Auth, request); ok {
			viewerID = viewer.ID
		}
	}
	item, err := handler.Repository.View(request.PathValue("question"), viewerID)
	if errors.Is(err, storage.ErrNotFound) {
		writeAPIError(writer, http.StatusNotFound, "question-not-found", "The question was not found.")
		return
	}
	if err != nil {
		writeAPIError(writer, http.StatusInternalServerError, "questions-failed", "The question could not be loaded.")
		return
	}
	if views, viewErr := handler.Repository.RecordView(item.Question.ID); viewErr == nil {
		item.ViewCount = views
	}
	_ = xmlcodec.Write(writer, http.StatusOK, item)
}

func (handler QuestionsHandler) Create(writer http.ResponseWriter, request *http.Request) {
	actor, ok := handler.requireAccount(writer, request)
	if !ok || !handler.validMutation(writer, request) {
		return
	}
	var input CreateQuestionRequest
	if err := xmlcodec.Decode(request.Body, &input); err != nil {
		writeAPIError(writer, http.StatusBadRequest, "invalid-xml", "The question request is not valid XML.")
		return
	}
	view, err := handler.Service.Create(actor, questiondomain.CreateInput{Title: input.Title, Body: input.Body,
		Tags: input.Tags, WaveVersion: input.WaveVersion, Platform: input.Platform})
	if err != nil {
		handler.writeQuestionError(writer, err)
		return
	}
	writer.Header().Set("Location", "/api/v1/questions/"+view.Question.ID)
	_ = xmlcodec.Write(writer, http.StatusCreated, view)
}

func (handler QuestionsHandler) Answer(writer http.ResponseWriter, request *http.Request) {
	actor, ok := handler.requireAccount(writer, request)
	if !ok || !handler.validMutation(writer, request) {
		return
	}
	var input CreateAnswerRequest
	if err := xmlcodec.Decode(request.Body, &input); err != nil {
		writeAPIError(writer, http.StatusBadRequest, "invalid-xml", "The answer request is not valid XML.")
		return
	}
	view, err := handler.Service.Answer(actor, questiondomain.AnswerInput{QuestionID: request.PathValue("question"), Body: input.Body})
	if err != nil {
		handler.writeQuestionError(writer, err)
		return
	}
	_ = xmlcodec.Write(writer, http.StatusCreated, view)
}

func (handler QuestionsHandler) Vote(writer http.ResponseWriter, request *http.Request) {
	actor, ok := handler.requireAccount(writer, request)
	if !ok || !handler.validMutation(writer, request) {
		return
	}
	var input QuestionVoteRequest
	if err := xmlcodec.Decode(request.Body, &input); err != nil {
		writeAPIError(writer, http.StatusBadRequest, "invalid-xml", "The vote request is not valid XML.")
		return
	}
	score, err := handler.Service.Vote(actor.ID, request.PathValue("question"), strings.TrimSpace(input.TargetType), strings.TrimSpace(input.TargetID), input.Value)
	if err != nil {
		handler.writeQuestionError(writer, err)
		return
	}
	_ = xmlcodec.Write(writer, http.StatusOK, QuestionVoteResponse{Score: score, ViewerVote: input.Value})
}

func (handler QuestionsHandler) Accept(writer http.ResponseWriter, request *http.Request) {
	actor, ok := handler.requireAccount(writer, request)
	if !ok || !handler.validMutation(writer, request) {
		return
	}
	var input AcceptAnswerRequest
	if err := xmlcodec.Decode(request.Body, &input); err != nil {
		writeAPIError(writer, http.StatusBadRequest, "invalid-xml", "The accept request is not valid XML.")
		return
	}
	view, err := handler.Service.Accept(actor, request.PathValue("question"), input.AnswerID, handler.Auth.Service.IsAdministrator(actor.ID))
	if err != nil {
		handler.writeQuestionError(writer, err)
		return
	}
	_ = xmlcodec.Write(writer, http.StatusOK, view)
}

func (handler QuestionsHandler) requireAccount(writer http.ResponseWriter, request *http.Request) (account.Account, bool) {
	setPrivateResponseHeaders(writer)
	if handler.Service == nil || handler.Auth == nil {
		writeAPIError(writer, http.StatusServiceUnavailable, "questions-unavailable", "The questions service is unavailable.")
		return account.Account{}, false
	}
	actor, ok := AuthenticatedAccount(*handler.Auth, request)
	if !ok {
		writeAPIError(writer, http.StatusUnauthorized, "not-authenticated", "Authentication is required.")
		return account.Account{}, false
	}
	return actor, true
}

func (handler QuestionsHandler) validMutation(writer http.ResponseWriter, request *http.Request) bool {
	if sameOrigin(request) {
		return true
	}
	writeAPIError(writer, http.StatusForbidden, "invalid-origin", "The request origin is not allowed.")
	return false
}

func (handler QuestionsHandler) writeQuestionError(writer http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, storage.ErrNotFound):
		writeAPIError(writer, http.StatusNotFound, "question-not-found", "The question was not found.")
	case errors.Is(err, questiondomain.ErrForbidden):
		writeAPIError(writer, http.StatusForbidden, "question-forbidden", strings.TrimPrefix(err.Error(), questiondomain.ErrForbidden.Error()+": "))
	case errors.Is(err, questiondomain.ErrQuestionClosed):
		writeAPIError(writer, http.StatusConflict, "question-closed", "This question is closed.")
	case errors.Is(err, questiondomain.ErrInvalidQuestion):
		writeAPIError(writer, http.StatusUnprocessableEntity, "invalid-question", strings.TrimPrefix(err.Error(), questiondomain.ErrInvalidQuestion.Error()+": "))
	default:
		writeAPIError(writer, http.StatusInternalServerError, "questions-failed", "The question request could not be completed.")
	}
}
