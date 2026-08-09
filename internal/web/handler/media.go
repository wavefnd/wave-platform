package handler

import (
	"encoding/xml"
	"errors"
	"net/http"
	"os"

	mediadomain "github.com/wavefnd/wave-platform/internal/media"
	"github.com/wavefnd/wave-platform/internal/mediapolicy"
	"github.com/wavefnd/wave-platform/internal/xmlcodec"
)

type MediaHandler struct {
	Service *mediadomain.Service
	Auth    *AuthHandler
}

type MediaUploadResponse struct {
	XMLName xml.Name `xml:"https://wave-lang.dev/ns/platform/api/v1 media-upload"`
	ID      string   `xml:"id"`
	URL     string   `xml:"url"`
	Width   int      `xml:"width"`
	Height  int      `xml:"height"`
	Bytes   int64    `xml:"bytes"`
}

func (handler MediaHandler) UploadLunaStevImage(writer http.ResponseWriter, request *http.Request) {
	setPrivateResponseHeaders(writer)
	if handler.Service == nil || handler.Auth == nil {
		writeAPIError(writer, http.StatusServiceUnavailable, "media-unavailable", "Image uploads are unavailable.")
		return
	}
	actor, authenticated := AuthenticatedAccount(*handler.Auth, request)
	if !authenticated {
		writeAPIError(writer, http.StatusUnauthorized, "not-authenticated", "Authentication is required.")
		return
	}
	if !handler.Auth.Service.IsOwner(actor.ID) {
		writeAPIError(writer, http.StatusForbidden, "owner-required", "Only the platform owner can upload LunaStev images.")
		return
	}
	if !sameOrigin(request) {
		writeAPIError(writer, http.StatusForbidden, "invalid-origin", "The request origin is not allowed.")
		return
	}

	request.Body = http.MaxBytesReader(writer, request.Body, mediadomain.MaxInputBytes+(1<<20))
	if err := request.ParseMultipartForm(1 << 20); err != nil {
		var limitError *http.MaxBytesError
		if errors.As(err, &limitError) {
			writeAPIError(writer, http.StatusRequestEntityTooLarge, "image-too-large", "The image must be 12 MiB or smaller.")
		} else {
			writeAPIError(writer, http.StatusBadRequest, "invalid-upload", "The image upload request is invalid.")
		}
		return
	}
	if request.MultipartForm != nil {
		defer request.MultipartForm.RemoveAll()
	}
	file, header, err := request.FormFile("image")
	if err != nil {
		writeAPIError(writer, http.StatusBadRequest, "image-required", "Choose an image to upload.")
		return
	}
	defer file.Close()
	if header.Size < 1 || header.Size > mediadomain.MaxInputBytes {
		writeAPIError(writer, http.StatusRequestEntityTooLarge, "image-too-large", "The image must be 12 MiB or smaller.")
		return
	}

	asset, err := handler.Service.Store(request.Context(), file, actor.ID)
	if err != nil {
		handler.writeMediaError(writer, err)
		return
	}
	_ = xmlcodec.Write(writer, http.StatusCreated, MediaUploadResponse{
		ID: asset.ID, URL: "/media/lunastev/" + asset.Filename,
		Width: asset.Width, Height: asset.Height, Bytes: asset.Bytes,
	})
}

func (handler MediaHandler) LunaStevImage(writer http.ResponseWriter, request *http.Request) {
	if handler.Service == nil {
		http.NotFound(writer, request)
		return
	}
	file, info, err := handler.Service.Open(request.PathValue("image"))
	if errors.Is(err, os.ErrNotExist) {
		http.NotFound(writer, request)
		return
	}
	if err != nil {
		http.Error(writer, "failed to read image", http.StatusInternalServerError)
		return
	}
	defer file.Close()
	writer.Header().Set("Content-Type", "image/webp")
	writer.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	writer.Header().Set("X-Content-Type-Options", "nosniff")
	http.ServeContent(writer, request, info.Name(), info.ModTime(), file)
}

func (handler MediaHandler) writeMediaError(writer http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, mediadomain.ErrUnavailable), errors.Is(err, mediapolicy.ErrUnavailable):
		writeAPIError(writer, http.StatusServiceUnavailable, "media-unavailable", "The Wave media policy is unavailable.")
	case errors.Is(err, mediapolicy.ErrInputTooLarge):
		writeAPIError(writer, http.StatusRequestEntityTooLarge, "image-too-large", "The image must be 12 MiB or smaller.")
	case errors.Is(err, mediapolicy.ErrInvalidDimensions), errors.Is(err, mediapolicy.ErrPixelLimit):
		writeAPIError(writer, http.StatusUnprocessableEntity, "image-dimensions", "The image dimensions are too large.")
	case errors.Is(err, mediapolicy.ErrUnsupportedFormat), errors.Is(err, mediadomain.ErrInvalidImage):
		writeAPIError(writer, http.StatusUnsupportedMediaType, "unsupported-image", "Upload a JPEG, PNG, or WebP image.")
	case errors.Is(err, mediadomain.ErrOutputTooLarge):
		writeAPIError(writer, http.StatusUnprocessableEntity, "converted-image-too-large", "The converted WebP image is too large.")
	default:
		writeAPIError(writer, http.StatusInternalServerError, "image-processing-failed", "The image could not be processed.")
	}
}
