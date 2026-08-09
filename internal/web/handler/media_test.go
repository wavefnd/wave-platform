package handler

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/png"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	mediadomain "github.com/wavefnd/wave-platform/internal/media"
	"github.com/wavefnd/wave-platform/internal/mediapolicy"
	"github.com/wavefnd/wave-platform/internal/storage"
	"github.com/wavefnd/wave-platform/internal/testsupport"
)

type handlerMediaPolicy struct{}

func (handlerMediaPolicy) Plan(_ context.Context, width, height int, _ int64, _ mediapolicy.Format) (mediapolicy.Plan, error) {
	return mediapolicy.Plan{Width: width, Height: height}, nil
}

func TestLunaStevImageUploadRequiresOwnerAndServesWebP(t *testing.T) {
	database, err := storage.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = database.Close() })
	identities, err := testsupport.NewIdentity(database)
	if err != nil {
		t.Fatal(err)
	}
	owner, _, err := identities.BootstrapTOTPAdmin("Wave Owner", "wave-owner", "owner@example.net", testsupport.TOTPSecret)
	if err != nil {
		t.Fatal(err)
	}
	reader, err := testsupport.Register(identities, "Community Reader")
	if err != nil {
		t.Fatal(err)
	}
	_, ownerToken, _, err := testsupport.Authenticate(identities, owner.Email)
	if err != nil {
		t.Fatal(err)
	}
	_, readerToken, _, err := testsupport.Authenticate(identities, reader.Email)
	if err != nil {
		t.Fatal(err)
	}

	root := filepath.Join(t.TempDir(), "images")
	service, err := mediadomain.NewService(root, handlerMediaPolicy{}, nil)
	if err != nil {
		t.Fatal(err)
	}
	auth := AuthHandler{Service: identities}
	handler := MediaHandler{Service: service, Auth: &auth}

	readerResponse := httptest.NewRecorder()
	handler.UploadLunaStevImage(readerResponse, imageUploadRequest(t, readerToken))
	if readerResponse.Code != http.StatusForbidden {
		t.Fatalf("reader status=%d body=%s", readerResponse.Code, readerResponse.Body.String())
	}

	ownerResponse := httptest.NewRecorder()
	handler.UploadLunaStevImage(ownerResponse, imageUploadRequest(t, ownerToken))
	if ownerResponse.Code != http.StatusCreated {
		t.Fatalf("owner status=%d body=%s", ownerResponse.Code, ownerResponse.Body.String())
	}
	match := regexp.MustCompile(`/media/lunastev/(image-[0-9]+-[0-9a-f]{32}\.webp)`).FindStringSubmatch(ownerResponse.Body.String())
	if len(match) != 2 {
		t.Fatalf("upload body=%s", ownerResponse.Body.String())
	}

	imageRequest := httptest.NewRequest(http.MethodGet, "http://wave.test/media/lunastev/"+match[1], nil)
	imageRequest.SetPathValue("image", match[1])
	imageResponse := httptest.NewRecorder()
	handler.LunaStevImage(imageResponse, imageRequest)
	if imageResponse.Code != http.StatusOK || imageResponse.Header().Get("Content-Type") != "image/webp" {
		t.Fatalf("image status=%d type=%s", imageResponse.Code, imageResponse.Header().Get("Content-Type"))
	}
	if body := imageResponse.Body.Bytes(); len(body) < 12 || string(body[:4]) != "RIFF" || string(body[8:12]) != "WEBP" {
		t.Fatal("served content is not WebP")
	}
}

func imageUploadRequest(t *testing.T, token string) *http.Request {
	t.Helper()
	input := image.NewNRGBA(image.Rect(0, 0, 32, 16))
	for y := 0; y < 16; y++ {
		for x := 0; x < 32; x++ {
			input.Set(x, y, color.NRGBA{R: 102, G: 84, B: 241, A: 255})
		}
	}
	var encoded bytes.Buffer
	if err := png.Encode(&encoded, input); err != nil {
		t.Fatal(err)
	}
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("image", "wave.png")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := part.Write(encoded.Bytes()); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodPost, "http://wave.test/api/v1/media/lunastev/images", &body)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	request.Header.Set("Origin", "http://wave.test")
	request.Header.Set("Accept", "application/xml")
	request.AddCookie(&http.Cookie{Name: SessionCookieName, Value: strings.TrimSpace(token)})
	return request
}
