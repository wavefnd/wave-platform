package handler

import (
	"encoding/xml"
	"errors"
	"net/http"

	"github.com/wavefnd/wave-platform/internal/storage"
	webhookdomain "github.com/wavefnd/wave-platform/internal/webhook"
	"github.com/wavefnd/wave-platform/internal/xmlcodec"
)

type WebhookHandler struct {
	Service *webhookdomain.Service
	Auth    *AuthHandler
}

type WebhooksResponse struct {
	XMLName    xml.Name                     `xml:"https://wave-lang.dev/ns/platform/api/v1 webhooks"`
	Events     []string                     `xml:"supported-events>event"`
	Endpoints  []webhookdomain.EndpointView `xml:"endpoints>webhook"`
	Deliveries []webhookdomain.Delivery     `xml:"deliveries>delivery"`
}

func (handler WebhookHandler) List(writer http.ResponseWriter, request *http.Request) {
	if _, ok := handler.authorize(writer, request, false); !ok {
		return
	}
	endpoints, err := handler.Service.Endpoints()
	if err != nil {
		writeAPIError(writer, http.StatusInternalServerError, "webhooks-unavailable", "Webhooks could not be loaded.")
		return
	}
	deliveries, err := handler.Service.Deliveries(50)
	if err != nil {
		writeAPIError(writer, http.StatusInternalServerError, "webhooks-unavailable", "Webhook deliveries could not be loaded.")
		return
	}
	_ = xmlcodec.Write(writer, http.StatusOK, WebhooksResponse{Events: webhookdomain.SupportedEvents(), Endpoints: endpoints, Deliveries: deliveries})
}

func (handler WebhookHandler) Save(writer http.ResponseWriter, request *http.Request) {
	actorID, ok := handler.authorize(writer, request, true)
	if !ok {
		return
	}
	var input webhookdomain.EndpointInput
	if xmlcodec.Decode(request.Body, &input) != nil {
		writeAPIError(writer, http.StatusBadRequest, "invalid-xml", "The webhook request is not valid XML.")
		return
	}
	item, err := handler.Service.SaveEndpointScoped(actorID, input, true)
	if errors.Is(err, webhookdomain.ErrInvalidEndpoint) {
		writeAPIError(writer, http.StatusUnprocessableEntity, "invalid-webhook", err.Error())
		return
	}
	if errors.Is(err, storage.ErrNotFound) {
		writeAPIError(writer, http.StatusNotFound, "webhook-not-found", "The webhook was not found.")
		return
	}
	if err != nil {
		writeAPIError(writer, http.StatusInternalServerError, "webhook-save-failed", "The webhook could not be saved.")
		return
	}
	_ = xmlcodec.Write(writer, http.StatusOK, item)
}

func (handler WebhookHandler) Delete(writer http.ResponseWriter, request *http.Request) {
	actorID, ok := handler.authorize(writer, request, true)
	if !ok {
		return
	}
	if err := handler.Service.DeleteEndpointScoped(actorID, request.PathValue("webhook"), true); errors.Is(err, storage.ErrNotFound) {
		writeAPIError(writer, http.StatusNotFound, "webhook-not-found", "The webhook was not found.")
	} else if err != nil {
		writeAPIError(writer, http.StatusInternalServerError, "webhook-delete-failed", "The webhook could not be deleted.")
	} else {
		writer.WriteHeader(http.StatusNoContent)
	}
}

func (handler WebhookHandler) Test(writer http.ResponseWriter, request *http.Request) {
	actorID, ok := handler.authorize(writer, request, true)
	if !ok {
		return
	}
	delivery, err := handler.Service.TestEndpointScoped(request.Context(), actorID, request.PathValue("webhook"), true)
	if errors.Is(err, storage.ErrNotFound) {
		writeAPIError(writer, http.StatusNotFound, "webhook-not-found", "The webhook was not found.")
		return
	}
	if err != nil {
		writeAPIError(writer, http.StatusBadGateway, "webhook-delivery-failed", delivery.LastError)
		return
	}
	_ = xmlcodec.Write(writer, http.StatusOK, delivery)
}

func (handler WebhookHandler) UserList(writer http.ResponseWriter, request *http.Request) {
	actorID, ok := handler.userAuthorize(writer, request, false)
	if !ok {
		return
	}
	endpoints, err := handler.Service.EndpointsFor(actorID)
	if err != nil {
		writeAPIError(writer, http.StatusInternalServerError, "webhooks-unavailable", "Webhooks could not be loaded.")
		return
	}
	deliveries, err := handler.Service.DeliveriesFor(actorID, 30)
	if err != nil {
		writeAPIError(writer, http.StatusInternalServerError, "webhooks-unavailable", "Webhook deliveries could not be loaded.")
		return
	}
	_ = xmlcodec.Write(writer, http.StatusOK, WebhooksResponse{Events: webhookdomain.SupportedEvents(), Endpoints: endpoints, Deliveries: deliveries})
}

