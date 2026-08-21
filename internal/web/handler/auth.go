package handler

import (
	"encoding/xml"
	"errors"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/wavefnd/wave-platform/internal/account"
	"github.com/wavefnd/wave-platform/internal/auth"
	"github.com/wavefnd/wave-platform/internal/identity"
	"github.com/wavefnd/wave-platform/internal/session"
	"github.com/wavefnd/wave-platform/internal/xmlcodec"
)

const SessionCookieName = "wave_session"

type AuthHandler struct {
	Service          *identity.Service
	MailDomain       string
	RegistrationOpen bool
	SecureCookies    bool
	Challenge        auth.TurnstileVerifier
}

type AuthConfigResponse struct {
	XMLName          xml.Name `xml:"https://wave-lang.dev/ns/platform/api/v1 auth-config"`
	MailDomain       string   `xml:"mail-domain"`
	RegistrationOpen bool     `xml:"registration-open"`
	TOTPConfigured   bool     `xml:"totp-configured"`
	TurnstileSiteKey string   `xml:"turnstile-site-key,omitempty"`
}

type RegistrationAddressResponse struct {
	XMLName        xml.Name `xml:"https://wave-lang.dev/ns/platform/api/v1 registration-address"`
	LocalPart      string   `xml:"local-part"`
	ChoiceRequired bool     `xml:"choice-required"`
}

type BeginRegistrationRequest struct {
	DisplayName   string `xml:"display-name"`
	Username      string `xml:"username"`
	RecoveryEmail string `xml:"recovery-email"`
	Challenge     string `xml:"challenge-token"`
}

type FinishEnrollmentRequest struct {
	Token string `xml:"token"`
	Code  string `xml:"code"`
}

type LoginRequest struct {
	Identifier string `xml:"identifier"`
	Code       string `xml:"code"`
	Challenge  string `xml:"challenge-token"`
}

type RecoveryRequestInput struct {
	Identifier string `xml:"identifier"`
	Challenge  string `xml:"challenge-token"`
}

type RecoveryTokenRequest struct {
	Token string `xml:"token"`
	Code  string `xml:"code,omitempty"`
}

type RecoveryEmailRequest struct {
	Email string `xml:"email"`
	Code  string `xml:"code"`
}

type RotationRequest struct {
	Code string `xml:"code"`
}

type EnrollmentResponse struct {
	XMLName   xml.Name  `xml:"https://wave-lang.dev/ns/platform/api/v1 totp-enrollment"`
	Token     string    `xml:"token"`
	Secret    string    `xml:"secret"`
	URI       string    `xml:"uri"`
	ExpiresAt time.Time `xml:"expires-at"`
}

type SecurityResponse struct {
	XMLName          xml.Name `xml:"https://wave-lang.dev/ns/platform/api/v1 account-security"`
	TOTPEnabled      bool     `xml:"totp-enabled"`
	RecoveryEmail    string   `xml:"recovery-email"`
	RecoveryVerified bool     `xml:"recovery-verified"`
}

type AccountResponse struct {
	XMLName          xml.Name  `xml:"https://wave-lang.dev/ns/platform/api/v1 account-session"`
	ID               string    `xml:"id"`
	Username         string    `xml:"username"`
	DisplayName      string    `xml:"display-name"`
	Email            string    `xml:"email"`
	TimeZone         string    `xml:"time-zone"`
	Administrator    bool      `xml:"administrator"`
	Owner            bool      `xml:"owner"`
	SourceMaintainer bool      `xml:"source-maintainer"`
	ExpiresAt        time.Time `xml:"expires-at"`
}

type ErrorResponse struct {
	XMLName xml.Name `xml:"https://wave-lang.dev/ns/platform/api/v1 error"`
	Code    string   `xml:"code"`
	Message string   `xml:"message"`
}

