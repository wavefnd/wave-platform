package handler

import (
	"encoding/xml"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/wavefnd/wave-platform/internal/account"
	mailingdomain "github.com/wavefnd/wave-platform/internal/mailinglist"
	"github.com/wavefnd/wave-platform/internal/storage"
	"github.com/wavefnd/wave-platform/internal/xmlcodec"
)

type MailingListsResponse struct {
	XMLName xml.Name                    `xml:"https://wave-lang.dev/ns/platform/api/v1 mailing-lists"`
	Items   []mailingdomain.ListSummary `xml:"list"`
}

type MailingListThreadsResponse struct {
	XMLName xml.Name                      `xml:"https://wave-lang.dev/ns/platform/api/v1 mailing-list-threads"`
	ListID  string                        `xml:"list-id"`
	Items   []mailingdomain.ThreadSummary `xml:"thread"`
}

type MailingListSubscriptionRequest struct {
	XMLName    xml.Name `xml:"https://wave-lang.dev/ns/platform/api/v1 mailing-list-subscription"`
	Subscribed bool     `xml:"subscribed"`
}

type MailingListPostRequest struct {
	XMLName xml.Name `xml:"https://wave-lang.dev/ns/platform/api/v1 mailing-list-post"`
	Subject string   `xml:"subject"`
	Body    string   `xml:"body"`
}

type MailingListReplyRequest struct {
	XMLName         xml.Name `xml:"https://wave-lang.dev/ns/platform/api/v1 mailing-list-reply"`
	ParentMessageID string   `xml:"parent-message-id,omitempty"`
	Body            string   `xml:"body"`
}

type MailingListHandler struct {
	Service *mailingdomain.Service
	Auth    *AuthHandler
}

func (handler MailingListHandler) Lists(writer http.ResponseWriter, request *http.Request) {
	actor, ok := handler.authorize(writer, request, false)
	if !ok {
		return
	}
	items, err := handler.Service.Lists(actor.ID)
	if err != nil {
		handler.writeError(writer, err)
		return
	}
	_ = xmlcodec.Write(writer, http.StatusOK, MailingListsResponse{Items: items})
}

func (handler MailingListHandler) Subscriptions(writer http.ResponseWriter, request *http.Request) {
	actor, ok := handler.authorize(writer, request, false)
	if !ok {
		return
	}
	items, err := handler.Service.Subscriptions(actor.ID)
	if err != nil {
		handler.writeError(writer, err)
		return
	}
	_ = xmlcodec.Write(writer, http.StatusOK, MailingListsResponse{Items: items})
}