func (handler WebhookHandler) UserSave(writer http.ResponseWriter, request *http.Request) {
	actorID, ok := handler.userAuthorize(writer, request, true)
	if !ok {
		return
	}
	var input webhookdomain.EndpointInput
	if xmlcodec.Decode(request.Body, &input) != nil {
		writeAPIError(writer, http.StatusBadRequest, "invalid-xml", "The webhook request is not valid XML.")
		return
	}
	item, err := handler.Service.SaveEndpointScoped(actorID, input, false)
	if errors.Is(err, webhookdomain.ErrForbidden) {
		writeAPIError(writer, http.StatusForbidden, "webhook-forbidden", "The webhook does not belong to this account.")
		return
	}
	if errors.Is(err, webhookdomain.ErrInvalidEndpoint) {
		writeAPIError(writer, http.StatusUnprocessableEntity, "invalid-webhook", err.Error())
		return
	}
	if errors.Is(err, storage.ErrNotFound) {
		writeAPIError(writer, http.StatusNotFound, "webhook-not-found", "The webhook was not found.")
		return
	}
	if err != nil {
		writeAPIError(writer, http.StatusInternalServerError, "webhook-save-failed", "The webhook could not be saved.")
		return
	}
	_ = xmlcodec.Write(writer, http.StatusOK, item)
}

func (handler WebhookHandler) UserDelete(writer http.ResponseWriter, request *http.Request) {
	actorID, ok := handler.userAuthorize(writer, request, true)
	if !ok {
		return
	}
	err := handler.Service.DeleteEndpointScoped(actorID, request.PathValue("webhook"), false)
	if errors.Is(err, webhookdomain.ErrForbidden) {
		writeAPIError(writer, http.StatusForbidden, "webhook-forbidden", "The webhook does not belong to this account.")
	} else if errors.Is(err, storage.ErrNotFound) {
		writeAPIError(writer, http.StatusNotFound, "webhook-not-found", "The webhook was not found.")
	} else if err != nil {
		writeAPIError(writer, http.StatusInternalServerError, "webhook-delete-failed", "The webhook could not be deleted.")
	} else {
		writer.WriteHeader(http.StatusNoContent)
	}
}

func (handler WebhookHandler) UserTest(writer http.ResponseWriter, request *http.Request) {
	actorID, ok := handler.userAuthorize(writer, request, true)
	if !ok {
		return
	}
	delivery, err := handler.Service.TestEndpointScoped(request.Context(), actorID, request.PathValue("webhook"), false)
	if errors.Is(err, webhookdomain.ErrForbidden) {
		writeAPIError(writer, http.StatusForbidden, "webhook-forbidden", "The webhook does not belong to this account.")
		return
	}
	if errors.Is(err, storage.ErrNotFound) {
		writeAPIError(writer, http.StatusNotFound, "webhook-not-found", "The webhook was not found.")
		return
	}
	if err != nil {
		writeAPIError(writer, http.StatusBadGateway, "webhook-delivery-failed", delivery.LastError)
		return
	}
	_ = xmlcodec.Write(writer, http.StatusOK, delivery)
}

func (handler WebhookHandler) userAuthorize(writer http.ResponseWriter, request *http.Request, mutation bool) (string, bool) {
	setPrivateResponseHeaders(writer)
	if handler.Service == nil || handler.Auth == nil {
		writeAPIError(writer, http.StatusServiceUnavailable, "webhooks-unavailable", "Webhooks are unavailable.")
		return "", false
	}
	actor, authenticated := AuthenticatedAccount(*handler.Auth, request)
	if !authenticated {
		writeAPIError(writer, http.StatusUnauthorized, "not-authenticated", "Authentication is required.")
		return "", false
	}
	if mutation && !sameOrigin(request) {
		writeAPIError(writer, http.StatusForbidden, "invalid-origin", "The request origin is not allowed.")
		return "", false
	}
	return actor.ID, true
}

func (handler WebhookHandler) authorize(writer http.ResponseWriter, request *http.Request, mutation bool) (string, bool) {
	setPrivateResponseHeaders(writer)
	if handler.Service == nil || handler.Auth == nil {
		writeAPIError(writer, http.StatusServiceUnavailable, "webhooks-unavailable", "Webhook administration is unavailable.")
		return "", false
	}
	actor, authenticated := AuthenticatedAccount(*handler.Auth, request)
	if !authenticated {
		writeAPIError(writer, http.StatusUnauthorized, "not-authenticated", "Authentication is required.")
		return "", false
	}
	if !handler.Auth.Service.IsAdministrator(actor.ID) {
		writeAPIError(writer, http.StatusForbidden, "administrator-required", "Administrator access is required.")
		return "", false
	}
	if mutation && !sameOrigin(request) {
		writeAPIError(writer, http.StatusForbidden, "invalid-origin", "The request origin is not allowed.")
		return "", false
	}
	return actor.ID, true
}
