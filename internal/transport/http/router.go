package http

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	fiber "github.com/gofiber/fiber/v2"
	swagger "github.com/gofiber/swagger"

	"trading_server/internal/domain"
	"trading_server/internal/usecase"
)

type CalendarService interface {
	Sync(ctx context.Context) (int, error)
	List(ctx context.Context, opts domain.ListEventsOptions) ([]domain.CalendarEvent, error)
}

type TradingService interface {
	UpsertPosition(ctx context.Context, position domain.UserPosition) error
	RecordTrade(ctx context.Context, trade domain.UserTrade) error
	GenerateReport(ctx context.Context, userID string, limit int) (domain.UserPerformanceReport, error)
	ProcessTradingDataByLogin(ctx context.Context, login string, dataType string, positionData map[string]any, dealsData []map[string]any, commonData map[string]any, rawPayload []byte) error
	ListUsers(ctx context.Context, limit int) ([]domain.User, error)
}

type Router struct {
	app             *fiber.App
	calendarService CalendarService
	tradingService  TradingService
}

func New(calendar CalendarService, trading TradingService) *Router {
	app := fiber.New()

	r := &Router{
		app:             app,
		calendarService: calendar,
		tradingService:  trading,
	}

	api := app.Group("/api")
	v1 := api.Group("/v1")

	v1.Get("/events", r.listEvents)
	v1.Post("/events/sync", r.syncEvents)
	v1.Get("/users", r.listUsers)
	v1.Post("/users/:user_id/positions", r.upsertPosition)

	v1.Post("/users/:user_id/trades", r.recordTrade)
	v1.Get("/users/:user_id/report", r.getUserReport)

	// Unified endpoint that handles both Open positions and History deals
	v1.Post("/trading-data", r.handleTradingData)

	app.Get("/swagger/*", swagger.HandlerDefault)

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	return r
}

func (r *Router) App() *fiber.App {
	return r.app
}

// listEvents godoc
// @Summary List calendar events
// @Tags calendar
// @Produce json
// @Param limit query int false "Maximum number of events"
// @Param date_from query string false "Filter events with event_date greater than or equal to this ISO8601 timestamp"
// @Param date_to query string false "Filter events with event_date less than or equal to this ISO8601 timestamp"
// @Success 200 {array} domain.CalendarEvent
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /events [get]
func (r *Router) listEvents(c *fiber.Ctx) error {
	if r.calendarService == nil {
		return fiber.NewError(fiber.StatusServiceUnavailable, "calendar service unavailable")
	}

	limit := 100
	if v := c.Query("limit"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	opts := domain.ListEventsOptions{Limit: limit}

	if fromStr := c.Query("date_from"); fromStr != "" {
		from := parseTime(fromStr, time.RFC3339, "2006-01-02T15:04:05", "2006-01-02 15:04:05", "2006-01-02")
		if from.IsZero() {
			return fiber.NewError(fiber.StatusBadRequest, "invalid date_from")
		}
		fromCopy := from
		opts.From = &fromCopy
	}

	if toStr := c.Query("date_to"); toStr != "" {
		to := parseTime(toStr, time.RFC3339, "2006-01-02T15:04:05", "2006-01-02 15:04:05", "2006-01-02")
		if to.IsZero() {
			return fiber.NewError(fiber.StatusBadRequest, "invalid date_to")
		}
		toCopy := to
		opts.To = &toCopy
	}

	ctx, cancel := context.WithTimeout(userContext(c), 5*time.Second)
	defer cancel()

	events, err := r.calendarService.List(ctx, opts)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(events)
}

// syncEvents godoc
// @Summary Trigger a calendar sync
// @Tags calendar
// @Produce json
// @Success 200 {object} map[string]int
// @Failure 500 {object} map[string]string
// @Router /events/sync [post]
func (r *Router) syncEvents(c *fiber.Ctx) error {
	if r.calendarService == nil {
		return fiber.NewError(fiber.StatusServiceUnavailable, "calendar service unavailable")
	}
	ctx, cancel := context.WithTimeout(userContext(c), 30*time.Second)
	defer cancel()

	count, err := r.calendarService.Sync(ctx)
	if err != nil {
		if errors.Is(err, usecase.ErrNoEvents) {
			return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
				"synced": 0,
				"status": "no events available",
			})
		}
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(fiber.Map{
		"synced": count,
	})
}

func userContext(c *fiber.Ctx) context.Context {
	if ctx := c.UserContext(); ctx != nil {
		return ctx
	}
	return context.Background()
}