func (handler AuthHandler) Config(writer http.ResponseWriter, _ *http.Request) {
	setPrivateResponseHeaders(writer)
	_ = xmlcodec.Write(writer, http.StatusOK, AuthConfigResponse{MailDomain: handler.MailDomain,
		RegistrationOpen: handler.RegistrationOpen, TOTPConfigured: handler.Service.TOTPConfigured(),
		TurnstileSiteKey: handler.Challenge.SiteKey})
}

func (handler AuthHandler) RegistrationAddress(writer http.ResponseWriter, request *http.Request) {
	setPrivateResponseHeaders(writer)
	localPart, required, err := handler.Service.RegistrationAddress(request.URL.Query().Get("display-name"))
	if err != nil {
		handler.writeAuthError(writer, err)
		return
	}
	_ = xmlcodec.Write(writer, http.StatusOK, RegistrationAddressResponse{LocalPart: localPart, ChoiceRequired: required})
}

func (handler AuthHandler) BeginRegistration(writer http.ResponseWriter, request *http.Request) {
	setPrivateResponseHeaders(writer)
	if !handler.allowMutation(writer, request, "register") {
		return
	}
	var input BeginRegistrationRequest
	if xmlcodec.Decode(request.Body, &input) != nil {
		writeAPIError(writer, http.StatusBadRequest, "invalid-xml", "The registration request is not valid XML.")
		return
	}
	if err := handler.Challenge.Verify(request.Context(), input.Challenge, remoteIP(request), "register"); err != nil {
		writeAPIError(writer, http.StatusForbidden, "challenge-failed", "Human verification failed.")
		return
	}
	result, err := handler.Service.BeginTOTPRegistration(input.DisplayName, input.Username, input.RecoveryEmail)
	if err != nil {
		handler.writeAuthError(writer, err)
		return
	}
	handler.writeEnrollment(writer, result)
}

func (handler AuthHandler) FinishRegistration(writer http.ResponseWriter, request *http.Request) {
	setPrivateResponseHeaders(writer)
	if !handler.allowMutation(writer, request, "") {
		return
	}
	var input FinishEnrollmentRequest
	if xmlcodec.Decode(request.Body, &input) != nil {
		writeAPIError(writer, http.StatusBadRequest, "invalid-xml", "The enrollment request is not valid XML.")
		return
	}
	item, token, currentSession, err := handler.Service.CompleteTOTPRegistration(input.Token, input.Code, request.UserAgent())
	if err != nil {
		handler.writeAuthError(writer, err)
		return
	}
	handler.setCookie(writer, token, currentSession)
	handler.writeAccount(writer, http.StatusCreated, item, currentSession)
}

func (handler AuthHandler) Login(writer http.ResponseWriter, request *http.Request) {
	setPrivateResponseHeaders(writer)
	if !handler.allowMutation(writer, request, "login") {
		return
	}
	var input LoginRequest
	if xmlcodec.Decode(request.Body, &input) != nil {
		writeAPIError(writer, http.StatusBadRequest, "invalid-xml", "The login request is not valid XML.")
		return
	}
	if err := handler.Challenge.Verify(request.Context(), input.Challenge, remoteIP(request), "login"); err != nil {
		writeAPIError(writer, http.StatusForbidden, "challenge-failed", "Human verification failed.")
		return
	}
	item, token, currentSession, err := handler.Service.AuthenticateTOTP(input.Identifier, input.Code, request.UserAgent())
	if err != nil {
		writeAPIError(writer, http.StatusUnauthorized, "invalid-credentials", "The username, email address, or authenticator code is incorrect.")
		return
	}
	handler.setCookie(writer, token, currentSession)
	handler.writeAccount(writer, http.StatusOK, item, currentSession)
}

