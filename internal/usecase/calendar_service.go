package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/rs/zerolog/log"

	"trading_server/internal/domain"
)

var ErrNoEvents = errors.New("no events fetched")

type CalendarService struct {
	feed domain.EventFeed
	repo domain.EventRepository
}

func NewCalendarService(feed domain.EventFeed, repo domain.EventRepository) (*CalendarService, error) {
	if feed == nil {
		return nil, fmt.Errorf("feed is required")
	}
	if repo == nil {
		return nil, fmt.Errorf("repository is required")
	}

	return &CalendarService{
		feed: feed,
		repo: repo,
	}, nil
}

func (s *CalendarService) Sync(ctx context.Context) (int, error) {
	log.Debug().Msg("fetching events from feed")
	events, err := s.feed.FetchEvents(ctx)
	if err != nil {
		log.Error().Err(err).Msg("failed to fetch events from feed")
		return 0, err
	}
	if len(events) == 0 {
		log.Warn().Msg("no events fetched from feed")
		return 0, ErrNoEvents
	}

	log.Debug().Int("count", len(events)).Msg("events fetched from feed")

	unique := make(map[string]domain.CalendarEvent, len(events))
	for _, ev := range events {
		hashed := ev.WithHash()
		unique[hashed.Hash] = hashed
	}

	collated := make([]domain.CalendarEvent, 0, len(unique))
	for _, ev := range unique {
		collated = append(collated, ev)
	}

	log.Debug().Int("unique_count", len(collated)).Msg("upserting events to database")
	if err := s.repo.UpsertEvents(ctx, collated); err != nil {
		log.Error().Err(err).Int("count", len(collated)).Msg("failed to upsert events")
		return 0, err
	}

	log.Info().Int("count", len(collated)).Msg("events synced successfully")
	return len(collated), nil
}

func (s *CalendarService) List(ctx context.Context, opts domain.ListEventsOptions) ([]domain.CalendarEvent, error) {
	if opts.Limit <= 0 {
		opts.Limit = 100
	}

	return s.repo.ListEvents(ctx, opts)
}
