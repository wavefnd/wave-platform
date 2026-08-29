package handler

import (
	"encoding/xml"
	"errors"
	"net/http"
	"strconv"

	"github.com/wavefnd/wave-platform/internal/account"
	notificationdomain "github.com/wavefnd/wave-platform/internal/notification"
	"github.com/wavefnd/wave-platform/internal/storage"
	"github.com/wavefnd/wave-platform/internal/xmlcodec"
)

type NotificationHandler struct {
	Service *notificationdomain.Service
	Auth    *AuthHandler
}

type notificationListResponse struct {
	XMLName xml.Name                  `xml:"https://wave-lang.dev/ns/platform/api/v1 notification-list"`
	Unread  int                       `xml:"unread-count"`
	Items   []notificationdomain.Item `xml:"notifications>notification"`
}

type notificationResponse struct {
	XMLName xml.Name                `xml:"https://wave-lang.dev/ns/platform/api/v1 notification-result"`
	Item    notificationdomain.Item `xml:"notification"`
}

func (handler NotificationHandler) List(writer http.ResponseWriter, request *http.Request) {
	actor, ok := handler.authorize(writer, request, false)
	if !ok {
		return
	}
	limit, _ := strconv.Atoi(request.URL.Query().Get("limit"))
	items, unread, err := handler.Service.List(actor.ID, limit)
	if err != nil {
		writeAPIError(writer, http.StatusInternalServerError, "notifications-unavailable", "Notifications could not be loaded.")
		return
	}
	_ = xmlcodec.Write(writer, http.StatusOK, notificationListResponse{Unread: unread, Items: items})
}

func (handler NotificationHandler) MarkRead(writer http.ResponseWriter, request *http.Request) {
	actor, ok := handler.authorize(writer, request, true)
	if !ok {
		return
	}
	item, err := handler.Service.MarkRead(actor.ID, request.PathValue("notification"))
	if errors.Is(err, storage.ErrNotFound) {
		writeAPIError(writer, http.StatusNotFound, "notification-not-found", "The notification was not found.")
		return
	}
	if err != nil {
		writeAPIError(writer, http.StatusInternalServerError, "notification-update-failed", "The notification could not be updated.")
		return
	}
	_ = xmlcodec.Write(writer, http.StatusOK, notificationResponse{Item: item})
}

func (handler NotificationHandler) MarkAllRead(writer http.ResponseWriter, request *http.Request) {
	actor, ok := handler.authorize(writer, request, true)
	if !ok {
		return
	}
	if err := handler.Service.MarkAllRead(actor.ID); err != nil {
		writeAPIError(writer, http.StatusInternalServerError, "notification-update-failed", "Notifications could not be updated.")
		return
	}
	writer.WriteHeader(http.StatusNoContent)
}

func (handler NotificationHandler) authorize(writer http.ResponseWriter, request *http.Request, mutation bool) (account.Account, bool) {
	setPrivateResponseHeaders(writer)
	writer.Header().Set("X-Robots-Tag", "noindex, nofollow, noarchive")
	if handler.Service == nil || handler.Auth == nil {
		writeAPIError(writer, http.StatusServiceUnavailable, "notifications-unavailable", "Notifications are unavailable.")
		return account.Account{}, false
	}
	actor, ok := AuthenticatedAccount(*handler.Auth, request)
	if !ok {
		writeAPIError(writer, http.StatusUnauthorized, "not-authenticated", "Authentication is required.")
		return account.Account{}, false
	}
	if mutation && !sameOrigin(request) {
		writeAPIError(writer, http.StatusForbidden, "origin-not-allowed", "The request origin is not allowed.")
		return account.Account{}, false
	}
	return actor, true
}
