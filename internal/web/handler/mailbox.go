package handler

import (
	"encoding/xml"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/wavefnd/wave-platform/internal/identity"
	"github.com/wavefnd/wave-platform/internal/storage"
	"github.com/wavefnd/wave-platform/internal/xmlcodec"
)

type MailboxHandler struct{ Auth AuthHandler }

type MailboxResponse struct {
	XMLName xml.Name          `xml:"https://wave-lang.dev/ns/platform/api/v1 mailbox"`
	Address string            `xml:"address"`
	Folder  string            `xml:"folder"`
	Items   []MailboxItemView `xml:"items>item"`
}

type MailboxItemView struct {
	ID             string    `xml:"id,attr"`
	MessageID      string    `xml:"message-id"`
	From           string    `xml:"from"`
	To             []string  `xml:"to"`
	Subject        string    `xml:"subject"`
	ReceivedAt     time.Time `xml:"received-at"`
	Preview        string    `xml:"preview"`
	Flags          []string  `xml:"flags>flag,omitempty"`
	DeliveryStatus string    `xml:"delivery-status,omitempty"`
}

type MailMessageResponse struct {
	XMLName        xml.Name  `xml:"https://wave-lang.dev/ns/platform/api/v1 mail-message"`
	EntryID        string    `xml:"entry-id"`
	MessageID      string    `xml:"message-id"`
	From           string    `xml:"from"`
	To             []string  `xml:"to"`
	Cc             []string  `xml:"cc,omitempty"`
	Subject        string    `xml:"subject"`
	Date           time.Time `xml:"date"`
	Body           string    `xml:"body"`
	Flags          []string  `xml:"flags>flag,omitempty"`
	DeliveryStatus string    `xml:"delivery-status,omitempty"`
}

type SendMailRequest struct {
	XMLName xml.Name `xml:"https://wave-lang.dev/ns/platform/api/v1 send-mail"`
	To      string   `xml:"to"`
	Subject string   `xml:"subject"`
	Body    string   `xml:"body"`
}

type MailboxActionRequest struct {
	XMLName xml.Name `xml:"https://wave-lang.dev/ns/platform/api/v1 mailbox-action"`
	Action  string   `xml:"action"`
}

func (handler MailboxHandler) List(writer http.ResponseWriter, request *http.Request) {
	setPrivateResponseHeaders(writer)
	account, ok := AuthenticatedAccount(handler.Auth, request)
	if !ok {
		writeAPIError(writer, http.StatusUnauthorized, "not-authenticated", "Authentication is required.")
		return
	}
	folder := strings.TrimSpace(request.URL.Query().Get("folder"))
	query := strings.ToLower(strings.TrimSpace(request.URL.Query().Get("q")))
	if len([]rune(query)) > 200 {
		writeAPIError(writer, http.StatusBadRequest, "invalid-query", "The mailbox search is too long.")
		return
	}
	if folder == "" {
		folder = "Inbox"
	}
	box, err := handler.Auth.Service.Mailbox(account.ID)
	if err != nil {
		writeAPIError(writer, http.StatusInternalServerError, "mailbox-unavailable", "The mailbox could not be loaded.")
		return
	}
	items, err := handler.Auth.Service.MailboxItems(account.ID, folder)
	if err != nil {
		writeAPIError(writer, http.StatusInternalServerError, "mailbox-unavailable", "The mailbox could not be loaded.")
		return
	}
	views := make([]MailboxItemView, 0, len(items))
	for _, item := range items {
		if query != "" {
			haystack := strings.ToLower(item.Message.From + " " + strings.Join(item.Message.To, " ") + " " + item.Message.Subject + " " + item.Body)
			if !strings.Contains(haystack, query) {
				continue
			}
		}
		views = append(views, mailboxView(item))
	}
	_ = xmlcodec.Write(writer, http.StatusOK, MailboxResponse{Address: box.Address, Folder: folder, Items: views})
}

func (handler MailboxHandler) Message(writer http.ResponseWriter, request *http.Request) {
	setPrivateResponseHeaders(writer)
	account, ok := AuthenticatedAccount(handler.Auth, request)
	if !ok {
		writeAPIError(writer, http.StatusUnauthorized, "not-authenticated", "Authentication is required.")
		return
	}
	item, err := handler.Auth.Service.MailboxItem(account.ID, request.PathValue("entry"))
	if errors.Is(err, storage.ErrNotFound) {
		writeAPIError(writer, http.StatusNotFound, "message-not-found", "The mailbox message was not found.")
		return
	}
	if err != nil {
		writeAPIError(writer, http.StatusInternalServerError, "mailbox-unavailable", "The mailbox message could not be loaded.")
		return
	}
	_ = xmlcodec.Write(writer, http.StatusOK, mailMessageView(item))
}

