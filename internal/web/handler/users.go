package handler

import (
	"encoding/xml"
	"errors"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/wavefnd/wave-platform/internal/account"
	communitydomain "github.com/wavefnd/wave-platform/internal/community"
	"github.com/wavefnd/wave-platform/internal/identity"
	questiondomain "github.com/wavefnd/wave-platform/internal/question"
	"github.com/wavefnd/wave-platform/internal/storage"
	"github.com/wavefnd/wave-platform/internal/xmlcodec"
)

type UsersHandler struct {
	Community *communitydomain.Repository
	Questions *questiondomain.Repository
	Auth      *AuthHandler
}

type UserActivity struct {
	Kind      string `xml:"kind"`
	Title     string `xml:"title"`
	Excerpt   string `xml:"excerpt"`
	URL       string `xml:"url"`
	CreatedAt string `xml:"created-at"`
}

type UserProfileResponse struct {
	XMLName              xml.Name       `xml:"https://wave-lang.dev/ns/platform/api/v1 user-profile"`
	Username             string         `xml:"username"`
	DisplayName          string         `xml:"display-name"`
	Email                string         `xml:"email"`
	Bio                  string         `xml:"bio,omitempty"`
	TimeZone             string         `xml:"time-zone"`
	JoinedAt             time.Time      `xml:"joined-at"`
	Activities           []UserActivity `xml:"activities>activity"`
	AddressChoiceAllowed bool           `xml:"address-choice-allowed"`
}

type UserDirectoryResponse struct {
	XMLName xml.Name              `xml:"https://wave-lang.dev/ns/platform/api/v1 users"`
	Items   []UserProfileResponse `xml:"user"`
}

type UserProfileUpdateRequest struct {
	DisplayName string `xml:"display-name"`
	Bio         string `xml:"bio"`
	TimeZone    string `xml:"time-zone"`
}

type UserAddressUpdateRequest struct {
	LocalPart string `xml:"local-part"`
	Code      string `xml:"code"`
}

func (handler UsersHandler) Profile(writer http.ResponseWriter, request *http.Request) {
	if handler.Auth == nil || handler.Auth.Service == nil {
		writeAPIError(writer, http.StatusServiceUnavailable, "users-unavailable", "The user directory is unavailable.")
		return
	}
	localPart := strings.ToLower(strings.TrimSpace(request.PathValue("user")))
	if localPart == "" || len([]rune(localPart)) > 60 || strings.ContainsAny(localPart, "/@") {
		writeAPIError(writer, http.StatusNotFound, "user-not-found", "The user was not found.")
		return
	}
	item, err := handler.Auth.Service.PublicAccount(localPart)
	if errors.Is(err, storage.ErrNotFound) || (err == nil && item.Status != "active") {
		writeAPIError(writer, http.StatusNotFound, "user-not-found", "The user was not found.")
		return
	}
	if err != nil {
		writeAPIError(writer, http.StatusInternalServerError, "users-unavailable", "The user profile could not be loaded.")
		return
	}
	handler.writeProfile(writer, item)
}

func (handler UsersHandler) ProfileByID(writer http.ResponseWriter, request *http.Request) {
	if handler.Auth == nil || handler.Auth.Service == nil {
		writeAPIError(writer, http.StatusServiceUnavailable, "users-unavailable", "The user directory is unavailable.")
		return
	}
	accountID := strings.TrimSpace(request.PathValue("account"))
	if accountID == "" || len(accountID) > 100 || strings.ContainsAny(accountID, "/") {
		writeAPIError(writer, http.StatusNotFound, "user-not-found", "The user was not found.")
		return
	}
	item, err := handler.Auth.Service.PublicAccountByID(accountID)
	if errors.Is(err, storage.ErrNotFound) || (err == nil && item.Status != "active") {
		writeAPIError(writer, http.StatusNotFound, "user-not-found", "The user was not found.")
		return
	}
	if err != nil {
		writeAPIError(writer, http.StatusInternalServerError, "users-unavailable", "The user profile could not be loaded.")
		return
	}
	handler.writeProfile(writer, item)
}