func (handler MailingListHandler) Threads(writer http.ResponseWriter, request *http.Request) {
	actor, ok := handler.authorize(writer, request, false)
	if !ok {
		return
	}
	query := strings.TrimSpace(request.URL.Query().Get("q"))
	if len([]rune(query)) > 200 {
		writeAPIError(writer, http.StatusBadRequest, "invalid-query", "The mailing list search is too long.")
		return
	}
	limit, _ := strconv.Atoi(request.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(request.URL.Query().Get("offset"))
	if limit <= 0 || limit > 100 {
		limit = 30
	}
	if offset < 0 {
		offset = 0
	}
	listID := request.PathValue("list")
	items, err := handler.Service.Threads(actor.ID, listID, query, limit, offset)
	if err != nil {
		handler.writeError(writer, err)
		return
	}
	_ = xmlcodec.Write(writer, http.StatusOK, MailingListThreadsResponse{ListID: listID, Items: items})
}

func (handler MailingListHandler) Thread(writer http.ResponseWriter, request *http.Request) {
	actor, ok := handler.authorize(writer, request, false)
	if !ok {
		return
	}
	item, err := handler.Service.Thread(actor.ID, request.PathValue("list"), request.PathValue("thread"))
	if err != nil {
		handler.writeError(writer, err)
		return
	}
	_ = xmlcodec.Write(writer, http.StatusOK, item)
}

func (handler MailingListHandler) Subscription(writer http.ResponseWriter, request *http.Request) {
	actor, ok := handler.authorize(writer, request, true)
	if !ok {
		return
	}
	var input MailingListSubscriptionRequest
	if xmlcodec.Decode(request.Body, &input) != nil {
		writeAPIError(writer, http.StatusBadRequest, "invalid-xml", "The mailing list subscription request is not valid XML.")
		return
	}
	if err := handler.Service.Subscribe(actor.ID, request.PathValue("list"), input.Subscribed); err != nil {
		handler.writeError(writer, err)
		return
	}
	writer.WriteHeader(http.StatusNoContent)
}

func (handler MailingListHandler) Post(writer http.ResponseWriter, request *http.Request) {
	actor, ok := handler.authorize(writer, request, true)
	if !ok {
		return
	}
	var input MailingListPostRequest
	if xmlcodec.Decode(request.Body, &input) != nil {
		writeAPIError(writer, http.StatusBadRequest, "invalid-xml", "The mailing list post is not valid XML.")
		return
	}
	item, err := handler.Service.Post(actor.ID, request.PathValue("list"), mailingdomain.PostInput{Subject: input.Subject, Body: input.Body})
	if err != nil {
		handler.writeError(writer, err)
		return
	}
	writer.Header().Set("Location", "/api/v1/mailing-lists/"+item.ListID+"/threads/"+item.ID)
	_ = xmlcodec.Write(writer, http.StatusCreated, item)
}

func (handler MailingListHandler) Reply(writer http.ResponseWriter, request *http.Request) {
	actor, ok := handler.authorize(writer, request, true)
	if !ok {
		return
	}
	var input MailingListReplyRequest
	if xmlcodec.Decode(request.Body, &input) != nil {
		writeAPIError(writer, http.StatusBadRequest, "invalid-xml", "The mailing list reply is not valid XML.")
		return
	}
	item, err := handler.Service.Reply(actor.ID, request.PathValue("list"), request.PathValue("thread"),
		mailingdomain.ReplyInput{ParentMessageID: input.ParentMessageID, Body: input.Body})
	if err != nil {
		handler.writeError(writer, err)
		return
	}
	_ = xmlcodec.Write(writer, http.StatusCreated, item)
}

func (handler MailingListHandler) authorize(writer http.ResponseWriter, request *http.Request, mutation bool) (account.Account, bool) {
	setPrivateResponseHeaders(writer)
	writer.Header().Set("X-Robots-Tag", "noindex, nofollow, noarchive")
	if handler.Service == nil || handler.Auth == nil {
		writeAPIError(writer, http.StatusServiceUnavailable, "mailing-lists-unavailable", "Mailing lists are unavailable.")
		return account.Account{}, false
	}
	actor, ok := AuthenticatedAccount(*handler.Auth, request)
	if !ok {
		writeAPIError(writer, http.StatusUnauthorized, "not-authenticated", "Authentication is required.")
		return account.Account{}, false
	}
	if mutation && !sameOrigin(request) {
		writeAPIError(writer, http.StatusForbidden, "invalid-origin", "The request origin is not allowed.")
		return account.Account{}, false
	}
	return actor, true
}

func (handler MailingListHandler) writeError(writer http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, storage.ErrNotFound):
		writeAPIError(writer, http.StatusNotFound, "mailing-list-not-found", "The mailing list or thread was not found.")
	case errors.Is(err, mailingdomain.ErrNotSubscribed):
		writeAPIError(writer, http.StatusForbidden, "mailing-list-subscription-required", "Subscribe to this mailing list before posting.")
	case errors.Is(err, mailingdomain.ErrForbidden):
		writeAPIError(writer, http.StatusForbidden, "mailing-list-forbidden", "This Wave account cannot perform that mailing list action.")
	case errors.Is(err, mailingdomain.ErrInvalidPost):
		writeAPIError(writer, http.StatusUnprocessableEntity, "invalid-mailing-list-post", strings.TrimPrefix(err.Error(), mailingdomain.ErrInvalidPost.Error()+": "))
	case errors.Is(err, mailingdomain.ErrRateLimited):
		writer.Header().Set("Retry-After", "3600")
		writeAPIError(writer, http.StatusTooManyRequests, "mailing-list-rate-limited", "The daily posting limit for this mailing list has been reached.")
	case errors.Is(err, mailingdomain.ErrListFull):
		writeAPIError(writer, http.StatusInsufficientStorage, "mailing-list-full", "This mailing list has reached its storage quota.")
	default:
		writeAPIError(writer, http.StatusInternalServerError, "mailing-list-failed", "The mailing list request could not be completed.")
	}
}
