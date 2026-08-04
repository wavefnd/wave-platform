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
		body := `{"data":{"account":{"name":"Wave Programming Language","tiers":{"nodes":[{"name":"Ripple","slug":"ripple","amount":{"value":3,"currency":"USD"},"interval":"month"}]},"members":{"nodes":[{"isActive":true,"tier":{"slug":"ripple"},"account":{"name":"Alice","slug":"alice","imageUrl":"","website":"","type":"USER"}},{"isActive":false,"tier":{"slug":"ripple"},"account":{"name":"Former","slug":"former","type":"USER"}}]}}}}`
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
	if calls != 1 || len(first.Tiers) != 1 || len(first.Tiers[0].Members) != 1 || first.Tiers[0].Members[0].Name != "Alice" {
		t.Fatalf("calls=%d first=%+v second=%+v", calls, first, second)
	}
}
