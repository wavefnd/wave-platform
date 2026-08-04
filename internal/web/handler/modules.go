package handler

import (
	"encoding/xml"
	"net/http"

	"github.com/wavefnd/wave-platform/internal/xmlcodec"
)

type ModuleStatus struct {
	Name    string `xml:"name,attr"`
	Enabled bool   `xml:"enabled,attr"`
	Status  string `xml:"status,attr"`
}

type ModulesResponse struct {
	XMLName xml.Name       `xml:"https://wave-lang.dev/ns/platform/api/v1 modules"`
	Modules []ModuleStatus `xml:"module"`
}

type ModulesHandler struct {
	Modules []ModuleStatus
	Auth    *AuthHandler
}

func (handler ModulesHandler) Status(writer http.ResponseWriter, request *http.Request) {
	setPrivateResponseHeaders(writer)
	if handler.Auth == nil {
		writeAPIError(writer, http.StatusUnauthorized, "not-authenticated", "Authentication is required.")
		return
	}
	actor, ok := AuthenticatedAccount(*handler.Auth, request)
	if !ok {
		writeAPIError(writer, http.StatusUnauthorized, "not-authenticated", "Authentication is required.")
		return
	}
	if !handler.Auth.Service.IsAdministrator(actor.ID) {
		writeAPIError(writer, http.StatusForbidden, "administrator-required", "Administrator access is required.")
		return
	}
	response := ModulesResponse{
		Modules: handler.Modules,
	}

	if err := xmlcodec.Write(writer, http.StatusOK, response); err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
	}
}