type positionPayload struct {
	Ticket      int64   `json:"ticket"`
	Symbol      string  `json:"symbol"`
	Side        string  `json:"side"`
	Volume      float64 `json:"volume"`
	EntryTime   string  `json:"entryTime"`
	Entry       float64 `json:"entry"`
	SL          float64 `json:"sl"`
	TP          float64 `json:"tp"`
	Profit      float64 `json:"profit"`
	RiskPercent float64 `json:"riskPercent"`
	TickValue   float64 `json:"tickValue"`
	TickSize    float64 `json:"tickSize"`
	Magic       int64   `json:"magic"`
	Reason      string  `json:"reason"`
	Comment     string  `json:"comment"`
}

type PositionRequest struct {
	Platform string          `json:"platform"`
	Name     string          `json:"name"`
	Login    string          `json:"login"`
	Server   string          `json:"server"`
	Broker   string          `json:"broker"`
	Time     string          `json:"time"`
	Balance  float64         `json:"balance"`
	Position positionPayload `json:"position"`
}

type TradeRequest struct {
	Ticket      int64   `json:"ticket"`
	PID         int64   `json:"pid"`
	Symbol      string  `json:"symbol"`
	Side        string  `json:"side"`
	EntryType   string  `json:"entryType"`
	Volume      float64 `json:"volume"`
	EntryTime   string  `json:"entryTime"`
	Entry       float64 `json:"entry"`
	SL          float64 `json:"sl"`
	TP          float64 `json:"tp"`
	Profit      float64 `json:"profit"`
	RiskPercent float64 `json:"riskPercent"`
	TickValue   float64 `json:"tickValue"`
	TickSize    float64 `json:"tickSize"`
	Magic       int64   `json:"magic"`
	Reason      string  `json:"reason"`
	Comment     string  `json:"comment"`
}

type TradingDataPosition struct {
	Ticket      int64   `json:"ticket"`
	Symbol      string  `json:"symbol"`
	Side        string  `json:"side"`
	Volume      float64 `json:"volume"`
	EntryTime   string  `json:"entryTime"`
	Entry       float64 `json:"entry"`
	SL          float64 `json:"sl"`
	TP          float64 `json:"tp"`
	Profit      float64 `json:"profit"`
	RiskPercent float64 `json:"riskPercent"`
	TickValue   float64 `json:"tickValue"`
	TickSize    float64 `json:"tickSize"`
	Magic       int64   `json:"magic"`
	Reason      string  `json:"reason"`
	Comment     string  `json:"comment"`
}

type TradingDataRequest struct {
	Platform string  `json:"platform"`
	Name     string  `json:"name"`
	Login    string  `json:"login"`
	Server   string  `json:"server"`
	Broker   string  `json:"broker"`
	Time     string  `json:"time"`
	TimeGMT  string  `json:"timeGMT"`
	Balance  float64 `json:"balance"`
	// Type indicates payload category. Use "History" for historical trades; any other value is treated as a live position snapshot.
	Type     string               `json:"type"`
	Position *TradingDataPosition `json:"position,omitempty"`
	Deals    []TradeRequest       `json:"deals,omitempty"`
}

// upsertPosition godoc
// @Summary Upsert a live position snapshot
// @Tags trading
// @Accept json
// @Produce json
// @Param user_id path string true "User ID"
// @Param request body PositionRequest true "Position payload"
// @Success 201 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /users/{user_id}/positions [post]
func (r *Router) upsertPosition(c *fiber.Ctx) error {
	if r.tradingService == nil {
		return fiber.NewError(fiber.StatusServiceUnavailable, "trading service unavailable")
	}

	userID := c.Params("user_id")
	if userID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "user_id required")
	}

	var payload map[string]any
	if err := c.BodyParser(&payload); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid payload")
	}
	raw, _ := json.Marshal(payload)

	position, err := decodePositionPayload(userID, payload, raw)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	ctx, cancel := context.WithTimeout(userContext(c), 10*time.Second)
	defer cancel()

	if err := r.tradingService.UpsertPosition(ctx, position); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"status": "ok"})
}