func (handler AuthHandler) RequestRecovery(writer http.ResponseWriter, request *http.Request) {
	setPrivateResponseHeaders(writer)
	if !handler.allowMutation(writer, request, "recovery") {
		return
	}
	var input RecoveryRequestInput
	if xmlcodec.Decode(request.Body, &input) != nil {
		writeAPIError(writer, http.StatusBadRequest, "invalid-xml", "The recovery request is not valid XML.")
		return
	}
	if err := handler.Challenge.Verify(request.Context(), input.Challenge, remoteIP(request), "recovery"); err != nil {
		writeAPIError(writer, http.StatusForbidden, "challenge-failed", "Human verification failed.")
		return
	}
	_ = handler.Service.RequestTOTPRecovery(input.Identifier)
	writer.WriteHeader(http.StatusNoContent)
}

func (handler AuthHandler) RecoveryEnrollment(writer http.ResponseWriter, request *http.Request) {
	setPrivateResponseHeaders(writer)
	if !handler.allowMutation(writer, request, "") {
		return
	}
	var input RecoveryTokenRequest
	if xmlcodec.Decode(request.Body, &input) != nil {
		writeAPIError(writer, http.StatusBadRequest, "invalid-xml", "The recovery token is not valid.")
		return
	}
	result, err := handler.Service.TOTPRecovery(input.Token)
	if err != nil {
		writeAPIError(writer, http.StatusGone, "recovery-expired", "This recovery link is invalid or expired.")
		return
	}
	handler.writeEnrollment(writer, result)
}

func (handler AuthHandler) FinishRecovery(writer http.ResponseWriter, request *http.Request) {
	setPrivateResponseHeaders(writer)
	if !handler.allowMutation(writer, request, "") {
		return
	}
	var input RecoveryTokenRequest
	if xmlcodec.Decode(request.Body, &input) != nil || handler.Service.CompleteTOTPRecovery(input.Token, input.Code) != nil {
		writeAPIError(writer, http.StatusUnauthorized, "invalid-code", "The authenticator code is incorrect or the link expired.")
		return
	}
	writer.WriteHeader(http.StatusNoContent)
}

func (handler AuthHandler) Security(writer http.ResponseWriter, request *http.Request) {
	setPrivateResponseHeaders(writer)
	actor, ok := AuthenticatedAccount(handler, request)
	if !ok {
		writeAPIError(writer, http.StatusUnauthorized, "not-authenticated", "Authentication is required.")
		return
	}
	factor, err := handler.Service.SecurityStatus(actor.ID)
	if err != nil {
		writeAPIError(writer, http.StatusConflict, "totp-not-configured", "TOTP is not configured for this account.")
		return
	}
	_ = xmlcodec.Write(writer, http.StatusOK, SecurityResponse{TOTPEnabled: true,
		RecoveryEmail: auth.MaskEmail(factor.RecoveryEmail), RecoveryVerified: factor.RecoveryVerified})
}

func (handler AuthHandler) BeginRotation(writer http.ResponseWriter, request *http.Request) {
	setPrivateResponseHeaders(writer)
	actor, ok := AuthenticatedAccount(handler, request)
	if !ok || !handler.allowMutation(writer, request, "") {
		if !ok {
			writeAPIError(writer, http.StatusUnauthorized, "not-authenticated", "Authentication is required.")
		}
		return
	}
	var input RotationRequest
	if xmlcodec.Decode(request.Body, &input) != nil {
		writeAPIError(writer, http.StatusBadRequest, "invalid-xml", "The request is not valid XML.")
		return
	}
	result, err := handler.Service.BeginTOTPRotation(actor.ID, input.Code)
	if err != nil {
		handler.writeAuthError(writer, err)
		return
	}
	handler.writeEnrollment(writer, result)
}

