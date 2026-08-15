package sponsor

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

func TestCollectiveGroupsActiveMembersByTier(t *testing.T) {
	calls := 0
	client := &http.Client{Transport: roundTripFunc(func(_ *http.Request) (*http.Response, error) {
		calls++
		body := `{"data":{"account":{"name":"Wave Programming Language","tiers":{"nodes":[{"name":"Ripple","slug":"ripple","amount":{"value":3,"currency":"USD"},"interval":"month"}]},"members":{"nodes":[{"isActive":true,"tier":{"slug":"ripple"},"account":{"name":"Alice","slug":"alice","imageUrl":"","website":"","type":"USER"}},{"isActive":false,"tier":{"slug":"ripple"},"account":{"name":"Former","slug":"former","type":"USER"}}]},"activeContributors":{"nodes":[{"name":"Bob","slug":"bob"},{"name":"Private","slug":"private"}]},"contributors":{"nodes":[{"id":"alice","isBacker":true,"hasPublicProfile":true,"totalAmountContributed":{"value":9,"currency":"USD"},"account":{"name":"Alice","slug":"alice","type":"USER"}},{"id":"bob","isBacker":true,"hasPublicProfile":true,"totalAmountContributed":{"value":20,"currency":"USD"},"account":{"name":"Bob","slug":"bob","type":"USER"}},{"id":"former","isBacker":true,"hasPublicProfile":true,"totalAmountContributed":{"value":30,"currency":"USD"},"account":{"name":"Former","slug":"former","type":"USER"}},{"id":"private","isBacker":true,"hasPublicProfile":false,"totalAmountContributed":{"value":50,"currency":"USD"},"account":{"name":"Private","slug":"private","type":"USER"}}]}}}}`
		return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(body))}, nil
	})}
	service := NewServiceWithClient("https://opencollective.test/graphql", client)
	first, err := service.Collective(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	second, err := service.Collective(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if calls != 1 || len(first.Tiers) != 2 || len(first.Tiers[0].Members) != 1 || first.Tiers[0].Members[0].Name != "Alice" {
		t.Fatalf("calls=%d first=%+v second=%+v", calls, first, second)
	}
	oneTime := first.Tiers[1]
	if oneTime.Slug != "one-time" || len(oneTime.Members) != 1 || oneTime.Members[0].Name != "Bob" || oneTime.Members[0].Amount != 20 {
		t.Fatalf("one-time supporters=%+v", oneTime)
	}
}
