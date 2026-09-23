package auth

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (function roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return function(request)
}

func TestTurnstileVerifierChecksAction(t *testing.T) {
	client := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		_ = request.ParseForm()
		if request.FormValue("secret") != "secret" || request.FormValue("response") != "token" {
			t.Error("missing verification form")
		}
		return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(`{"success":true,"action":"login"}`))}, nil
	})}
	verifier := TurnstileVerifier{SiteKey: "site", Secret: "secret", Endpoint: "https://turnstile.test/verify", Client: client}
	if err := verifier.Verify(context.Background(), "token", "127.0.0.1", "login"); err != nil {
		t.Fatal(err)
	}
	if err := verifier.Verify(context.Background(), "token", "127.0.0.1", "register"); err == nil {
		t.Fatal("action mismatch should fail")
	}
	if err := verifier.Verify(context.Background(), "token", "127.0.0.1", ""); err != nil {
		t.Fatal(err)
	}
}

func TestTurnstileVerifierRejectsOmittedAction(t *testing.T) {
	client := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(`{"success":true}`))}, nil
	})}
	verifier := TurnstileVerifier{SiteKey: "site", Secret: "secret", Endpoint: "https://turnstile.test/verify", Client: client}
	if err := verifier.Verify(context.Background(), "token", "127.0.0.1", "login"); err == nil {
		t.Fatal("omitted action should fail when an action is expected")
	}
	if err := verifier.Verify(context.Background(), "token", "127.0.0.1", ""); err != nil {
		t.Fatal(err)
	}
}

func TestTurnstileDisabledDoesNotCallNetwork(t *testing.T) {
	if err := (TurnstileVerifier{}).Verify(context.Background(), "", "", "login"); err != nil {
		t.Fatal(err)
	}
}
