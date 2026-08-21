package webhook

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/wavefnd/wave-platform/internal/storage"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (function roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return function(request)
}

func TestServiceEncryptsEndpointAndDeliversSignedGenericEvent(t *testing.T) {
	database, err := storage.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	key := base64.StdEncoding.EncodeToString([]byte("0123456789abcdef0123456789abcdef"))
	service, err := NewService(database, key, "https://wave.example")
	if err != nil {
		t.Fatal(err)
	}
	service.now = func() time.Time { return time.Date(2026, 8, 21, 4, 0, 0, 0, time.UTC) }

	created, err := service.SaveEndpoint("owner", EndpointInput{Name: "Release automation", Kind: "generic",
		URL: "https://hooks.example.test/wave", Events: []string{EventReleasePublished}, Enabled: true})
	if err != nil {
		t.Fatal(err)
	}
	if created.SigningSecret == "" || created.Destination != "hooks.example.test" {
		t.Fatalf("created = %#v", created)
	}
	stored, err := service.repository.Endpoint(created.ID)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(stored.EncryptedURL, "hooks.example") || stored.EncryptedURL == "" {
		t.Fatalf("URL was not encrypted: %q", stored.EncryptedURL)
	}

	var receivedBody string
	service.client = &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		body, readErr := io.ReadAll(request.Body)
		if readErr != nil {
			t.Fatal(readErr)
		}
		receivedBody = string(body)
		if request.Header.Get("X-Wave-Signature-256") == "" || request.Header.Get("X-Wave-Event") != EventReleasePublished {
			t.Fatalf("headers = %#v", request.Header)
		}
		return &http.Response{StatusCode: http.StatusNoContent, Body: io.NopCloser(strings.NewReader("")), Header: make(http.Header)}, nil
	})}
	if err := service.Publish(Event{Type: EventReleasePublished, Title: "Wave v0.3.0", Summary: "Release notes preview", AuthorName: "LunaStev",
		ResourceID: "blog/wave-v0.3.0", URL: "/releases/wave-v0.3.0"}); err != nil {
		t.Fatal(err)
	}
	if err := service.ProcessPending(context.Background()); err != nil {
		t.Fatal(err)
	}
	deliveries, err := service.Deliveries(10)
	if err != nil {
		t.Fatal(err)
	}
	if len(deliveries) != 1 || deliveries[0].Status != "delivered" || deliveries[0].Summary != "Release notes preview" || deliveries[0].AuthorName != "LunaStev" ||
		!strings.Contains(receivedBody, "Wave v0.3.0") || !strings.Contains(receivedBody, "Release notes preview") || !strings.Contains(receivedBody, "https://wave.example/releases/") {
		t.Fatalf("deliveries=%#v body=%q", deliveries, receivedBody)
	}
}

func TestServiceRejectsUnsafeAndImpersonatedDiscordDestinations(t *testing.T) {
	database, err := storage.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	key := base64.StdEncoding.EncodeToString([]byte("0123456789abcdef0123456789abcdef"))
	service, err := NewService(database, key, "https://wave.example")
	if err != nil {
		t.Fatal(err)
	}
	for _, input := range []EndpointInput{
		{Name: "Private", Kind: "generic", URL: "https://127.0.0.1/hook", Events: []string{EventPatchReceived}, Enabled: true},
		{Name: "Fake Discord", Kind: "discord", URL: "https://discord.example/api/webhooks/1/token", Events: []string{EventPatchReceived}, Enabled: true},
		{Name: "Insecure", Kind: "generic", URL: "http://hooks.example.test/hook", Events: []string{EventPatchReceived}, Enabled: true},
	} {
		if _, err := service.SaveEndpoint("owner", input); err == nil {
			t.Fatalf("unsafe input was accepted: %#v", input)
		}
	}
}

func TestSupportedEventsIncludeCommunityAndFounderPosts(t *testing.T) {
	events := SupportedEvents()
	if !contains(events, EventCommunityPost) || !contains(events, EventFounderPost) {
		t.Fatalf("supported events = %#v", events)
	}
}

func TestAccountScopedEndpointsEnforceOwnership(t *testing.T) {
	database, err := storage.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	key := base64.StdEncoding.EncodeToString([]byte("0123456789abcdef0123456789abcdef"))
	service, err := NewService(database, key, "https://wave.example")
	if err != nil {
		t.Fatal(err)
	}
	created, err := service.SaveEndpointScoped("account-a", EndpointInput{Name: "Community feed", Kind: "generic", URL: "https://hooks.example.test/community", Events: []string{EventCommunityPost}, Enabled: true}, false)
	if err != nil {
		t.Fatal(err)
	}
	other, err := service.EndpointsFor("account-b")
	if err != nil || len(other) != 0 {
		t.Fatalf("other endpoints=%#v err=%v", other, err)
	}
	_, err = service.SaveEndpointScoped("account-b", EndpointInput{ID: created.ID, Name: "Stolen", Kind: "generic", Events: []string{EventCommunityPost}, Enabled: true}, false)
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("update error = %v", err)
	}
	if err := service.DeleteEndpointScoped("account-b", created.ID, false); !errors.Is(err, ErrForbidden) {
		t.Fatalf("delete error = %v", err)
	}
	if _, err := service.TestEndpointScoped(context.Background(), "account-b", created.ID, false); !errors.Is(err, ErrForbidden) {
		t.Fatalf("test error = %v", err)
	}
}

func TestDiscordPayloadUsesLinkedEmbedAndTruncatedPreview(t *testing.T) {
	eventTime := time.Date(2026, 8, 21, 12, 30, 0, 0, time.UTC)
	event := Event{Type: EventFounderPost, Title: "Wave compiler work log",
		Summary: "First line.\n" + strings.Repeat("가", 180), AuthorName: "LunaStev",
		URL: "https://wave-lang.dev/lunastev/thread/example", OccurredAt: eventTime}
	payload, err := eventPayload("discord", event)
	if err != nil {
		t.Fatal(err)
	}
	var value struct {
		Username        string `json:"username"`
		Content         string `json:"content"`
		AllowedMentions struct {
			Parse []string `json:"parse"`
		} `json:"allowed_mentions"`
		Embeds []struct {
			Title       string `json:"title"`
			URL         string `json:"url"`
			Description string `json:"description"`
			Color       int    `json:"color"`
			Timestamp   string `json:"timestamp"`
			Author      struct {
				Name string `json:"name"`
			} `json:"author"`
			Footer struct {
				Text string `json:"text"`
			} `json:"footer"`
		} `json:"embeds"`
	}
	if err := json.Unmarshal(payload, &value); err != nil {
		t.Fatal(err)
	}
	if value.Username != "Wave" || value.Content != "" || len(value.AllowedMentions.Parse) != 0 || len(value.Embeds) != 1 {
		t.Fatalf("discord payload = %s", payload)
	}
	embed := value.Embeds[0]
	if embed.Title != event.Title || embed.URL != event.URL || embed.Author.Name != "LunaStev" || embed.Color != 0x6654F1 || embed.Timestamp != eventTime.Format(time.RFC3339) || embed.Footer.Text != "Wave · LunaStev post" {
		t.Fatalf("discord embed = %#v", embed)
	}
	if characters := []rune(embed.Description); len(characters) != 121 || characters[len(characters)-1] != '…' || strings.Contains(embed.Description, "\n") {
		t.Fatalf("discord description = %q (%d runes)", embed.Description, len(characters))
	}
}