func (handler AuthHandler) FinishRotation(writer http.ResponseWriter, request *http.Request) {
	setPrivateResponseHeaders(writer)
	actor, ok := AuthenticatedAccount(handler, request)
	if !ok || !handler.allowMutation(writer, request, "") {
		if !ok {
			writeAPIError(writer, http.StatusUnauthorized, "not-authenticated", "Authentication is required.")
		}
		return
	}
	var input FinishEnrollmentRequest
	if xmlcodec.Decode(request.Body, &input) != nil || handler.Service.CompleteTOTPRotation(actor.ID, input.Token, input.Code) != nil {
		writeAPIError(writer, http.StatusUnauthorized, "invalid-code", "The authenticator code is incorrect.")
		return
	}
	handler.clearCookie(writer)
	writer.WriteHeader(http.StatusNoContent)
}

func (handler AuthHandler) ChangeRecoveryEmail(writer http.ResponseWriter, request *http.Request) {
	setPrivateResponseHeaders(writer)
	actor, ok := AuthenticatedAccount(handler, request)
	if !ok || !handler.allowMutation(writer, request, "") {
		if !ok {
			writeAPIError(writer, http.StatusUnauthorized, "not-authenticated", "Authentication is required.")
		}
		return
	}
	var input RecoveryEmailRequest
	if xmlcodec.Decode(request.Body, &input) != nil || handler.Service.ChangeRecoveryEmail(actor.ID, input.Code, input.Email) != nil {
		writeAPIError(writer, http.StatusUnauthorized, "invalid-code", "The authenticator code or email address is invalid.")
		return
	}
	writer.WriteHeader(http.StatusNoContent)
}

func (handler AuthHandler) VerifyRecoveryEmail(writer http.ResponseWriter, request *http.Request) {
	setPrivateResponseHeaders(writer)
	if !handler.allowMutation(writer, request, "") {
		return
	}
	var input RecoveryTokenRequest
	if xmlcodec.Decode(request.Body, &input) != nil || handler.Service.VerifyRecoveryEmail(input.Token) != nil {
		writeAPIError(writer, http.StatusGone, "verification-expired", "This verification link is invalid or expired.")
		return
	}
	writer.WriteHeader(http.StatusNoContent)
}

func (handler AuthHandler) Current(writer http.ResponseWriter, request *http.Request) {
	setPrivateResponseHeaders(writer)
	token, ok := sessionToken(request)
	if !ok {
		writeAPIError(writer, http.StatusUnauthorized, "not-authenticated", "Authentication is required.")
		return
	}
	item, currentSession, err := handler.Service.AccountForToken(token)
	if err != nil {
		handler.clearCookie(writer)
		writeAPIError(writer, http.StatusUnauthorized, "not-authenticated", "Authentication is required.")
		return
	}
	handler.writeAccount(writer, http.StatusOK, item, currentSession)
}

func (handler AuthHandler) Logout(writer http.ResponseWriter, request *http.Request) {
	setPrivateResponseHeaders(writer)
	if !handler.allowMutation(writer, request, "") {
		return
	}
	if token, ok := sessionToken(request); ok {
		_ = handler.Service.Revoke(token)
	}
	handler.clearCookie(writer)
	writer.WriteHeader(http.StatusNoContent)
}

func (handler AuthHandler) writeEnrollment(writer http.ResponseWriter, result auth.EnrollmentResult) {
	_ = xmlcodec.Write(writer, http.StatusOK, EnrollmentResponse{Token: result.Token, Secret: result.Secret,
		URI: result.URI, ExpiresAt: result.ExpiresAt})
}

func (handler AuthHandler) writeAuthError(writer http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, identity.ErrRegistrationClosed):
		writeAPIError(writer, http.StatusForbidden, "registration-closed", "Account registration is closed.")
	case errors.Is(err, account.ErrConflict):
		writeAPIError(writer, http.StatusConflict, "account-conflict", "That Wave address is already in use.")
	case errors.Is(err, auth.ErrTOTPNotConfigured):
		writeAPIError(writer, http.StatusServiceUnavailable, "totp-not-configured", "TOTP authentication is not configured on this server.")
	case errors.Is(err, identity.ErrAddressChoiceRequired):
		writeAPIError(writer, http.StatusConflict, "address-choice-required", err.Error())
	case errors.Is(err, identity.ErrInvalidRegistration):
		writeAPIError(writer, http.StatusUnprocessableEntity, "invalid-account", strings.TrimPrefix(err.Error(), identity.ErrInvalidRegistration.Error()+": "))
	default:
		writeAPIError(writer, http.StatusUnauthorized, "invalid-credentials", "The authentication request could not be completed.")
	}
}

