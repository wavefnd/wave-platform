package handler

import (
	"net/http/httptest"
	"testing"
)

func TestDocumentLocaleSupportsCanonicalDocumentationLocales(t *testing.T) {
	for _, locale := range []string{"en", "ko", "ja", "zh", "es", "de", "ru", "id", "vi"} {
		request := httptest.NewRequest("GET", "/api/v1/documents?locale="+locale, nil)
		writer := httptest.NewRecorder()
		got, ok := documentLocale(writer, request)
		if !ok || got != locale {
			t.Fatalf("locale %q got=%q ok=%v status=%d", locale, got, ok, writer.Code)
		}
	}
}

func TestDocumentLocaleRejectsSeparateMalayLocale(t *testing.T) {
	request := httptest.NewRequest("GET", "/api/v1/documents?locale=ms", nil)
	writer := httptest.NewRecorder()
	if locale, ok := documentLocale(writer, request); ok || locale != "" || writer.Code != 400 {
		t.Fatalf("locale=%q ok=%v status=%d", locale, ok, writer.Code)
	}
}
