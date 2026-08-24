package handler

import (
	"encoding/xml"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/wavefnd/wave-platform/internal/account"
	communitydomain "github.com/wavefnd/wave-platform/internal/community"
	"github.com/wavefnd/wave-platform/internal/storage"
	"github.com/wavefnd/wave-platform/internal/xmlcodec"
)

type CommunitySpacesResponse struct {
	XMLName xml.Name                `xml:"https://wave-lang.dev/ns/platform/api/v1 community-spaces"`
	Items   []communitydomain.Space `xml:"space"`
}

type CommunityThreadsResponse struct {
	XMLName xml.Name                        `xml:"https://wave-lang.dev/ns/platform/api/v1 community-threads"`
	Items   []communitydomain.ThreadSummary `xml:"thread"`
}

type CommunityHandler struct {
	Repository *communitydomain.Repository
	Service    *communitydomain.Service
	Auth       *AuthHandler
}

type CreateCommunityPostRequest struct {
	XMLName xml.Name `xml:"https://wave-lang.dev/ns/platform/api/v1 community-post"`
	SpaceID string   `xml:"space-id"`
	Title   string   `xml:"title"`
	Body    string   `xml:"body"`
	Tags    []string `xml:"tags>tag,omitempty"`
}

type CreateCommunityReplyRequest struct {
	XMLName         xml.Name `xml:"https://wave-lang.dev/ns/platform/api/v1 community-reply"`
	ParentMessageID string   `xml:"parent-message-id,omitempty"`
	Body            string   `xml:"body"`
}

type CommunityVoteRequest struct {
	XMLName    xml.Name `xml:"https://wave-lang.dev/ns/platform/api/v1 community-vote"`
	TargetType string   `xml:"target-type"`
	TargetID   string   `xml:"target-id"`
	Value      int      `xml:"value"`
}

type CommunityVoteResponse struct {
	XMLName    xml.Name `xml:"https://wave-lang.dev/ns/platform/api/v1 community-vote-result"`
	Score      int64    `xml:"score"`
	ViewerVote int      `xml:"viewer-vote"`
}

type CommunitySubscriptionRequest struct {
	XMLName    xml.Name `xml:"https://wave-lang.dev/ns/platform/api/v1 community-subscription"`
	Subscribed bool     `xml:"subscribed"`
}

func (handler CommunityHandler) Spaces(writer http.ResponseWriter, _ *http.Request) {
	if handler.Repository == nil {
		http.Error(writer, "community repository is unavailable", http.StatusServiceUnavailable)
		return
	}
	items, err := handler.Repository.Spaces()
	if err != nil {
		http.Error(writer, "failed to load community spaces", http.StatusInternalServerError)
		return
	}
	if err := xmlcodec.Write(writer, http.StatusOK, CommunitySpacesResponse{Items: items}); err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
	}
}

func (handler CommunityHandler) Threads(writer http.ResponseWriter, request *http.Request) {
	if handler.Repository == nil {
		http.Error(writer, "community repository is unavailable", http.StatusServiceUnavailable)
		return
	}
	limit, _ := strconv.Atoi(request.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(request.URL.Query().Get("offset"))
	if len([]rune(request.URL.Query().Get("q"))) > 200 {
		writeAPIError(writer, http.StatusBadRequest, "invalid-query", "The community search is too long.")
		return
	}
	if limit <= 0 || limit > 100 {
		limit = 30
	}
	if offset < 0 {
		offset = 0
	}
	items, err := handler.Repository.QueryThreads(request.URL.Query().Get("space"), request.URL.Query().Get("q"),
		request.URL.Query().Get("sort"), limit, offset)
	if err != nil {
		http.Error(writer, "failed to load community posts", http.StatusInternalServerError)
		return
	}
	if handler.Auth != nil {
		if viewer, ok := AuthenticatedAccount(*handler.Auth, request); ok {
			for index := range items {
				items[index].ViewerVote, _ = handler.Repository.VoteValue("thread", items[index].ID, viewer.ID)
			}
		}
	}
	if err := xmlcodec.Write(writer, http.StatusOK, CommunityThreadsResponse{Items: items}); err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
	}
}

func (handler CommunityHandler) Thread(writer http.ResponseWriter, request *http.Request) {
	if handler.Repository == nil {
		http.Error(writer, "community repository is unavailable", http.StatusServiceUnavailable)
		return
	}
	viewerID := ""
	if handler.Auth != nil {
		if viewer, ok := AuthenticatedAccount(*handler.Auth, request); ok {
			viewerID = viewer.ID
		}
	}
	item, err := handler.Repository.ViewFor(request.PathValue("thread"), viewerID)
	if errors.Is(err, storage.ErrNotFound) {
		http.NotFound(writer, request)
		return
	}
	if err != nil {
		http.Error(writer, "failed to load community post", http.StatusInternalServerError)
		return
	}
	if count, viewErr := handler.Repository.RecordView(request.PathValue("thread")); viewErr == nil {
		item.ViewCount = count
	}
	if err := xmlcodec.Write(writer, http.StatusOK, item); err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
	}
}