func (handler AuthHandler) allowMutation(writer http.ResponseWriter, request *http.Request, _ string) bool {
	if sameOrigin(request) {
		return true
	}
	writeAPIError(writer, http.StatusForbidden, "invalid-origin", "The request origin is not allowed.")
	return false
}

func (handler AuthHandler) writeAccount(writer http.ResponseWriter, status int, item account.Account, currentSession session.Session) {
	_ = xmlcodec.Write(writer, status, AccountResponse{ID: item.ID, Username: item.Username, DisplayName: item.DisplayName,
		Email: item.Email, TimeZone: normalizedTimeZone(item.TimeZone), Administrator: handler.Service.IsAdministrator(item.ID), Owner: handler.Service.IsOwner(item.ID),
		SourceMaintainer: handler.Service.IsSourceMaintainer(item.ID),
		ExpiresAt:        currentSession.ExpiresAt})
}

func normalizedTimeZone(value string) string {
	if strings.TrimSpace(value) == "" {
		return "UTC"
	}
	return value
}

func (handler AuthHandler) setCookie(writer http.ResponseWriter, token string, currentSession session.Session) {
	http.SetCookie(writer, &http.Cookie{Name: SessionCookieName, Value: token, Path: "/", Expires: currentSession.ExpiresAt,
		MaxAge: int(time.Until(currentSession.ExpiresAt).Seconds()), HttpOnly: true, Secure: handler.SecureCookies, SameSite: http.SameSiteLaxMode})
}

func (handler AuthHandler) clearCookie(writer http.ResponseWriter) {
	http.SetCookie(writer, &http.Cookie{Name: SessionCookieName, Value: "", Path: "/", MaxAge: -1, Expires: time.Unix(1, 0),
		HttpOnly: true, Secure: handler.SecureCookies, SameSite: http.SameSiteLaxMode})
}

func sessionToken(request *http.Request) (string, bool) {
	cookie, err := request.Cookie(SessionCookieName)
	if err != nil || cookie == nil || cookie.Value == "" {
		return "", false
	}
	return cookie.Value, true
}

func sameOrigin(request *http.Request) bool {
	origin := strings.TrimSpace(request.Header.Get("Origin"))
	if origin == "" {
		return true
	}
	expectedScheme := "http"
	if request.TLS != nil || request.Header.Get("X-Forwarded-Proto") == "https" {
		expectedScheme = "https"
	}
	return origin == expectedScheme+"://"+request.Host
}

func remoteIP(request *http.Request) string {
	if forwarded := strings.TrimSpace(strings.Split(request.Header.Get("X-Forwarded-For"), ",")[0]); forwarded != "" {
		return forwarded
	}
	host, _, _ := net.SplitHostPort(request.RemoteAddr)
	return host
}

func setPrivateResponseHeaders(writer http.ResponseWriter) {
	writer.Header().Set("Cache-Control", "no-store")
	writer.Header().Set("Pragma", "no-cache")
}

func writeAPIError(writer http.ResponseWriter, status int, code, message string) {
	_ = xmlcodec.Write(writer, status, ErrorResponse{Code: code, Message: message})
}

func AuthenticatedAccount(handler AuthHandler, request *http.Request) (account.Account, bool) {
	token, ok := sessionToken(request)
	if !ok {
		return account.Account{}, false
	}
	item, _, err := handler.Service.AccountForToken(token)
	return item, err == nil
}
