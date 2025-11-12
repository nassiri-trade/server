package httpclient

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"

	"trading_server/internal/domain"
)

type ForexFactoryFeed struct {
	client  *resty.Client
	baseURL string
}

type rawEvent struct {
	Title    string `json:"title"`
	Country  string `json:"country"`
	Date     string `json:"date"`
	Impact   string `json:"impact"`
	Actual   string `json:"actual"`
	Forecast string `json:"forecast"`
	Previous string `json:"previous"`
	URL      string `json:"url"`
}

func NewForexFactoryFeed(baseURL string, opts ...func(*resty.Client)) (*ForexFactoryFeed, error) {
	if strings.TrimSpace(baseURL) == "" {
		return nil, fmt.Errorf("baseURL is required")
	}

	client := resty.New().
		SetHeader("Accept", "application/json").
		SetTimeout(15 * time.Second)

	for _, opt := range opts {
		opt(client)
	}

	return &ForexFactoryFeed{
		client:  client,
		baseURL: baseURL,
	}, nil
}

func (f *ForexFactoryFeed) FetchEvents(ctx context.Context) ([]domain.CalendarEvent, error) {
	var payload []rawEvent

	resp, err := f.client.R().
		SetContext(ctx).
		SetResult(&payload).
		Get(f.baseURL)
	if err != nil {
		return nil, fmt.Errorf("fetch events: %w", err)
	}

	if resp.StatusCode() >= 400 {
		return nil, fmt.Errorf("feed responded with status %d", resp.StatusCode())
	}

	events := make([]domain.CalendarEvent, 0, len(payload))
	for _, item := range payload {
		ts, err := time.Parse(time.RFC3339, item.Date)
		if err != nil {
			// Skip malformed records while allowing the rest to be processed.
			continue
		}

		// Convert from Eastern Time (or any timezone) to UTC
		tsUTC := ts.UTC()

		ev := domain.CalendarEvent{
			Date:      tsUTC,
			Currency:  strings.TrimSpace(item.Country),
			Impact:    strings.TrimSpace(item.Impact),
			Detail:    strings.TrimSpace(item.Title),
			Actual:    strings.TrimSpace(item.Actual),
			Forecast:  strings.TrimSpace(item.Forecast),
			Previous:  strings.TrimSpace(item.Previous),
			SourceURL: strings.TrimSpace(item.URL),
		}

		events = append(events, ev.WithHash())
	}

	return events, nil
}