// recordTrade godoc
// @Summary Record a trade history entry
// @Tags trading
// @Accept json
// @Produce json
// @Param user_id path string true "User ID"
// @Param request body TradeRequest true "Trade payload"
// @Success 201 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /users/{user_id}/trades [post]
func (r *Router) recordTrade(c *fiber.Ctx) error {
	if r.tradingService == nil {
		return fiber.NewError(fiber.StatusServiceUnavailable, "trading service unavailable")
	}

	userID := c.Params("user_id")
	if userID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "user_id required")
	}

	var payload map[string]any
	if err := c.BodyParser(&payload); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid payload")
	}
	raw, _ := json.Marshal(payload)

	trade, err := decodeTradePayload(userID, payload, raw)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	ctx, cancel := context.WithTimeout(userContext(c), 10*time.Second)
	defer cancel()

	if err := r.tradingService.RecordTrade(ctx, trade); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"status": "ok"})
}

// getUserReport godoc
// @Summary Generate a performance report for a user
// @Tags trading
// @Produce json
// @Param user_id path string true "User ID"
// @Param limit query int false "Maximum history window"
// @Success 200 {object} domain.UserPerformanceReport
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /users/{user_id}/report [get]
func (r *Router) getUserReport(c *fiber.Ctx) error {
	if r.tradingService == nil {
		return fiber.NewError(fiber.StatusServiceUnavailable, "trading service unavailable")
	}

	userID := c.Params("user_id")
	if userID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "user_id required")
	}

	limit := 1000
	if v := c.Query("limit"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	ctx, cancel := context.WithTimeout(userContext(c), 30*time.Second)
	defer cancel()

	report, err := r.tradingService.GenerateReport(ctx, userID, limit)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(report)
}

// handleTradingData godoc
// @Summary Handle trading data (positions or history)
// @Tags trading
// @Accept json
// @Produce json
// @Param request body TradingDataRequest true "Trading data payload"
// @Success 201 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /trading-data [post]
func (r *Router) handleTradingData(c *fiber.Ctx) error {
	if r.tradingService == nil {
		return fiber.NewError(fiber.StatusServiceUnavailable, "trading service unavailable")
	}

	var payload map[string]any
	if err := c.BodyParser(&payload); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid payload")
	}
	raw, _ := json.Marshal(payload)

	// Extract login from payload
	login := toString(payload["login"])
	if login == "" {
		return fiber.NewError(fiber.StatusBadRequest, "login required")
	}

	// Extract type from payload
	dataType := toString(payload["type"])
	if dataType == "" {
		return fiber.NewError(fiber.StatusBadRequest, "type required")
	}
	normalizedType := strings.ToLower(strings.TrimSpace(dataType))

	// Extract position data for non-history types
	var positionData map[string]any
	if normalizedType != "history" {
		if pos, ok := payload["position"].(map[string]any); ok {
			positionData = pos
		} else {
			return fiber.NewError(fiber.StatusBadRequest, "position data required for live payloads")
		}
	}

	// Extract deals data if type is "History"
	var dealsData []map[string]any
	if normalizedType == "history" {
		if deals, ok := payload["deals"].([]any); ok {
			for _, deal := range deals {
				if dealMap, ok := deal.(map[string]any); ok {
					dealsData = append(dealsData, dealMap)
				}
			}
		}
		if len(dealsData) == 0 {
			return fiber.NewError(fiber.StatusBadRequest, "deals data required for history payloads")
		}
	}

	ctx, cancel := context.WithTimeout(userContext(c), 10*time.Second)
	defer cancel()

	if err := r.tradingService.ProcessTradingDataByLogin(ctx, login, normalizedType, positionData, dealsData, payload, raw); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"status": "ok"})
}

