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
		ImageURL:   "/media/lunastev/image-1787312400000-0123456789abcdef0123456789abcdef.webp",
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
		deliveries[0].ImageURL != "https://wave.example/media/lunastev/image-1787312400000-0123456789abcdef0123456789abcdef.webp" ||
		!strings.Contains(receivedBody, "Wave v0.3.0") || !strings.Contains(receivedBody, "Release notes preview") || !strings.Contains(receivedBody, "https://wave.example/releases/") || !strings.Contains(receivedBody, `"image_url":"https://wave.example/media/lunastev/`) {
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

func TestAccountScopedEndpointsRejectPlatformOnlyPatchEvents(t *testing.T) {
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

	if contains(UserSupportedEvents(), EventPatchReceived) {
		t.Fatalf("account-supported events expose %q: %#v", EventPatchReceived, UserSupportedEvents())
	}
	if !contains(SupportedEvents(), EventPatchReceived) {
		t.Fatalf("platform-supported events omit %q: %#v", EventPatchReceived, SupportedEvents())
	}
	_, err = service.SaveEndpointScoped("account-a", EndpointInput{
		Name:    "Patch feed",
		Kind:    "generic",
		URL:     "https://hooks.example.test/patches",
		Events:  []string{EventPatchReceived},
		Enabled: true,
	}, false)
	if !errors.Is(err, ErrInvalidEndpoint) {
		t.Fatalf("account-scoped patch endpoint error = %v", err)
	}
	endpoints, listErr := service.EndpointsFor("account-a")
	if listErr != nil || len(endpoints) != 0 {
		t.Fatalf("account endpoints=%#v err=%v", endpoints, listErr)
	}
}

func TestPlatformOnlyPatchPublishDoesNotLeakToAccountOrLegacyScopes(t *testing.T) {
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

	platform, err := service.SaveEndpointScoped("admin", EndpointInput{
		Name:    "Maintainer patch feed",
		Kind:    "generic",
		URL:     "https://hooks.example.test/platform-patches",
		Events:  []string{EventPatchReceived},
		Enabled: true,
	}, true)
	if err != nil {
		t.Fatal(err)
	}
	storedPlatform, err := service.repository.Endpoint(platform.ID)
	if err != nil {
		t.Fatal(err)
	}
	if storedPlatform.Scope != "platform" {
		t.Fatalf("platform endpoint scope = %q", storedPlatform.Scope)
	}

	account, err := service.SaveEndpointScoped("account-a", EndpointInput{
		Name:    "Account feed",
		Kind:    "generic",
		URL:     "https://hooks.example.test/account",
		Events:  []string{EventCommunityPost},
		Enabled: true,
	}, false)
	if err != nil {
		t.Fatal(err)
	}
	storedAccount, err := service.repository.Endpoint(account.ID)
	if err != nil {
		t.Fatal(err)
	}
	storedAccount.Events = []string{EventPatchReceived}
	if err := service.repository.PutEndpoint(storedAccount); err != nil {
		t.Fatal(err)
	}
	legacy := storedAccount
	legacy.ID = "legacy-empty-scope"
	legacy.Scope = ""
	if err := service.repository.PutEndpoint(legacy); err != nil {
		t.Fatal(err)
	}

	if err := service.Publish(Event{Type: EventPatchReceived, Title: "[PATCH] parser fix", URL: "/mail/lists/patchs/patch/example"}); err != nil {
		t.Fatal(err)
	}
	deliveries, err := service.Deliveries(10)
	if err != nil {
		t.Fatal(err)
	}
	if len(deliveries) != 1 || deliveries[0].EndpointID != platform.ID || deliveries[0].EventType != EventPatchReceived {
		t.Fatalf("platform-only deliveries = %#v", deliveries)
	}
	accountDeliveries, err := service.DeliveriesFor("account-a", 10)
	if err != nil || len(accountDeliveries) != 0 {
		t.Fatalf("account deliveries=%#v err=%v", accountDeliveries, err)
	}
	accountEndpoints, err := service.EndpointsFor("account-a")
	if err != nil || len(accountEndpoints) != 1 || accountEndpoints[0].ID != account.ID {
		t.Fatalf("account endpoints=%#v err=%v", accountEndpoints, err)
	}
}

func TestAccountOperationsCannotTakeOverPlatformEndpoint(t *testing.T) {
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
	platform, err := service.SaveEndpointScoped("admin", EndpointInput{Name: "Platform feed", Kind: "generic",
		URL: "https://hooks.example.test/platform", Events: []string{EventCommunityPost}, Enabled: true}, true)
	if err != nil {
		t.Fatal(err)
	}

	_, err = service.SaveEndpointScoped("admin", EndpointInput{ID: platform.ID, Name: "Taken over", Kind: "generic",
		Events: []string{EventCommunityPost}, Enabled: true}, false)
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("account update of platform endpoint error = %v", err)
	}
	if err := service.DeleteEndpointScoped("admin", platform.ID, false); !errors.Is(err, ErrForbidden) {
		t.Fatalf("account delete of platform endpoint error = %v", err)
	}
	if _, err := service.TestEndpointScoped(context.Background(), "admin", platform.ID, false); !errors.Is(err, ErrForbidden) {
		t.Fatalf("account test of platform endpoint error = %v", err)
	}
	if _, err := service.repository.Endpoint(platform.ID); err != nil {
		t.Fatalf("platform endpoint was changed or deleted: %v", err)
	}
	legacy, err := service.repository.Endpoint(platform.ID)
	if err != nil {
		t.Fatal(err)
	}
	legacy.ID = "legacy-unscoped"
	legacy.Scope = ""
	if err := service.repository.PutEndpoint(legacy); err != nil {
		t.Fatal(err)
	}
	if _, err := service.TestEndpointScoped(context.Background(), "admin", legacy.ID, true); !errors.Is(err, ErrInvalidEndpoint) {
		t.Fatalf("legacy endpoint test error = %v", err)
	}
	if err := service.Publish(Event{Type: EventCommunityPost, Title: "Public discussion", URL: "/community/thread/example"}); err != nil {
		t.Fatal(err)
	}
	deliveries, err := service.Deliveries(10)
	if err != nil || len(deliveries) != 1 || deliveries[0].EndpointID != platform.ID {
		t.Fatalf("legacy unscoped endpoint delivered: %#v err=%v", deliveries, err)
	}
}

