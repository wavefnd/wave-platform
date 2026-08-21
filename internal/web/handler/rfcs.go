package handler

import (
	"encoding/xml"
	"errors"
	"net/http"
	"strconv"
	"strings"

	rfcdomain "github.com/wavefnd/wave-platform/internal/rfc"
	"github.com/wavefnd/wave-platform/internal/storage"
	"github.com/wavefnd/wave-platform/internal/xmlcodec"
)

type RFCsResponse struct {
	XMLName xml.Name             `xml:"https://wave-lang.dev/ns/platform/api/v1 rfcs"`
	Items   []rfcdomain.Proposal `xml:"rfc"`
}

type RFCHandler struct {
	Service *rfcdomain.Service
	Auth    *AuthHandler
}

func (handler RFCHandler) List(writer http.ResponseWriter, request *http.Request) {
	if handler.Service == nil {
		writeAPIError(writer, http.StatusServiceUnavailable, "rfcs-unavailable", "RFCs are unavailable.")
		return
	}
	query := strings.TrimSpace(request.URL.Query().Get("q"))
	status := strings.ToLower(strings.TrimSpace(request.URL.Query().Get("status")))
	if len([]rune(query)) > 120 {
		writeAPIError(writer, http.StatusBadRequest, "invalid-query", "The RFC search is too long.")
		return
	}
	items, err := handler.Service.Repository().Proposals(query, status)
	if err != nil {
		writeAPIError(writer, http.StatusInternalServerError, "rfcs-unavailable", "RFCs could not be loaded.")
		return
	}
	_ = xmlcodec.Write(writer, http.StatusOK, RFCsResponse{Items: items})
}

func (handler RFCHandler) Get(writer http.ResponseWriter, request *http.Request) {
	number, ok := handler.number(writer, request)
	if !ok {
		return
	}
	item, err := handler.Service.Repository().Proposal(number)
	if errors.Is(err, storage.ErrNotFound) {
		writeAPIError(writer, http.StatusNotFound, "rfc-not-found", "The RFC was not found.")
		return
	}
	if err != nil {
		writeAPIError(writer, http.StatusInternalServerError, "rfcs-unavailable", "The RFC could not be loaded.")
		return
	}
	_ = xmlcodec.Write(writer, http.StatusOK, item)
}

func (handler RFCHandler) Create(writer http.ResponseWriter, request *http.Request) {
	actorID, ok := handler.authorize(writer, request)
	if !ok {
		return
	}
	var input rfcdomain.ProposalInput
	if xmlcodec.Decode(request.Body, &input) != nil {
		writeAPIError(writer, http.StatusBadRequest, "invalid-xml", "The RFC request is not valid XML.")
		return
	}
	item, err := handler.Service.Create(actorID, input)
	handler.writeProposal(writer, http.StatusCreated, item, err)
}

func (handler RFCHandler) Update(writer http.ResponseWriter, request *http.Request) {
	actorID, ok := handler.authorize(writer, request)
	if !ok {
		return
	}
	number, ok := handler.number(writer, request)
	if !ok {
		return
	}
	var input rfcdomain.ProposalInput
	if xmlcodec.Decode(request.Body, &input) != nil {
		writeAPIError(writer, http.StatusBadRequest, "invalid-xml", "The RFC request is not valid XML.")
		return
	}
	item, err := handler.Service.Update(actorID, number, input)
	handler.writeProposal(writer, http.StatusOK, item, err)
}

func (handler RFCHandler) UpdateStatus(writer http.ResponseWriter, request *http.Request) {
	actorID, ok := handler.authorize(writer, request)
	if !ok {
		return
	}
	number, ok := handler.number(writer, request)
	if !ok {
		return
	}
	var input rfcdomain.StatusInput
	if xmlcodec.Decode(request.Body, &input) != nil {
		writeAPIError(writer, http.StatusBadRequest, "invalid-xml", "The RFC status request is not valid XML.")
		return
	}
	item, err := handler.Service.UpdateStatus(actorID, number, input.Status)
	handler.writeProposal(writer, http.StatusOK, item, err)
}

func (handler RFCHandler) Comment(writer http.ResponseWriter, request *http.Request) {
	actorID, ok := handler.authorize(writer, request)
	if !ok {
		return
	}
	number, ok := handler.number(writer, request)
	if !ok {
		return
	}
	var input rfcdomain.CommentInput
	if xmlcodec.Decode(request.Body, &input) != nil {
		writeAPIError(writer, http.StatusBadRequest, "invalid-xml", "The RFC comment request is not valid XML.")
		return
	}
	item, err := handler.Service.AddComment(actorID, number, input)
	if errors.Is(err, storage.ErrNotFound) {
		writeAPIError(writer, http.StatusNotFound, "rfc-not-found", "The RFC was not found.")
		return
	}
	if errors.Is(err, rfcdomain.ErrInvalidComment) {
		writeAPIError(writer, http.StatusUnprocessableEntity, "invalid-rfc-comment", err.Error())
		return
	}
	if err != nil {
		writeAPIError(writer, http.StatusInternalServerError, "rfc-comment-failed", "The RFC comment could not be saved.")
		return
	}
	_ = xmlcodec.Write(writer, http.StatusCreated, item)
}

func (handler RFCHandler) authorize(writer http.ResponseWriter, request *http.Request) (string, bool) {
	setPrivateResponseHeaders(writer)
	if handler.Service == nil || handler.Auth == nil {
		writeAPIError(writer, http.StatusServiceUnavailable, "rfcs-unavailable", "RFC writing is unavailable.")
		return "", false
	}
	actor, authenticated := AuthenticatedAccount(*handler.Auth, request)
	if !authenticated {
		writeAPIError(writer, http.StatusUnauthorized, "not-authenticated", "Authentication is required.")
		return "", false
	}
	if !sameOrigin(request) {
		writeAPIError(writer, http.StatusForbidden, "invalid-origin", "The request origin is not allowed.")
		return "", false
	}
	return actor.ID, true
}

func (handler RFCHandler) number(writer http.ResponseWriter, request *http.Request) (uint64, bool) {
	if handler.Service == nil {
		writeAPIError(writer, http.StatusServiceUnavailable, "rfcs-unavailable", "RFCs are unavailable.")
		return 0, false
	}
	number, err := strconv.ParseUint(request.PathValue("number"), 10, 64)
	if err != nil || number == 0 {
		writeAPIError(writer, http.StatusBadRequest, "invalid-rfc-number", "The RFC number is invalid.")
		return 0, false
	}
	return number, true
}

func (handler RFCHandler) writeProposal(writer http.ResponseWriter, status int, item rfcdomain.Proposal, err error) {
	if errors.Is(err, storage.ErrNotFound) {
		writeAPIError(writer, http.StatusNotFound, "rfc-not-found", "The RFC was not found.")
		return
	}
	if errors.Is(err, rfcdomain.ErrForbidden) {
		writeAPIError(writer, http.StatusForbidden, "rfc-forbidden", err.Error())
		return
	}
	if errors.Is(err, rfcdomain.ErrInvalidProposal) || errors.Is(err, rfcdomain.ErrInvalidStatus) {
		writeAPIError(writer, http.StatusUnprocessableEntity, "invalid-rfc", err.Error())
		return
	}
	if err != nil {
		writeAPIError(writer, http.StatusInternalServerError, "rfc-save-failed", "The RFC could not be saved.")
		return
	}
	_ = xmlcodec.Write(writer, status, item)
}