// listUsers godoc
// @Summary List users
// @Tags users
// @Produce json
// @Param limit query int false "Maximum number of users"
// @Success 200 {array} domain.User
// @Failure 500 {object} map[string]string
// @Router /users [get]
func (r *Router) listUsers(c *fiber.Ctx) error {
	if r.tradingService == nil {
		return fiber.NewError(fiber.StatusServiceUnavailable, "trading service unavailable")
	}

	limit := 100
	if v := c.Query("limit"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	ctx, cancel := context.WithTimeout(userContext(c), 5*time.Second)
	defer cancel()

	users, err := r.tradingService.ListUsers(ctx, limit)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(users)
}

func decodePositionPayload(userID string, payload map[string]any, raw []byte) (domain.UserPosition, error) {
	positionData, ok := payload["position"].(map[string]any)
	if !ok {
		return domain.UserPosition{}, errors.New("position payload missing")
	}

	ticket := toInt64(positionData["ticket"])
	if ticket == 0 {
		return domain.UserPosition{}, errors.New("ticket required")
	}

	entryTime := parseTime(positionData["entryTime"], time.RFC3339, "2006.01.02 15:04", "2006.01.02 15:04:05")

	return domain.UserPosition{
		UserID:       userID,
		Ticket:       ticket,
		Platform:     toString(payload["platform"]),
		AccountName:  toString(payload["name"]),
		AccountLogin: toString(payload["login"]),
		BrokerServer: toString(payload["server"]),
		BrokerName:   toString(payload["broker"]),
		Balance:      toFloat(payload["balance"]),
		Status:       domain.PositionStatusOpen,
		Symbol:       toString(positionData["symbol"]),
		Side:         domain.PositionSide(toString(positionData["side"])),
		Volume:       toFloat(positionData["volume"]),
		EntryTime:    entryTime,
		EntryPrice:   toFloat(positionData["entry"]),
		StopLoss:     toFloat(positionData["sl"]),
		TakeProfit:   toFloat(positionData["tp"]),
		Profit:       toFloat(positionData["profit"]),
		RiskPercent:  toFloat(positionData["riskPercent"]),
		TickValue:    toFloat(positionData["tickValue"]),
		TickSize:     toFloat(positionData["tickSize"]),
		Magic:        toInt64(positionData["magic"]),
		Reason:       toString(positionData["reason"]),
		Comment:      toString(positionData["comment"]),
		LastUpdate:   parseTime(payload["time"], "2006.01.02 15:04:05", "2006.01.02 15:04", time.RFC3339),
		RawPayload:   raw,
	}, nil
}

func decodeTradePayload(userID string, payload map[string]any, raw []byte) (domain.UserTrade, error) {
	ticket := toInt64(payload["ticket"])
	if ticket == 0 {
		return domain.UserTrade{}, errors.New("ticket required")
	}

	entryTime := parseTime(payload["entryTime"], "2006.01.02 15:04:05", "2006.01.02 15:04", time.RFC3339)

	var positionTicket *int64
	if pt := toInt64(payload["pid"]); pt != 0 {
		positionTicket = &pt
	}

	return domain.UserTrade{
		UserID:         userID,
		Ticket:         ticket,
		PositionTicket: positionTicket,
		Platform:       toString(payload["platform"]),
		Symbol:         toString(payload["symbol"]),
		Side:           domain.TradeSide(toString(payload["side"])),
		EntryType:      domain.TradeEntryType(toString(payload["entryType"])),
		Volume:         toFloat(payload["volume"]),
		EntryTime:      entryTime,
		EntryPrice:     toFloat(payload["entry"]),
		StopLoss:       toFloat(payload["sl"]),
		TakeProfit:     toFloat(payload["tp"]),
		Profit:         toFloat(payload["profit"]),
		RiskPercent:    toFloat(payload["riskPercent"]),
		TickValue:      toFloat(payload["tickValue"]),
		TickSize:       toFloat(payload["tickSize"]),
		Magic:          toInt64(payload["magic"]),
		Reason:         toString(payload["reason"]),
		Comment:        toString(payload["comment"]),
		RawPayload:     raw,
	}, nil
}

func toFloat(v any) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case float32:
		return float64(val)
	case int:
		return float64(val)
	case int64:
		return float64(val)
	case json.Number:
		f, _ := val.Float64()
		return f
	case string:
		f, _ := strconv.ParseFloat(val, 64)
		return f
	default:
		return 0
	}
}

func toInt64(v any) int64 {
	switch val := v.(type) {
	case float64:
		return int64(val)
	case float32:
		return int64(val)
	case int:
		return int64(val)
	case int64:
		return val
	case json.Number:
		i, _ := val.Int64()
		return i
	case string:
		i, _ := strconv.ParseInt(val, 10, 64)
		return i
	default:
		return 0
	}
}

func toString(v any) string {
	switch val := v.(type) {
	case string:
		return val
	case json.Number:
		return val.String()
	case fmt.Stringer:
		return val.String()
	default:
		return ""
	}
}

func parseTime(value any, layouts ...string) time.Time {
	raw := toString(value)
	if raw == "" {
		return time.Time{}
	}
	for _, layout := range layouts {
		if layout == "" {
			continue
		}
		if t, err := time.Parse(layout, raw); err == nil {
			return t
		}
		if strings.Contains(raw, ".") {
			normalized := strings.ReplaceAll(raw, ".", "-")
			normalizedLayout := strings.ReplaceAll(layout, ".", "-")
			if t, err := time.Parse(normalizedLayout, normalized); err == nil {
				return t
			}
		}
	}
	return time.Time{}
}