func (handler CommunityHandler) CreatePost(writer http.ResponseWriter, request *http.Request) {
	actor, ok := handler.requireAccount(writer, request)
	if !ok {
		return
	}
	if !sameOrigin(request) {
		writeAPIError(writer, http.StatusForbidden, "invalid-origin", "The request origin is not allowed.")
		return
	}
	var input CreateCommunityPostRequest
	if err := xmlcodec.Decode(request.Body, &input); err != nil {
		writeAPIError(writer, http.StatusBadRequest, "invalid-xml", "The post request is not valid XML.")
		return
	}
	view, err := handler.Service.CreatePost(actor, communitydomain.CreatePostInput{
		SpaceID: input.SpaceID, Title: input.Title, Body: input.Body, Tags: input.Tags,
	})
	if err != nil {
		handler.writeCommunityError(writer, err)
		return
	}
	writer.Header().Set("Location", "/api/v1/community/threads/"+view.Thread.ID)
	_ = xmlcodec.Write(writer, http.StatusCreated, view)
}

func (handler CommunityHandler) CreateReply(writer http.ResponseWriter, request *http.Request) {
	actor, ok := handler.requireAccount(writer, request)
	if !ok {
		return
	}
	if !sameOrigin(request) {
		writeAPIError(writer, http.StatusForbidden, "invalid-origin", "The request origin is not allowed.")
		return
	}
	var input CreateCommunityReplyRequest
	if err := xmlcodec.Decode(request.Body, &input); err != nil {
		writeAPIError(writer, http.StatusBadRequest, "invalid-xml", "The reply request is not valid XML.")
		return
	}
	view, err := handler.Service.AddReply(actor, communitydomain.CreateReplyInput{
		ThreadID: request.PathValue("thread"), ParentMessageID: input.ParentMessageID, Body: input.Body,
	})
	if err != nil {
		handler.writeCommunityError(writer, err)
		return
	}
	_ = xmlcodec.Write(writer, http.StatusCreated, view)
}

func (handler CommunityHandler) Vote(writer http.ResponseWriter, request *http.Request) {
	actor, ok := handler.requireAccount(writer, request)
	if !ok {
		return
	}
	if !sameOrigin(request) {
		writeAPIError(writer, http.StatusForbidden, "invalid-origin", "The request origin is not allowed.")
		return
	}
	var input CommunityVoteRequest
	if err := xmlcodec.Decode(request.Body, &input); err != nil {
		writeAPIError(writer, http.StatusBadRequest, "invalid-xml", "The vote request is not valid XML.")
		return
	}
	score, err := handler.Service.Vote(actor.ID, request.PathValue("thread"), strings.TrimSpace(input.TargetType), strings.TrimSpace(input.TargetID), input.Value)
	if err != nil {
		handler.writeCommunityError(writer, err)
		return
	}
	_ = xmlcodec.Write(writer, http.StatusOK, CommunityVoteResponse{Score: score, ViewerVote: input.Value})
}

func (handler CommunityHandler) Subscribe(writer http.ResponseWriter, request *http.Request) {
	actor, ok := handler.requireAccount(writer, request)
	if !ok {
		return
	}
	if !sameOrigin(request) {
		writeAPIError(writer, http.StatusForbidden, "invalid-origin", "The request origin is not allowed.")
		return
	}
	var input CommunitySubscriptionRequest
	if err := xmlcodec.Decode(request.Body, &input); err != nil {
		writeAPIError(writer, http.StatusBadRequest, "invalid-xml", "The subscription request is not valid XML.")
		return
	}
	if err := handler.Service.Subscribe(actor.ID, request.PathValue("thread"), input.Subscribed); err != nil {
		handler.writeCommunityError(writer, err)
		return
	}
	writer.WriteHeader(http.StatusNoContent)
}

func (handler CommunityHandler) requireAccount(writer http.ResponseWriter, request *http.Request) (account.Account, bool) {
	setPrivateResponseHeaders(writer)
	if handler.Service == nil || handler.Auth == nil {
		writeAPIError(writer, http.StatusServiceUnavailable, "community-unavailable", "The community service is unavailable.")
		return account.Account{}, false
	}
	actor, ok := AuthenticatedAccount(*handler.Auth, request)
	if !ok {
		writeAPIError(writer, http.StatusUnauthorized, "not-authenticated", "Authentication is required.")
		return account.Account{}, false
	}
	return actor, true
}

func (handler CommunityHandler) writeCommunityError(writer http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, storage.ErrNotFound):
		writeAPIError(writer, http.StatusNotFound, "not-found", "The community post was not found.")
	case errors.Is(err, communitydomain.ErrThreadLocked):
		writeAPIError(writer, http.StatusConflict, "thread-locked", "This post is locked.")
	case errors.Is(err, communitydomain.ErrPostingRestricted):
		writeAPIError(writer, http.StatusForbidden, "posting-restricted", "Only the platform owner can publish here. Readers can participate through comments.")
	case errors.Is(err, communitydomain.ErrEnglishRequired):
		writeAPIError(writer, http.StatusUnprocessableEntity, "english-required", communitydomain.ErrEnglishRequired.Error()+". Code blocks and inline code are excluded from this check.")
	case errors.Is(err, communitydomain.ErrInvalidPost):
		writeAPIError(writer, http.StatusUnprocessableEntity, "invalid-post", strings.TrimPrefix(err.Error(), communitydomain.ErrInvalidPost.Error()+": "))
	default:
		writeAPIError(writer, http.StatusInternalServerError, "community-failed", "The community request could not be completed.")
	}
}
