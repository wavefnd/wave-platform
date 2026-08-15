package handler

import (
	"encoding/xml"
	"net/http"
	"time"

	"github.com/wavefnd/wave-platform/internal/sponsor"
	"github.com/wavefnd/wave-platform/internal/xmlcodec"
)

type SponsorsHandler struct{ Service *sponsor.Service }

type SponsorsResponse struct {
	XMLName     xml.Name          `xml:"https://wave-lang.dev/ns/platform/api/v1 sponsors"`
	Name        string            `xml:"name"`
	URL         string            `xml:"url"`
	RefreshedAt time.Time         `xml:"refreshed-at"`
	Tiers       []SponsorTierView `xml:"tiers>tier"`
}

type SponsorTierView struct {
	Name     string              `xml:"name"`
	Slug     string              `xml:"slug"`
	Amount   float64             `xml:"amount"`
	Currency string              `xml:"currency"`
	Interval string              `xml:"interval"`
	Members  []SponsorMemberView `xml:"members>member"`
}

type SponsorMemberView struct {
	Name     string  `xml:"name"`
	Profile  string  `xml:"profile"`
	ImageURL string  `xml:"image-url,omitempty"`
	Website  string  `xml:"website,omitempty"`
	Type     string  `xml:"type,omitempty"`
	Amount   float64 `xml:"amount,omitempty"`
	Currency string  `xml:"currency,omitempty"`
}

func (handler SponsorsHandler) List(writer http.ResponseWriter, request *http.Request) {
	collective, err := handler.Service.Collective(request.Context())
	if err != nil {
		writeAPIError(writer, http.StatusBadGateway, "sponsors-unavailable", "The sponsor directory is temporarily unavailable.")
		return
	}
	response := SponsorsResponse{Name: collective.Name, URL: collective.URL, RefreshedAt: collective.RefreshedAt}
	for _, tier := range collective.Tiers {
		view := SponsorTierView{Name: tier.Name, Slug: tier.Slug, Amount: tier.Amount, Currency: tier.Currency, Interval: tier.Interval}
		for _, member := range tier.Members {
			view.Members = append(view.Members, SponsorMemberView{Name: member.Name, Profile: member.Profile,
				ImageURL: member.ImageURL, Website: member.Website, Type: member.Type,
				Amount: member.Amount, Currency: member.Currency})
		}
		response.Tiers = append(response.Tiers, view)
	}
	writer.Header().Set("Cache-Control", "public, max-age=300, stale-if-error=86400")
	_ = xmlcodec.Write(writer, http.StatusOK, response)
}
