package handler

import (
	"encoding/xml"
	"errors"
	"net/http"
	"strings"

	admindomain "github.com/wavefnd/wave-platform/internal/admin"
	"github.com/wavefnd/wave-platform/internal/storage"
	"github.com/wavefnd/wave-platform/internal/xmlcodec"
)

type PlatformPreferencesResponse struct {
	XMLName          xml.Name `xml:"https://wave-lang.dev/ns/platform/api/v1 platform-preferences"`
	LunaStevTimeZone string   `xml:"lunastev-time-zone"`
}

type AdministrationHandler struct {
	Service *admindomain.Service
	Auth    *AuthHandler
}

func (handler AdministrationHandler) Snapshot(writer http.ResponseWriter, request *http.Request) {
	actor, ok := handler.authorize(writer, request, false)
	if !ok {
		return
	}
	_ = actor
	query := strings.TrimSpace(request.URL.Query().Get("q"))
	if len([]rune(query)) > 120 {
		writeAPIError(writer, http.StatusBadRequest, "invalid-query", "The account search is too long.")
		return
	}
	snapshot, err := handler.Service.Snapshot(request.Context(), query)
	if err != nil {
		writeAPIError(writer, http.StatusInternalServerError, "administration-unavailable", "Administration data could not be loaded.")
		return
	}
	if err := xmlcodec.Write(writer, http.StatusOK, snapshot); err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
	}
}

func (handler AdministrationHandler) AccountStatus(writer http.ResponseWriter, request *http.Request) {
	actor, ok := handler.authorize(writer, request, true)
	if !ok {
		return
	}
	var input admindomain.AccountStatusInput
	if xmlcodec.Decode(request.Body, &input) != nil {
		writeAPIError(writer, http.StatusBadRequest, "invalid-xml", "The account status request is not valid XML.")
		return
	}
	err := handler.Service.UpdateAccountStatus(actor.ID, request.PathValue("account"), input.Status)
	if err != nil {
		handler.writeManagementError(writer, err)
		return
	}
	writer.WriteHeader(http.StatusNoContent)
}

func (handler AdministrationHandler) AccountRole(writer http.ResponseWriter, request *http.Request) {
	actor, ok := handler.authorize(writer, request, true)
	if !ok {
		return
	}
	var input admindomain.AccountRoleInput
	if xmlcodec.Decode(request.Body, &input) != nil {
		writeAPIError(writer, http.StatusBadRequest, "invalid-xml", "The account role request is not valid XML.")
		return
	}
	err := handler.Service.UpdateAdministrator(actor.ID, request.PathValue("account"), input.Administrator)
	if err != nil {
		handler.writeManagementError(writer, err)
		return
	}
	writer.WriteHeader(http.StatusNoContent)
}

func (handler AdministrationHandler) SourceMaintainer(writer http.ResponseWriter, request *http.Request) {
	actor, ok := handler.authorize(writer, request, true)
	if !ok {
		return
	}
	var input admindomain.SourceMaintainerInput
	if xmlcodec.Decode(request.Body, &input) != nil {
		writeAPIError(writer, http.StatusBadRequest, "invalid-xml", "The source maintainer request is not valid XML.")
		return
	}
	if err := handler.Service.UpdateSourceMaintainer(actor.ID, request.PathValue("account"), input.Enabled); err != nil {
		handler.writeManagementError(writer, err)
		return
	}
	writer.WriteHeader(http.StatusNoContent)
}

func (handler AdministrationHandler) RFCMaintainer(writer http.ResponseWriter, request *http.Request) {
	actor, ok := handler.authorize(writer, request, true)
	if !ok {
		return
	}
	var input admindomain.RFCMaintainerInput
	if xmlcodec.Decode(request.Body, &input) != nil {
		writeAPIError(writer, http.StatusBadRequest, "invalid-xml", "The RFC maintainer request is not valid XML.")
		return
	}
	if err := handler.Service.UpdateRFCMaintainer(actor.ID, request.PathValue("account"), input.Enabled); err != nil {
		handler.writeManagementError(writer, err)
		return
	}
	writer.WriteHeader(http.StatusNoContent)
}

func (handler AdministrationHandler) PlatformPreferences(writer http.ResponseWriter, _ *http.Request) {
	if handler.Service == nil {
		writeAPIError(writer, http.StatusServiceUnavailable, "preferences-unavailable", "Platform preferences are unavailable.")
		return
	}
	_ = xmlcodec.Write(writer, http.StatusOK, PlatformPreferencesResponse{LunaStevTimeZone: handler.Service.LunaStevTimeZone()})
}

func (handler AdministrationHandler) LunaStevTimeZone(writer http.ResponseWriter, request *http.Request) {
	actor, ok := handler.authorize(writer, request, true)
	if !ok {
		return
	}
	var input admindomain.TimeZoneInput
	if xmlcodec.Decode(request.Body, &input) != nil {
		writeAPIError(writer, http.StatusBadRequest, "invalid-xml", "The time zone request is not valid XML.")
		return
	}
	if err := handler.Service.UpdateLunaStevTimeZone(actor.ID, input.TimeZone); err != nil {
		writeAPIError(writer, http.StatusUnprocessableEntity, "invalid-time-zone", "Choose a valid IANA time zone.")
		return
	}
	writer.WriteHeader(http.StatusNoContent)
}

func (handler AdministrationHandler) authorize(writer http.ResponseWriter, request *http.Request, mutation bool) (accountView struct{ ID string }, ok bool) {
	setPrivateResponseHeaders(writer)
	if handler.Service == nil || handler.Auth == nil {
		writeAPIError(writer, http.StatusServiceUnavailable, "administration-unavailable", "Administration is unavailable.")
		return accountView, false
	}
	actor, authenticated := AuthenticatedAccount(*handler.Auth, request)
	if !authenticated {
		writeAPIError(writer, http.StatusUnauthorized, "not-authenticated", "Authentication is required.")
		return accountView, false
	}
	if !handler.Auth.Service.IsAdministrator(actor.ID) {
		writeAPIError(writer, http.StatusForbidden, "administrator-required", "Administrator access is required.")
		return accountView, false
	}
	if mutation && !sameOrigin(request) {
		writeAPIError(writer, http.StatusForbidden, "invalid-origin", "The request origin is not allowed.")
		return accountView, false
	}
	accountView.ID = actor.ID
	return accountView, true
}

func (handler AdministrationHandler) writeManagementError(writer http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, storage.ErrNotFound):
		writeAPIError(writer, http.StatusNotFound, "account-not-found", "The account was not found.")
	case errors.Is(err, admindomain.ErrForbidden), errors.Is(err, admindomain.ErrSelfAction):
		writeAPIError(writer, http.StatusForbidden, "management-forbidden", err.Error())
	case errors.Is(err, admindomain.ErrInvalidStatus):
		writeAPIError(writer, http.StatusUnprocessableEntity, "invalid-status", "Account status must be active or suspended.")
	default:
		writeAPIError(writer, http.StatusInternalServerError, "management-failed", "The management action could not be completed.")
	}
}