func (handler UsersHandler) writeProfile(writer http.ResponseWriter, item account.Account) {
	activities, err := handler.activities(item.ID)
	if err != nil {
		writeAPIError(writer, http.StatusInternalServerError, "users-unavailable", "The user activity could not be loaded.")
		return
	}
	_ = xmlcodec.Write(writer, http.StatusOK, UserProfileResponse{Username: item.Username, DisplayName: item.DisplayName,
		Email: item.Email, Bio: item.Bio, TimeZone: normalizedTimeZone(item.TimeZone), JoinedAt: item.CreatedAt, Activities: activities,
		AddressChoiceAllowed: handler.Auth.Service.AddressChoiceAllowed(item)})
}

func (handler UsersHandler) Directory(writer http.ResponseWriter, _ *http.Request) {
	if handler.Auth == nil || handler.Auth.Service == nil {
		writeAPIError(writer, http.StatusServiceUnavailable, "users-unavailable", "The user directory is unavailable.")
		return
	}
	accounts, err := handler.Auth.Service.PublicAccounts()
	if err != nil {
		writeAPIError(writer, http.StatusInternalServerError, "users-unavailable", "The user directory could not be loaded.")
		return
	}
	items := make([]UserProfileResponse, 0, len(accounts))
	for _, item := range accounts {
		items = append(items, UserProfileResponse{Username: item.Username, DisplayName: item.DisplayName,
			Email: item.Email, Bio: item.Bio, TimeZone: normalizedTimeZone(item.TimeZone), JoinedAt: item.CreatedAt,
			AddressChoiceAllowed: handler.Auth.Service.AddressChoiceAllowed(item)})
	}
	_ = xmlcodec.Write(writer, http.StatusOK, UserDirectoryResponse{Items: items})
}

func (handler UsersHandler) UpdateProfile(writer http.ResponseWriter, request *http.Request) {
	setPrivateResponseHeaders(writer)
	actor, ok := handler.requireAccount(writer, request)
	if !ok || !sameOrigin(request) {
		if ok {
			writeAPIError(writer, http.StatusForbidden, "invalid-origin", "The request origin is not allowed.")
		}
		return
	}
	var input UserProfileUpdateRequest
	if xmlcodec.Decode(request.Body, &input) != nil {
		writeAPIError(writer, http.StatusBadRequest, "invalid-xml", "The profile request is not valid XML.")
		return
	}
	item, err := handler.Auth.Service.UpdateProfile(actor.ID, input.DisplayName, input.Bio, input.TimeZone)
	if err != nil {
		handler.writeProfileError(writer, err)
		return
	}
	_ = xmlcodec.Write(writer, http.StatusOK, UserProfileResponse{Username: item.Username, DisplayName: item.DisplayName,
		Email: item.Email, Bio: item.Bio, TimeZone: normalizedTimeZone(item.TimeZone), JoinedAt: item.CreatedAt,
		AddressChoiceAllowed: handler.Auth.Service.AddressChoiceAllowed(item)})
}

func (handler UsersHandler) UpdateAddress(writer http.ResponseWriter, request *http.Request) {
	setPrivateResponseHeaders(writer)
	actor, ok := handler.requireAccount(writer, request)
	if !ok || !sameOrigin(request) {
		if ok {
			writeAPIError(writer, http.StatusForbidden, "invalid-origin", "The request origin is not allowed.")
		}
		return
	}
	var input UserAddressUpdateRequest
	if xmlcodec.Decode(request.Body, &input) != nil {
		writeAPIError(writer, http.StatusBadRequest, "invalid-xml", "The address request is not valid XML.")
		return
	}
	item, err := handler.Auth.Service.ChangeWaveAddress(actor.ID, input.Code, input.LocalPart)
	if err != nil {
		handler.writeProfileError(writer, err)
		return
	}
	_ = xmlcodec.Write(writer, http.StatusOK, UserProfileResponse{Username: item.Username, DisplayName: item.DisplayName,
		Email: item.Email, Bio: item.Bio, TimeZone: normalizedTimeZone(item.TimeZone), JoinedAt: item.CreatedAt,
		AddressChoiceAllowed: handler.Auth.Service.AddressChoiceAllowed(item)})
}