func TestMailingListEventsAreRestrictedToPlatformEndpoints(t *testing.T) {
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

	if contains(UserSupportedEvents(), EventMailingListPost) || !contains(SupportedEvents(), EventMailingListPost) {
		t.Fatalf("mailing-list event visibility: platform=%#v account=%#v", SupportedEvents(), UserSupportedEvents())
	}
	if _, err := service.SaveEndpointScoped("member", EndpointInput{Name: "Private list leak", Kind: "generic",
		URL: "https://hooks.example.test/account-list", Events: []string{EventMailingListPost}, Enabled: true}, false); !errors.Is(err, ErrInvalidEndpoint) {
		t.Fatalf("account-scoped mailing-list endpoint error = %v", err)
	}
	platform, err := service.SaveEndpointScoped("admin", EndpointInput{Name: "Public list feed", Kind: "generic",
		URL: "https://hooks.example.test/platform-list", Events: []string{EventMailingListPost}, Enabled: true}, true)
	if err != nil {
		t.Fatal(err)
	}
	account, err := service.SaveEndpointScoped("member", EndpointInput{Name: "Community feed", Kind: "generic",
		URL: "https://hooks.example.test/community", Events: []string{EventCommunityPost}, Enabled: true}, false)
	if err != nil {
		t.Fatal(err)
	}
	storedAccount, err := service.repository.Endpoint(account.ID)
	if err != nil {
		t.Fatal(err)
	}
	storedAccount.Events = []string{EventMailingListPost}
	if err := service.repository.PutEndpoint(storedAccount); err != nil {
		t.Fatal(err)
	}
	legacy := storedAccount
	legacy.ID = "legacy-list-endpoint"
	legacy.Scope = ""
	if err := service.repository.PutEndpoint(legacy); err != nil {
		t.Fatal(err)
	}

	if err := service.Publish(Event{Type: EventMailingListPost, Title: "Wave ABI", Summary: "Internal discussion preview",
		AuthorName: "Wave Member", ResourceID: "mailing-list/development/thread/thread-1",
		URL: "/mail/lists/development/thread/thread-1"}); err != nil {
		t.Fatal(err)
	}
	deliveries, err := service.Deliveries(10)
	if err != nil {
		t.Fatal(err)
	}
	if len(deliveries) != 1 || deliveries[0].EndpointID != platform.ID || deliveries[0].EventType != EventMailingListPost ||
		deliveries[0].Title != "Wave ABI" || deliveries[0].AuthorName != "Wave Member" ||
		deliveries[0].ResourceID != "mailing-list/development/thread/thread-1" ||
		deliveries[0].ResourceURL != "https://wave.example/mail/lists/development/thread/thread-1" {
		t.Fatalf("mailing-list deliveries = %#v", deliveries)
	}
	accountDeliveries, err := service.DeliveriesFor("member", 10)
	if err != nil || len(accountDeliveries) != 0 {
		t.Fatalf("account mailing-list deliveries=%#v err=%v", accountDeliveries, err)
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
		ImageURL: "https://wave-lang.dev/media/lunastev/image-1787312400000-0123456789abcdef0123456789abcdef.webp",
		URL:      "https://wave-lang.dev/lunastev/thread/example", OccurredAt: eventTime}
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
			Image struct {
				URL string `json:"url"`
			} `json:"image"`
			Footer struct {
				Text string `json:"text"`
			} `json:"footer"`
		} `json:"embeds"`
	}
	if err := json.Unmarshal(payload, &value); err != nil {
		t.Fatal(err)
	}
	if value.Username != "Wave" || value.Content != "" || len(value.AllowedMentions.Parse) != 0 || len(value.Embeds) != 1 || strings.Contains(string(payload), `"attachments"`) {
		t.Fatalf("discord payload = %s", payload)
	}
	embed := value.Embeds[0]
	if embed.Title != event.Title || embed.URL != event.URL || embed.Author.Name != "LunaStev" || embed.Image.URL != event.ImageURL || embed.Color != 0x6654F1 || embed.Timestamp != eventTime.Format(time.RFC3339) || embed.Footer.Text != "Wave · LunaStev post" {
		t.Fatalf("discord embed = %#v", embed)
	}
	if characters := []rune(embed.Description); len(characters) != 121 || characters[len(characters)-1] != '…' || strings.Contains(embed.Description, "\n") {
		t.Fatalf("discord description = %q (%d runes)", embed.Description, len(characters))
	}
}

