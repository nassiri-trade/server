package domain

import (
	"context"
	"time"
)

// EventFeed provides an abstraction to fetch calendar events from external sources.
type EventFeed interface {
	FetchEvents(ctx context.Context) ([]CalendarEvent, error)
}

// EventRepository persists calendar events.
type EventRepository interface {
	UpsertEvents(ctx context.Context, events []CalendarEvent) error
	ListEvents(ctx context.Context, limit int) ([]CalendarEvent, error)
}

type PositionRepository interface {
	UpsertPosition(ctx context.Context, position UserPosition) error
	ClosePosition(ctx context.Context, ticket int64, closeTime time.Time, profit float64) error
	ListOpenPositions(ctx context.Context, userID string) ([]UserPosition, error)
}

type TradeRepository interface {
	AddTrade(ctx context.Context, trade UserTrade) error
	ListTrades(ctx context.Context, userID string, limit int) ([]UserTrade, error)
}

type UserRepository interface {
	UpsertUser(ctx context.Context, user User) error
	GetUser(ctx context.Context, userID string) (User, error)
}