func (handler UsersHandler) requireAccount(writer http.ResponseWriter, request *http.Request) (account.Account, bool) {
	if handler.Auth == nil {
		writeAPIError(writer, http.StatusServiceUnavailable, "users-unavailable", "The user directory is unavailable.")
		return account.Account{}, false
	}
	actor, ok := AuthenticatedAccount(*handler.Auth, request)
	if !ok {
		writeAPIError(writer, http.StatusUnauthorized, "not-authenticated", "Authentication is required.")
	}
	return actor, ok
}

func (handler UsersHandler) writeProfileError(writer http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, account.ErrConflict):
		writeAPIError(writer, http.StatusConflict, "address-conflict", "That Wave mail address is already in use.")
	case errors.Is(err, identity.ErrInvalidCredentials):
		writeAPIError(writer, http.StatusUnauthorized, "invalid-code", "The authenticator code is incorrect.")
	case errors.Is(err, identity.ErrInvalidProfile):
		writeAPIError(writer, http.StatusUnprocessableEntity, "invalid-profile", strings.TrimPrefix(err.Error(), identity.ErrInvalidProfile.Error()+": "))
	default:
		writeAPIError(writer, http.StatusInternalServerError, "profile-update-failed", "The profile could not be updated.")
	}
}

func (handler UsersHandler) activities(accountID string) ([]UserActivity, error) {
	items := make([]UserActivity, 0)
	if handler.Community != nil {
		threads, err := handler.Community.QueryThreads("", "", "latest", 0, 0)
		if err != nil {
			return nil, err
		}
		for _, thread := range threads {
			threadURL := "/community/thread/" + thread.ID
			if thread.SpaceID == "founder-notes" || thread.SpaceID == "development-log" {
				threadURL = "/lunastev/thread/" + thread.ID
			}
			if thread.AuthorAccountID == accountID {
				items = append(items, UserActivity{Kind: "community-post", Title: thread.Title, Excerpt: thread.Excerpt,
					URL: threadURL, CreatedAt: thread.CreatedAt})
			}
			view, err := handler.Community.View(thread.ID)
			if err != nil {
				return nil, err
			}
			for _, reply := range view.Replies {
				if reply.AuthorAccountID == accountID {
					items = append(items, UserActivity{Kind: "community-comment", Title: thread.Title, Excerpt: activityExcerpt(reply.Body),
						URL: threadURL, CreatedAt: reply.CreatedAt})
				}
			}
		}
	}
	if handler.Questions != nil {
		questions, err := handler.Questions.Query("", "newest", "", 0, 0, "")
		if err != nil {
			return nil, err
		}
		for _, question := range questions {
			if question.AuthorAccountID == accountID {
				items = append(items, UserActivity{Kind: "question", Title: question.Title, Excerpt: question.Excerpt,
					URL: "/questions/" + question.ID, CreatedAt: question.CreatedAt})
			}
			view, err := handler.Questions.View(question.ID, "")
			if err != nil {
				return nil, err
			}
			for _, answer := range view.Answers {
				if answer.AuthorAccountID == accountID {
					items = append(items, UserActivity{Kind: "answer", Title: question.Title, Excerpt: activityExcerpt(answer.Body),
						URL: "/questions/" + question.ID, CreatedAt: answer.CreatedAt})
				}
			}
		}
	}
	sort.SliceStable(items, func(left, right int) bool { return items[left].CreatedAt > items[right].CreatedAt })
	if len(items) > 100 {
		items = items[:100]
	}
	return items, nil
}

func activityExcerpt(value string) string {
	value = strings.Join(strings.Fields(value), " ")
	runes := []rune(value)
	if len(runes) > 220 {
		return strings.TrimSpace(string(runes[:220])) + "…"
	}
	return value
}