func TestPublishDropsExternalAndMalformedDiscordImageURLs(t *testing.T) {
	database, err := storage.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	key := base64.StdEncoding.EncodeToString([]byte("0123456789abcdef0123456789abcdef"))
	service, err := NewService(database, key, "https://wave-lang.dev")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.SaveEndpoint("owner", EndpointInput{Name: "Discord", Kind: "generic", URL: "https://hooks.example.test/wave",
		Events: []string{EventFounderPost}, Enabled: true}); err != nil {
		t.Fatal(err)
	}
	for _, imageURL := range []string{
		"https://example.net/tracker.webp",
		"/media/lunastev/not-an-upload.webp",
		"/media/lunastev/image-1787312400000-0123456789abcdef0123456789abcdef.webp?tracking=1",
	} {
		if err := service.Publish(Event{Type: EventFounderPost, Title: "Founder note", ImageURL: imageURL, URL: "/lunastev/thread/example"}); err != nil {
			t.Fatal(err)
		}
	}
	deliveries, err := service.Deliveries(10)
	if err != nil || len(deliveries) != 3 {
		t.Fatalf("deliveries=%#v err=%v", deliveries, err)
	}
	for _, delivery := range deliveries {
		if delivery.ImageURL != "" {
			t.Fatalf("unsafe image URL was retained: %#v", delivery)
		}
	}
}