func (handler MailboxHandler) Send(writer http.ResponseWriter, request *http.Request) {
	setPrivateResponseHeaders(writer)
	account, ok := AuthenticatedAccount(handler.Auth, request)
	if !ok {
		writeAPIError(writer, http.StatusUnauthorized, "not-authenticated", "Authentication is required.")
		return
	}
	if !sameOrigin(request) {
		writeAPIError(writer, http.StatusForbidden, "invalid-origin", "The request origin is not allowed.")
		return
	}
	var input SendMailRequest
	if err := xmlcodec.Decode(request.Body, &input); err != nil {
		writeAPIError(writer, http.StatusBadRequest, "invalid-xml", "The mail request is not valid XML.")
		return
	}
	item, err := handler.Auth.Service.SendMail(account, identity.OutgoingMail{To: input.To, Subject: input.Subject, Body: input.Body})
	if errors.Is(err, identity.ErrInvalidMail) {
		writeAPIError(writer, http.StatusUnprocessableEntity, "invalid-mail", strings.TrimPrefix(err.Error(), identity.ErrInvalidMail.Error()+": "))
		return
	}
	if err != nil {
		writeAPIError(writer, http.StatusInternalServerError, "mail-send-failed", "The message could not be sent.")
		return
	}
	writer.Header().Set("Location", "/api/v1/mailbox/messages/"+item.Entry.ID)
	_ = xmlcodec.Write(writer, http.StatusCreated, mailMessageView(item))
}

func (handler MailboxHandler) Action(writer http.ResponseWriter, request *http.Request) {
	setPrivateResponseHeaders(writer)
	account, ok := AuthenticatedAccount(handler.Auth, request)
	if !ok {
		writeAPIError(writer, http.StatusUnauthorized, "not-authenticated", "Authentication is required.")
		return
	}
	if !sameOrigin(request) {
		writeAPIError(writer, http.StatusForbidden, "invalid-origin", "The request origin is not allowed.")
		return
	}
	var input MailboxActionRequest
	if err := xmlcodec.Decode(request.Body, &input); err != nil {
		writeAPIError(writer, http.StatusBadRequest, "invalid-xml", "The mailbox action is not valid XML.")
		return
	}
	item, err := handler.Auth.Service.UpdateMailboxEntry(account.ID, request.PathValue("entry"), strings.TrimSpace(input.Action))
	if errors.Is(err, storage.ErrNotFound) {
		writeAPIError(writer, http.StatusNotFound, "message-not-found", "The mailbox message was not found.")
		return
	}
	if errors.Is(err, identity.ErrInvalidMail) {
		writeAPIError(writer, http.StatusUnprocessableEntity, "invalid-action", strings.TrimPrefix(err.Error(), identity.ErrInvalidMail.Error()+": "))
		return
	}
	if err != nil {
		writeAPIError(writer, http.StatusInternalServerError, "mailbox-action-failed", "The mailbox action could not be completed.")
		return
	}
	_ = xmlcodec.Write(writer, http.StatusOK, mailMessageView(item))
}

func mailboxView(item identity.MailboxItem) MailboxItemView {
	when := item.Message.ReceivedAt
	if when.IsZero() {
		when = item.Message.CreatedAt
	}
	return MailboxItemView{ID: item.Entry.ID, MessageID: item.Message.ID, From: item.Message.From,
		To: item.Message.To, Subject: item.Message.Subject, ReceivedAt: when, Preview: mailPreview(item.Body, 140),
		Flags: item.Entry.Flags, DeliveryStatus: item.DeliveryStatus}
}

func mailMessageView(item identity.MailboxItem) MailMessageResponse {
	when := item.Message.ReceivedAt
	if when.IsZero() {
		when = item.Message.CreatedAt
	}
	return MailMessageResponse{EntryID: item.Entry.ID, MessageID: item.Message.ID, From: item.Message.From,
		To: item.Message.To, Cc: item.Message.Cc, Subject: item.Message.Subject, Date: when, Body: item.Body,
		Flags: item.Entry.Flags, DeliveryStatus: item.DeliveryStatus}
}

func mailPreview(value string, limit int) string {
	value = strings.Join(strings.Fields(value), " ")
	runes := []rune(value)
	if len(runes) <= limit {
		return value
	}
	return string(runes[:limit]) + "…"
}
