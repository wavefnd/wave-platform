package webhook

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/wavefnd/wave-platform/internal/audit"
	"github.com/wavefnd/wave-platform/internal/identifier"
	"github.com/wavefnd/wave-platform/internal/storage"
)

var (
	ErrInvalidEndpoint = errors.New("invalid webhook endpoint")
	ErrDeliveryFailed  = errors.New("webhook delivery failed")
	ErrForbidden       = errors.New("webhook access is forbidden")
)

var supportedEvents = map[string]bool{
	EventBlogPublished: true, EventCommunityPost: true, EventFounderPost: true,
	EventReleasePublished: true, EventPatchReceived: true,
}

var (
	lunaStevImagePathPattern = regexp.MustCompile(`^/media/lunastev/image-[0-9]+-[0-9a-f]{32}\.webp$`)
	markdownImagePattern     = regexp.MustCompile(`!\[([^\]\r\n]*)\]\([^\)\r\n]+\)`)
)

type Service struct {
	repository *Repository
	audit      *audit.Repository
	aead       cipher.AEAD
	client     *http.Client
	publicURL  string
	now        func() time.Time
}

func NewService(database *storage.Database, encodedKey, publicURL string) (*Service, error) {
	key, err := base64.StdEncoding.DecodeString(strings.TrimSpace(encodedKey))
	if err != nil || len(key) != 32 {
		return nil, errors.New("webhook encryption requires the configured 32-byte authentication key")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	return &Service{repository: NewRepository(database), audit: audit.NewRepository(database), aead: aead,
		client: safeHTTPClient(), publicURL: strings.TrimRight(publicURL, "/"), now: time.Now}, nil
}

func SupportedEvents() []string {
	items := make([]string, 0, len(supportedEvents))
	for item := range supportedEvents {
		items = append(items, item)
	}
	sort.Strings(items)
	return items
}

func (service *Service) Endpoints() ([]EndpointView, error) {
	items, err := service.repository.Endpoints()
	if err != nil {
		return nil, err
	}
	views := make([]EndpointView, 0, len(items))
	for _, item := range items {
		views = append(views, viewOf(item, ""))
	}
	return views, nil
}

func (service *Service) EndpointsFor(accountID string) ([]EndpointView, error) {
	items, err := service.repository.Endpoints()
	if err != nil {
		return nil, err
	}
	views := make([]EndpointView, 0)
	for _, item := range items {
		if item.OwnerAccountID == accountID {
			views = append(views, viewOf(item, ""))
		}
	}
	return views, nil
}

func (service *Service) Deliveries(limit int) ([]Delivery, error) {
	return service.repository.Deliveries(limit)
}

func (service *Service) DeliveriesFor(accountID string, limit int) ([]Delivery, error) {
	endpoints, err := service.repository.Endpoints()
	if err != nil {
		return nil, err
	}
	owned := map[string]bool{}
	for _, endpoint := range endpoints {
		if endpoint.OwnerAccountID == accountID {
			owned[endpoint.ID] = true
		}
	}
	items, err := service.repository.Deliveries(0)
	if err != nil {
		return nil, err
	}
	result := make([]Delivery, 0)
	for _, item := range items {
		if owned[item.EndpointID] {
			result = append(result, item)
			if limit > 0 && len(result) >= limit {
				break
			}
		}
	}
	return result, nil
}

func (service *Service) SaveEndpoint(actorID string, input EndpointInput) (EndpointView, error) {
	return service.SaveEndpointScoped(actorID, input, true)
}

func (service *Service) SaveEndpointScoped(actorID string, input EndpointInput, allowAll bool) (EndpointView, error) {
	input.ID, input.Name, input.Kind, input.URL = strings.TrimSpace(input.ID), strings.TrimSpace(input.Name), strings.ToLower(strings.TrimSpace(input.Kind)), strings.TrimSpace(input.URL)
	if len([]rune(input.Name)) < 2 || len([]rune(input.Name)) > 80 {
		return EndpointView{}, fmt.Errorf("%w: name must contain between 2 and 80 characters", ErrInvalidEndpoint)
	}
	if input.Kind != "generic" && input.Kind != "discord" {
		return EndpointView{}, fmt.Errorf("%w: kind must be generic or discord", ErrInvalidEndpoint)
	}
	events, err := normalizeEvents(input.Events)
	if err != nil {
		return EndpointView{}, err
	}
	now := service.now().UTC()
	item := Endpoint{ID: input.ID, CreatedAt: now}
	if item.ID != "" {
		item, err = service.repository.Endpoint(item.ID)
		if err != nil {
			return EndpointView{}, err
		}
		if !allowAll && item.OwnerAccountID != actorID {
			return EndpointView{}, ErrForbidden
		}
	} else {
		if !allowAll {
			owned, listErr := service.EndpointsFor(actorID)
			if listErr != nil {
				return EndpointView{}, listErr
			}
			if len(owned) >= 10 {
				return EndpointView{}, fmt.Errorf("%w: at most 10 endpoints are allowed per account", ErrInvalidEndpoint)
			}
		}
		item.ID, err = identifier.New("webhook")
		if err != nil {
			return EndpointView{}, err
		}
	}
	if item.OwnerAccountID == "" {
		item.OwnerAccountID = actorID
	}
	if input.URL != "" {
		destination, parseErr := validateURL(input.URL, input.Kind)
		if parseErr != nil {
			return EndpointView{}, parseErr
		}
		item.EncryptedURL, err = service.encrypt(input.URL)
		if err != nil {
			return EndpointView{}, err
		}
		item.Destination = destination
	} else if item.EncryptedURL == "" {
		return EndpointView{}, fmt.Errorf("%w: URL is required", ErrInvalidEndpoint)
	}
	secret := ""
	if item.EncryptedSigningSecret == "" || input.RotateSecret {
		secretBytes := make([]byte, 32)
		if _, err := rand.Read(secretBytes); err != nil {
			return EndpointView{}, err
		}
		secret = base64.RawURLEncoding.EncodeToString(secretBytes)
		item.EncryptedSigningSecret, err = service.encrypt(secret)
		if err != nil {
			return EndpointView{}, err
		}
	}
	item.Name, item.Kind, item.Events, item.Enabled, item.UpdatedAt = input.Name, input.Kind, events, input.Enabled, now
	if err := service.repository.PutEndpoint(item); err != nil {
		return EndpointView{}, err
	}
	if err := service.auditEvent(actorID, "webhook/"+item.ID, "admin.webhook.save", "success"); err != nil {
		return EndpointView{}, err
	}
	return viewOf(item, secret), nil
}

func (service *Service) DeleteEndpoint(actorID, id string) error {
	return service.DeleteEndpointScoped(actorID, id, true)
}

func (service *Service) DeleteEndpointScoped(actorID, id string, allowAll bool) error {
	item, err := service.repository.Endpoint(id)
	if err != nil {
		return err
	}
	if !allowAll && item.OwnerAccountID != actorID {
		return ErrForbidden
	}
	if err := service.repository.DeleteEndpoint(id); err != nil {
		return err
	}
	return service.auditEvent(actorID, "webhook/"+id, "admin.webhook.delete", "success")
}

func (service *Service) Publish(event Event) error {
	if !supportedEvents[event.Type] {
		return fmt.Errorf("unsupported webhook event %q", event.Type)
	}
	event.Title = truncateRunes(strings.Join(strings.Fields(event.Title), " "), 256)
	event.Summary = discordPreview(event.Summary, 1000)
	event.AuthorName = truncateRunes(strings.Join(strings.Fields(event.AuthorName), " "), 256)
	event.ImageURL = service.discordImageURL(event.ImageURL)
	if event.ID == "" {
		var err error
		event.ID, err = identifier.New("event")
		if err != nil {
			return err
		}
	}
	if event.OccurredAt.IsZero() {
		event.OccurredAt = service.now().UTC()
	}
	if strings.HasPrefix(event.URL, "/") {
		event.URL = service.publicURL + event.URL
	}
	endpoints, err := service.repository.Endpoints()
	if err != nil {
		return err
	}
	for _, endpoint := range endpoints {
		if !endpoint.Enabled || !contains(endpoint.Events, event.Type) {
			continue
		}
		id, idErr := identifier.New("webhook-delivery")
		if idErr != nil {
			return idErr
		}
		delivery := Delivery{ID: id, EndpointID: endpoint.ID, EventID: event.ID, EventType: event.Type, Title: event.Title,
			Summary: event.Summary, AuthorName: event.AuthorName, ImageURL: event.ImageURL,
			ResourceID: event.ResourceID, ResourceURL: event.URL, Status: "queued", NextAttemptAt: event.OccurredAt, CreatedAt: event.OccurredAt}
		if err := service.repository.PutDelivery(delivery); err != nil {
			return err
		}
	}
	return nil
}

func (service *Service) TestEndpoint(ctx context.Context, actorID, id string) (Delivery, error) {
	return service.TestEndpointScoped(ctx, actorID, id, true)
}

func (service *Service) TestEndpointScoped(ctx context.Context, actorID, id string, allowAll bool) (Delivery, error) {
	endpoint, endpointErr := service.repository.Endpoint(id)
	if endpointErr != nil {
		return Delivery{}, endpointErr
	}
	if !allowAll && endpoint.OwnerAccountID != actorID {
		return Delivery{}, ErrForbidden
	}
	now := service.now().UTC()
	eventID, err := identifier.New("event")
	if err != nil {
		return Delivery{}, err
	}
	deliveryID, err := identifier.New("webhook-delivery")
	if err != nil {
		return Delivery{}, err
	}
	delivery := Delivery{ID: deliveryID, EndpointID: id, EventID: eventID, EventType: "webhook.test", Title: "Wave webhook test",
		Summary: "Your Discord embed and webhook delivery are configured correctly.", AuthorName: "Wave Platform",
		ResourceID: "platform/webhooks", ResourceURL: service.publicURL + "/admin/webhooks", Status: "queued", CreatedAt: now, NextAttemptAt: now}
	if err := service.repository.PutDelivery(delivery); err != nil {
		return Delivery{}, err
	}
	delivery, err = service.deliver(ctx, delivery)
	result := "success"
	if err != nil {
		result = "failed"
	}
	_ = service.auditEvent(actorID, "webhook/"+id, "admin.webhook.test", result)
	return delivery, err
}

func (service *Service) Run(ctx context.Context) {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()
	for {
		_ = service.ProcessPending(ctx)
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func (service *Service) ProcessPending(ctx context.Context) error {
	items, err := service.repository.Pending(service.now().UTC(), 20)
	if err != nil {
		return err
	}
	for _, item := range items {
		_, _ = service.deliver(ctx, item)
	}
	return nil
}

func (service *Service) deliver(ctx context.Context, delivery Delivery) (Delivery, error) {
	endpoint, err := service.repository.Endpoint(delivery.EndpointID)
	if err != nil {
		return delivery, err
	}
	destination, err := service.decrypt(endpoint.EncryptedURL)
	if err != nil {
		return delivery, err
	}
	secret, err := service.decrypt(endpoint.EncryptedSigningSecret)
	if err != nil {
		return delivery, err
	}
	event := Event{ID: delivery.EventID, Type: delivery.EventType, Title: delivery.Title, Summary: delivery.Summary,
		AuthorName: delivery.AuthorName, ImageURL: delivery.ImageURL, ResourceID: delivery.ResourceID, URL: delivery.ResourceURL, OccurredAt: delivery.CreatedAt}
	payload, err := eventPayload(endpoint.Kind, event)
	if err != nil {
		return delivery, err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, destination, bytes.NewReader(payload))
	if err != nil {
		return delivery, err
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("User-Agent", "Wave-Platform-Webhook/1.0")
	request.Header.Set("X-Wave-Event", delivery.EventType)
	request.Header.Set("X-Wave-Delivery", delivery.ID)
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write(payload)
	request.Header.Set("X-Wave-Signature-256", "sha256="+hex.EncodeToString(mac.Sum(nil)))
	now := service.now().UTC()
	delivery.Attempts++
	delivery.LastAttemptAt = now
	response, requestErr := service.client.Do(request)
	if requestErr == nil {
		defer response.Body.Close()
		_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, 64<<10))
		delivery.HTTPStatus = response.StatusCode
	}
	if requestErr == nil && response.StatusCode >= 200 && response.StatusCode < 300 {
		delivery.Status, delivery.CompletedAt, delivery.LastError = "delivered", now, ""
		_ = service.repository.PutDelivery(delivery)
		return delivery, nil
	}
	if requestErr != nil {
		delivery.LastError = requestErr.Error()
	} else {
		delivery.LastError = fmt.Sprintf("remote endpoint returned HTTP %d", response.StatusCode)
	}
	if delivery.Attempts >= 5 {
		delivery.Status = "failed"
	} else {
		delivery.Status = "deferred"
		delivery.NextAttemptAt = now.Add(time.Duration(1<<min(delivery.Attempts, 6)) * time.Minute)
	}
	_ = service.repository.PutDelivery(delivery)
	return delivery, fmt.Errorf("%w: %s", ErrDeliveryFailed, delivery.LastError)
}

func eventPayload(kind string, event Event) ([]byte, error) {
	if kind == "discord" {
		embed := map[string]any{
			"title":       truncateRunes(strings.TrimSpace(event.Title), 256),
			"url":         event.URL,
			"color":       0x6654F1,
			"timestamp":   event.OccurredAt.UTC().Format(time.RFC3339),
			"footer":      map[string]string{"text": "Wave · " + discordEventLabel(event.Type)},
			"description": discordPreview(event.Summary, 120),
		}
		if strings.TrimSpace(event.AuthorName) != "" {
			embed["author"] = map[string]string{"name": truncateRunes(strings.TrimSpace(event.AuthorName), 256)}
		}
		if event.ImageURL != "" {
			embed["image"] = map[string]string{"url": event.ImageURL}
		}
		if embed["description"] == "" {
			delete(embed, "description")
		}
		return json.Marshal(map[string]any{
			"username":         "Wave",
			"allowed_mentions": map[string]any{"parse": []string{}},
			"embeds":           []any{embed},
		})
	}
	return json.Marshal(event)
}

func discordPreview(value string, limit int) string {
	value = markdownImagePattern.ReplaceAllString(value, "$1")
	value = strings.NewReplacer("\r", " ", "\n", " ", "\t", " ").Replace(value)
	value = strings.Join(strings.Fields(value), " ")
	return truncateRunes(value, limit)
}

func (service *Service) discordImageURL(value string) string {
	value = strings.TrimSpace(value)
	if !lunaStevImagePathPattern.MatchString(value) || service.publicURL == "" {
		return ""
	}
	return service.publicURL + value
}

func truncateRunes(value string, limit int) string {
	characters := []rune(strings.TrimSpace(value))
	if limit > 0 && len(characters) > limit {
		return strings.TrimSpace(string(characters[:limit])) + "…"
	}
	return string(characters)
}

func discordEventLabel(eventType string) string {
	switch eventType {
	case EventCommunityPost:
		return "Community post"
	case EventFounderPost:
		return "LunaStev post"
	case EventBlogPublished:
		return "Blog post"
	case EventReleasePublished:
		return "Release"
	case EventPatchReceived:
		return "Git patch"
	case "webhook.test":
		return "Webhook test"
	default:
		return eventType
	}
}

func normalizeEvents(values []string) ([]string, error) {
	seen := map[string]bool{}
	result := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.ToLower(strings.TrimSpace(value))
		if !supportedEvents[value] {
			return nil, fmt.Errorf("%w: unsupported event %q", ErrInvalidEndpoint, value)
		}
		if !seen[value] {
			seen[value] = true
			result = append(result, value)
		}
	}
	if len(result) == 0 {
		return nil, fmt.Errorf("%w: select at least one event", ErrInvalidEndpoint)
	}
	sort.Strings(result)
	return result, nil
}

func validateURL(value, kind string) (string, error) {
	parsed, err := url.Parse(value)
	if err != nil || parsed.Scheme != "https" || parsed.Hostname() == "" || parsed.User != nil || parsed.Fragment != "" {
		return "", fmt.Errorf("%w: an HTTPS URL without credentials or fragments is required", ErrInvalidEndpoint)
	}
	host := strings.ToLower(parsed.Hostname())
	if host == "localhost" || strings.HasSuffix(host, ".localhost") {
		return "", fmt.Errorf("%w: local destinations are not allowed", ErrInvalidEndpoint)
	}
	if ip := net.ParseIP(host); ip != nil && !publicIP(ip) {
		return "", fmt.Errorf("%w: private network destinations are not allowed", ErrInvalidEndpoint)
	}
	if kind == "discord" && !((host == "discord.com" || host == "discordapp.com") && strings.HasPrefix(parsed.Path, "/api/webhooks/")) {
		return "", fmt.Errorf("%w: Discord endpoints must use an official Discord webhook URL", ErrInvalidEndpoint)
	}
	return host, nil
}

func safeHTTPClient() *http.Client {
	dialer := &net.Dialer{Timeout: 5 * time.Second, KeepAlive: 30 * time.Second}
	transport := &http.Transport{TLSHandshakeTimeout: 5 * time.Second, ResponseHeaderTimeout: 8 * time.Second}
	transport.DialContext = func(ctx context.Context, network, address string) (net.Conn, error) {
		host, port, err := net.SplitHostPort(address)
		if err != nil {
			return nil, err
		}
		ips, err := net.DefaultResolver.LookupIPAddr(ctx, host)
		if err != nil {
			return nil, err
		}
		for _, item := range ips {
			if publicIP(item.IP) {
				return dialer.DialContext(ctx, network, net.JoinHostPort(item.IP.String(), port))
			}
		}
		return nil, errors.New("webhook destination resolved only to private or reserved addresses")
	}
	return &http.Client{Transport: transport, Timeout: 12 * time.Second, CheckRedirect: func(_ *http.Request, _ []*http.Request) error { return errors.New("webhook redirects are not allowed") }}
}

func publicIP(ip net.IP) bool {
	return ip != nil && !ip.IsLoopback() && !ip.IsPrivate() && !ip.IsUnspecified() && !ip.IsMulticast() && !ip.IsLinkLocalUnicast() && !ip.IsLinkLocalMulticast()
}
func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
func min(left, right int) int {
	if left < right {
		return left
	}
	return right
}

func (service *Service) encrypt(value string) (string, error) {
	nonce := make([]byte, service.aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(service.aead.Seal(nonce, nonce, []byte(value), nil)), nil
}
func (service *Service) decrypt(value string) (string, error) {
	data, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil || len(data) < service.aead.NonceSize() {
		return "", errors.New("invalid encrypted webhook secret")
	}
	plain, err := service.aead.Open(nil, data[:service.aead.NonceSize()], data[service.aead.NonceSize():], nil)
	return string(plain), err
}
func viewOf(item Endpoint, secret string) EndpointView {
	return EndpointView{ID: item.ID, OwnerAccountID: item.OwnerAccountID, Name: item.Name, Kind: item.Kind, Events: item.Events, Destination: item.Destination, Enabled: item.Enabled, CreatedAt: item.CreatedAt, UpdatedAt: item.UpdatedAt, SigningSecret: secret}
}
func (service *Service) auditEvent(actorID, resourceID, action, result string) error {
	id, err := identifier.New("audit")
	if err != nil {
		return err
	}
	return service.audit.Append(audit.Event{ID: id, ActorID: "account/" + actorID, ResourceID: resourceID, Action: action, Result: result, OccurredAt: service.now().UTC()})
}
