package sponsor

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"sort"
	"sync"
	"time"
)

const (
	DefaultCollectiveURL = "https://opencollective.com/wave-lang"
	defaultEndpoint      = "https://api.opencollective.com/graphql/v2"
)

type Member struct {
	Name     string
	Profile  string
	ImageURL string
	Website  string
	Type     string
}

type Tier struct {
	Name     string
	Slug     string
	Amount   float64
	Currency string
	Interval string
	Members  []Member
}

type Collective struct {
	Name        string
	URL         string
	Tiers       []Tier
	RefreshedAt time.Time
}

type Service struct {
	endpoint string
	client   *http.Client
	now      func() time.Time
	mu       sync.Mutex
	cached   Collective
	expires  time.Time
}

func NewService() *Service {
	return &Service{endpoint: defaultEndpoint, client: &http.Client{Timeout: 8 * time.Second}, now: time.Now}
}

func NewServiceWithClient(endpoint string, client *http.Client) *Service {
	return &Service{endpoint: endpoint, client: client, now: time.Now}
}

func (service *Service) Collective(ctx context.Context) (Collective, error) {
	service.mu.Lock()
	defer service.mu.Unlock()
	now := service.now().UTC()
	if !service.expires.IsZero() && now.Before(service.expires) {
		return service.cached, nil
	}
	collective, err := service.fetch(ctx)
	if err != nil {
		if !service.cached.RefreshedAt.IsZero() {
			return service.cached, nil
		}
		return Collective{}, err
	}
	collective.RefreshedAt = now
	service.cached = collective
	service.expires = now.Add(time.Hour)
	return collective, nil
}

func (service *Service) fetch(ctx context.Context) (Collective, error) {
	const query = `query($slug:String!){account(slug:$slug){name slug ... on AccountWithContributions {tiers(limit:50){nodes{name slug amount{value currency} interval}}} members(role:BACKER,limit:100){nodes{isActive tier{name slug} account{name slug imageUrl website type}}}}}`
	payload, _ := json.Marshal(map[string]any{"query": query, "variables": map[string]string{"slug": "wave-lang"}})
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, service.endpoint, bytes.NewReader(payload))
	if err != nil {
		return Collective{}, err
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("User-Agent", "Wave-Platform/1 OpenCollective sponsor directory")
	response, err := service.client.Do(request)
	if err != nil {
		return Collective{}, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return Collective{}, errors.New("Open Collective request failed")
	}
	var document struct {
		Data struct {
			Account struct {
				Name  string `json:"name"`
				Tiers struct {
					Nodes []struct {
						Name     string `json:"name"`
						Slug     string `json:"slug"`
						Interval string `json:"interval"`
						Amount   struct {
							Value    float64 `json:"value"`
							Currency string  `json:"currency"`
						} `json:"amount"`
					} `json:"nodes"`
				} `json:"tiers"`
				Members struct {
					Nodes []struct {
						Active bool `json:"isActive"`
						Tier   struct {
							Slug string `json:"slug"`
						} `json:"tier"`
						Account struct {
							Name     string `json:"name"`
							Slug     string `json:"slug"`
							ImageURL string `json:"imageUrl"`
							Website  string `json:"website"`
							Type     string `json:"type"`
						} `json:"account"`
					} `json:"nodes"`
				} `json:"members"`
			} `json:"account"`
		} `json:"data"`
		Errors []json.RawMessage `json:"errors"`
	}
	if err := json.NewDecoder(response.Body).Decode(&document); err != nil || len(document.Errors) > 0 || document.Data.Account.Name == "" {
		return Collective{}, errors.New("Open Collective returned an invalid response")
	}
	tiers := make([]Tier, 0, len(document.Data.Account.Tiers.Nodes))
	bySlug := make(map[string]int, len(document.Data.Account.Tiers.Nodes))
	for _, source := range document.Data.Account.Tiers.Nodes {
		bySlug[source.Slug] = len(tiers)
		tiers = append(tiers, Tier{Name: source.Name, Slug: source.Slug, Amount: source.Amount.Value,
			Currency: source.Amount.Currency, Interval: source.Interval})
	}
	for _, source := range document.Data.Account.Members.Nodes {
		index, exists := bySlug[source.Tier.Slug]
		if !source.Active || !exists || source.Account.Name == "" {
			continue
		}
		tiers[index].Members = append(tiers[index].Members, Member{Name: source.Account.Name,
			Profile: "https://opencollective.com/" + source.Account.Slug, ImageURL: source.Account.ImageURL,
			Website: source.Account.Website, Type: source.Account.Type})
	}
	sort.SliceStable(tiers, func(left, right int) bool { return tiers[left].Amount > tiers[right].Amount })
	for index := range tiers {
		sort.SliceStable(tiers[index].Members, func(left, right int) bool {
			return tiers[index].Members[left].Name < tiers[index].Members[right].Name
		})
	}
	return Collective{Name: document.Data.Account.Name, URL: DefaultCollectiveURL, Tiers: tiers}, nil
}
