package handler

import (
	"encoding/xml"
	"net/http"
	"strings"

	patchdomain "github.com/wavefnd/wave-platform/internal/patcharchive"
	"github.com/wavefnd/wave-platform/internal/xmlcodec"
)

type PatchesHandler struct{ Service *patchdomain.Service }

type PatchesResponse struct {
	XMLName xml.Name            `xml:"https://wave-lang.dev/ns/platform/api/v1 patches"`
	Address string              `xml:"address"`
	Items   []patchdomain.Patch `xml:"patch"`
}

func (handler PatchesHandler) List(writer http.ResponseWriter, request *http.Request) {
	if handler.Service == nil {
		writeAPIError(writer, http.StatusServiceUnavailable, "patches-unavailable", "The patch archive is unavailable.")
		return
	}
	query := strings.TrimSpace(request.URL.Query().Get("q"))
	if len([]rune(query)) > 120 {
		writeAPIError(writer, http.StatusBadRequest, "invalid-query", "The patch search is too long.")
		return
	}
	items, err := handler.Service.List(query, 100)
	if err != nil {
		writeAPIError(writer, http.StatusInternalServerError, "patches-unavailable", "The patch archive could not be loaded.")
		return
	}
	_ = xmlcodec.Write(writer, http.StatusOK, PatchesResponse{Address: handler.Service.Address(), Items: items})
}

func (handler PatchesHandler) Get(writer http.ResponseWriter, request *http.Request) {
	if handler.Service == nil {
		writeAPIError(writer, http.StatusServiceUnavailable, "patches-unavailable", "The patch archive is unavailable.")
		return
	}
	item, err := handler.Service.Get(request.PathValue("patch"))
	if patchdomain.IsNotFound(err) {
		writeAPIError(writer, http.StatusNotFound, "patch-not-found", "The patch was not found.")
		return
	}
	if err != nil {
		writeAPIError(writer, http.StatusInternalServerError, "patches-unavailable", "The patch could not be loaded.")
		return
	}
	_ = xmlcodec.Write(writer, http.StatusOK, item)
}
