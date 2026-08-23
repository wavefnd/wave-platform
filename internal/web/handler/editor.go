package handler

import (
	"encoding/xml"
	"errors"
	"net/http"

	editor "github.com/wavefnd/wave-platform/internal/editor"
	"github.com/wavefnd/wave-platform/internal/xmlcodec"
)

type EditorTransformRequest struct {
	XMLName xml.Name `xml:"https://wave-lang.dev/ns/platform/api/v1 editor-transform"`
	editor.Request
}

type EditorTransformResponse struct {
	XMLName xml.Name `xml:"https://wave-lang.dev/ns/platform/api/v1 editor-result"`
	editor.Result
}

type EditorHandler struct {
	Engine editor.Engine
	Auth   *AuthHandler
}

func (handler EditorHandler) Transform(writer http.ResponseWriter, request *http.Request) {
	setPrivateResponseHeaders(writer)
	writer.Header().Set("X-Robots-Tag", "noindex, nofollow, noarchive")
	if handler.Engine == nil || handler.Auth == nil {
		writeAPIError(writer, http.StatusServiceUnavailable, "editor-unavailable", "WaveEditor is unavailable.")
		return
	}
	if _, authenticated := AuthenticatedAccount(*handler.Auth, request); !authenticated {
		writeAPIError(writer, http.StatusUnauthorized, "not-authenticated", "Authentication is required.")
		return
	}
	if !sameOrigin(request) {
		writeAPIError(writer, http.StatusForbidden, "invalid-origin", "The request origin is not allowed.")
		return
	}
	request.Body = http.MaxBytesReader(writer, request.Body, 2<<20)
	var input EditorTransformRequest
	if xmlcodec.Decode(request.Body, &input) != nil {
		writeAPIError(writer, http.StatusBadRequest, "invalid-xml", "The WaveEditor request is not valid XML.")
		return
	}
	result, err := handler.Engine.Transform(request.Context(), input.Request)
	if errors.Is(err, editor.ErrInvalidRequest) {
		writeAPIError(writer, http.StatusUnprocessableEntity, "invalid-editor-request", err.Error())
		return
	}
	if err != nil {
		writeAPIError(writer, http.StatusInternalServerError, "editor-failed", "WaveEditor could not transform the document.")
		return
	}
	_ = xmlcodec.Write(writer, http.StatusOK, EditorTransformResponse{Result: result})
}
