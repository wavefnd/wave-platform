package auth

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var ErrChallengeFailed = errors.New("human verification failed")

type TurnstileVerifier struct {
	SiteKey  string
	Secret   string
	Endpoint string
	Client   *http.Client
}

type turnstileResponse struct {
	Success    bool     `json:"success"`
	Action     string   `json:"action"`
	ErrorCodes []string `json:"error-codes"`
}

func (verifier TurnstileVerifier) Enabled() bool {
	return strings.TrimSpace(verifier.SiteKey) != "" && strings.TrimSpace(verifier.Secret) != ""
}

func (verifier TurnstileVerifier) Verify(ctx context.Context, token, remoteIP, action string) error {
	if !verifier.Enabled() {
		return nil
	}
	if strings.TrimSpace(token) == "" {
		return ErrChallengeFailed
	}
	endpoint := verifier.Endpoint
	if endpoint == "" {
		endpoint = "https://challenges.cloudflare.com/turnstile/v0/siteverify"
	}
	client := verifier.Client
	if client == nil {
		client = &http.Client{Timeout: 5 * time.Second}
	}
	form := url.Values{"secret": {verifier.Secret}, "response": {token}}
	if remoteIP != "" {
		form.Set("remoteip", remoteIP)
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return ErrChallengeFailed
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response, err := client.Do(request)
	if err != nil {
		return ErrChallengeFailed
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return ErrChallengeFailed
	}
	var result turnstileResponse
	if err := json.NewDecoder(response.Body).Decode(&result); err != nil || !result.Success {
		return ErrChallengeFailed
	}
	if result.Action != "" && action != "" && result.Action != action {
		return ErrChallengeFailed
	}
	return nil
}
