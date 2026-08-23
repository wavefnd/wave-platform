package handler

import (
	"encoding/xml"
	"errors"
	"fmt"
	"net/http"
	"strings"

	patchdomain "github.com/wavefnd/wave-platform/internal/patcharchive"
	"github.com/wavefnd/wave-platform/internal/xmlcodec"
)

type PatchesHandler struct {
	Service *patchdomain.Service
	Auth    *AuthHandler
}

type PatchesResponse struct {
	XMLName xml.Name            `xml:"https://wave-lang.dev/ns/platform/api/v1 patches"`
	Address string              `xml:"address"`
	Items   []patchdomain.Patch `xml:"patch"`
}

func (handler PatchesHandler) List(writer http.ResponseWriter, request *http.Request) {
	if _, ok := handler.authorizeMember(writer, request); !ok {
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
	if _, ok := handler.authorizeMember(writer, request); !ok {
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

func (handler PatchesHandler) authorizeMember(writer http.ResponseWriter, request *http.Request) (string, bool) {
	setPrivateResponseHeaders(writer)
	writer.Header().Set("X-Robots-Tag", "noindex, nofollow, noarchive")
	if handler.Service == nil || handler.Auth == nil {
		writeAPIError(writer, http.StatusServiceUnavailable, "patches-unavailable", "The patch archive is unavailable.")
		return "", false
	}
	actor, authenticated := AuthenticatedAccount(*handler.Auth, request)
	if !authenticated {
		writeAPIError(writer, http.StatusUnauthorized, "not-authenticated", "A Wave account is required to view the patch list.")
		return "", false
	}
	return actor.ID, true
}

func (handler PatchesHandler) Review(writer http.ResponseWriter, request *http.Request) {
	actorID, ok := handler.authorizeMaintainer(writer, request, true)
	if !ok {
		return
	}
	var input patchdomain.ReviewInput
	if xmlcodec.Decode(request.Body, &input) != nil {
		writeAPIError(writer, http.StatusBadRequest, "invalid-xml", "The patch review request is not valid XML.")
		return
	}
	item, err := handler.Service.UpdateReview(actorID, request.PathValue("patch"), input)
	if patchdomain.IsNotFound(err) {
		writeAPIError(writer, http.StatusNotFound, "patch-not-found", "The patch was not found.")
		return
	}
	if errors.Is(err, patchdomain.ErrInvalidReview) {
		writeAPIError(writer, http.StatusUnprocessableEntity, "invalid-review", err.Error())
		return
	}
	if errors.Is(err, patchdomain.ErrForbidden) {
		writeAPIError(writer, http.StatusForbidden, "source-maintainer-required", err.Error())
		return
	}
	if err != nil {
		writeAPIError(writer, http.StatusInternalServerError, "patch-review-failed", "The patch review could not be saved.")
		return
	}
	_ = xmlcodec.Write(writer, http.StatusOK, item)
}

func (handler PatchesHandler) AddReviewComment(writer http.ResponseWriter, request *http.Request) {
	actorID, ok := handler.authorizeMaintainer(writer, request, true)
	if !ok {
		return
	}
	var input patchdomain.ReviewCommentInput
	if xmlcodec.Decode(request.Body, &input) != nil {
		writeAPIError(writer, http.StatusBadRequest, "invalid-xml", "The patch review comment is not valid XML.")
		return
	}
	item, err := handler.Service.AddReviewComment(actorID, request.PathValue("patch"), input)
	handler.writeReviewComment(writer, item, err)
}

func (handler PatchesHandler) ResolveReviewComment(writer http.ResponseWriter, request *http.Request) {
	actorID, ok := handler.authorizeMaintainer(writer, request, true)
	if !ok {
		return
	}
	var input patchdomain.ReviewCommentResolutionInput
	if xmlcodec.Decode(request.Body, &input) != nil {
		writeAPIError(writer, http.StatusBadRequest, "invalid-xml", "The patch review comment resolution is not valid XML.")
		return
	}
	item, err := handler.Service.ResolveReviewComment(actorID, request.PathValue("patch"), request.PathValue("comment"), input.Resolved)
	handler.writeReviewComment(writer, item, err)
}

func (handler PatchesHandler) writeReviewComment(writer http.ResponseWriter, item patchdomain.ReviewComment, err error) {
	if patchdomain.IsNotFound(err) {
		writeAPIError(writer, http.StatusNotFound, "patch-review-comment-not-found", "The patch or review comment was not found.")
		return
	}
	if errors.Is(err, patchdomain.ErrInvalidComment) {
		writeAPIError(writer, http.StatusUnprocessableEntity, "invalid-review-comment", err.Error())
		return
	}
	if errors.Is(err, patchdomain.ErrForbidden) {
		writeAPIError(writer, http.StatusForbidden, "source-maintainer-required", err.Error())
		return
	}
	if err != nil {
		writeAPIError(writer, http.StatusInternalServerError, "patch-review-comment-failed", "The patch review comment could not be saved.")
		return
	}
	_ = xmlcodec.Write(writer, http.StatusOK, item)
}

func (handler PatchesHandler) Download(writer http.ResponseWriter, request *http.Request) {
	actorID, ok := handler.authorizeMaintainer(writer, request, false)
	if !ok {
		return
	}
	includeSeries := request.URL.Query().Get("series") == "1"
	content, filename, err := handler.Service.DownloadMbox(actorID, request.PathValue("patch"), includeSeries)
	if patchdomain.IsNotFound(err) {
		writeAPIError(writer, http.StatusNotFound, "patch-not-found", "The patch was not found.")
		return
	}
	if errors.Is(err, patchdomain.ErrForbidden) {
		writeAPIError(writer, http.StatusForbidden, "source-maintainer-required", err.Error())
		return
	}
	if err != nil {
		writeAPIError(writer, http.StatusInternalServerError, "patch-download-failed", "The patch download could not be prepared.")
		return
	}
	writer.Header().Set("Content-Type", "application/mbox")
	writer.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
	writer.Header().Set("Content-Length", fmt.Sprintf("%d", len(content)))
	writer.WriteHeader(http.StatusOK)
	_, _ = writer.Write(content)
}

func (handler PatchesHandler) authorizeMaintainer(writer http.ResponseWriter, request *http.Request, mutation bool) (string, bool) {
	setPrivateResponseHeaders(writer)
	writer.Header().Set("X-Robots-Tag", "noindex, nofollow, noarchive")
	if handler.Service == nil || handler.Auth == nil {
		writeAPIError(writer, http.StatusServiceUnavailable, "patches-unavailable", "Patch maintenance is unavailable.")
		return "", false
	}
	actor, authenticated := AuthenticatedAccount(*handler.Auth, request)
	if !authenticated {
		writeAPIError(writer, http.StatusUnauthorized, "not-authenticated", "Authentication is required.")
		return "", false
	}
	allowed, err := handler.Service.CanMaintain(actor.ID)
	if err != nil {
		writeAPIError(writer, http.StatusInternalServerError, "patches-unavailable", "Patch permissions could not be checked.")
		return "", false
	}
	if !allowed {
		writeAPIError(writer, http.StatusForbidden, "source-maintainer-required", "Source maintainer access is required.")
		return "", false
	}
	if mutation && !sameOrigin(request) {
		writeAPIError(writer, http.StatusForbidden, "invalid-origin", "The request origin is not allowed.")
		return "", false
	}
	return actor.ID, true
}
