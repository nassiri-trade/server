package repository

import (
	"time"

	"gorm.io/datatypes"

	"trading_server/internal/domain"
)

type CalendarEventModel struct {
	ID        int64     `gorm:"column:id"`
	Hash      string    `gorm:"column:event_hash;uniqueIndex;not null"`
	Date      time.Time `gorm:"column:event_date;not null"`
	Currency  string    `gorm:"column:currency;not null"`
	Impact    string    `gorm:"column:impact;not null"`
	Detail    string    `gorm:"column:detail;not null"`
	Actual    string    `gorm:"column:actual"`
	Forecast  string    `gorm:"column:forecast"`
	Previous  string    `gorm:"column:previous"`
	SourceURL string    `gorm:"column:source_url"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (CalendarEventModel) TableName() string {
	return "calendar_events"
}

func toCalendarEventModel(ev domain.CalendarEvent) CalendarEventModel {
	if ev.Hash == "" {
		ev = ev.WithHash()
	}
	return CalendarEventModel{
		Hash:      ev.Hash,
		Date:      ev.Date,
		Currency:  ev.Currency,
		Impact:    ev.Impact,
		Detail:    ev.Detail,
		Actual:    ev.Actual,
		Forecast:  ev.Forecast,
		Previous:  ev.Previous,
		SourceURL: ev.SourceURL,
	}
}

func (m CalendarEventModel) toDomain() domain.CalendarEvent {
	return domain.CalendarEvent{
		Hash:      m.Hash,
		Date:      m.Date,
		Currency:  m.Currency,
		Impact:    m.Impact,
		Detail:    m.Detail,
		Actual:    m.Actual,
		Forecast:  m.Forecast,
		Previous:  m.Previous,
		SourceURL: m.SourceURL,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}

type UserPositionModel struct {
	ID           int64          `gorm:"column:id"`
	UserID       string         `gorm:"column:user_id;not null"`
	Ticket       int64          `gorm:"column:ticket;not null;uniqueIndex"`
	Platform     *string        `gorm:"column:platform"`
	AccountName  *string        `gorm:"column:account_name"`
	AccountLogin *string        `gorm:"column:account_login"`
	BrokerServer *string        `gorm:"column:broker_server"`
	BrokerName   *string        `gorm:"column:broker_name"`
	Balance      float64        `gorm:"column:balance"`
	Status       string         `gorm:"column:status;not null"`
	Symbol       *string        `gorm:"column:symbol"`
	Side         *string        `gorm:"column:side"`
	Volume       float64        `gorm:"column:volume"`
	EntryTime    time.Time      `gorm:"column:entry_time"`
	EntryPrice   float64        `gorm:"column:entry_price"`
	StopLoss     float64        `gorm:"column:stop_loss"`
	TakeProfit   float64        `gorm:"column:take_profit"`
	Profit       float64        `gorm:"column:profit"`
	RiskPercent  float64        `gorm:"column:risk_percent"`
	TickValue    float64        `gorm:"column:tick_value"`
	TickSize     float64        `gorm:"column:tick_size"`
	Magic        int64          `gorm:"column:magic"`
	Reason       *string        `gorm:"column:reason"`
	Comment      *string        `gorm:"column:comment"`
	LastUpdate   time.Time      `gorm:"column:last_update"`
	RawPayload   datatypes.JSON `gorm:"column:raw_payload;type:jsonb"`
	CreatedAt    time.Time      `gorm:"column:created_at"`
	UpdatedAt    time.Time      `gorm:"column:updated_at"`
}

func (UserPositionModel) TableName() string {
	return "user_positions"
}

func toUserPositionModel(position domain.UserPosition) UserPositionModel {
	return UserPositionModel{
		UserID:       position.UserID,
		Ticket:       position.Ticket,
		Platform:     stringPointerOrNil(position.Platform),
		AccountName:  stringPointerOrNil(position.AccountName),
		AccountLogin: stringPointerOrNil(position.AccountLogin),
		BrokerServer: stringPointerOrNil(position.BrokerServer),
		BrokerName:   stringPointerOrNil(position.BrokerName),
		Balance:      position.Balance,
		Status:       string(position.Status),
		Symbol:       stringPointerOrNil(position.Symbol),
		Side:         stringPointerOrNil(string(position.Side)),
		Volume:       position.Volume,
		EntryTime:    position.EntryTime,
		EntryPrice:   position.EntryPrice,
		StopLoss:     position.StopLoss,
		TakeProfit:   position.TakeProfit,
		Profit:       position.Profit,
		RiskPercent:  position.RiskPercent,
		TickValue:    position.TickValue,
		TickSize:     position.TickSize,
		Magic:        position.Magic,
		Reason:       stringPointerOrNil(position.Reason),
		Comment:      stringPointerOrNil(position.Comment),
		LastUpdate:   position.LastUpdate,
		RawPayload:   jsonOrEmpty(position.RawPayload),
	}
}

func (m UserPositionModel) toDomain() domain.UserPosition {
	return domain.UserPosition{
		UserID:       m.UserID,
		Ticket:       m.Ticket,
		Platform:     stringValueOrEmpty(m.Platform),
		AccountName:  stringValueOrEmpty(m.AccountName),
		AccountLogin: stringValueOrEmpty(m.AccountLogin),
		BrokerServer: stringValueOrEmpty(m.BrokerServer),
		BrokerName:   stringValueOrEmpty(m.BrokerName),
		Balance:      m.Balance,
		Status:       domain.PositionStatus(m.Status),
		Symbol:       stringValueOrEmpty(m.Symbol),
		Side:         domain.PositionSide(stringValueOrEmpty(m.Side)),
		Volume:       m.Volume,
		EntryTime:    m.EntryTime,
		EntryPrice:   m.EntryPrice,
		StopLoss:     m.StopLoss,
		TakeProfit:   m.TakeProfit,
		Profit:       m.Profit,
		RiskPercent:  m.RiskPercent,
		TickValue:    m.TickValue,
		TickSize:     m.TickSize,
		Magic:        m.Magic,
		Reason:       stringValueOrEmpty(m.Reason),
		Comment:      stringValueOrEmpty(m.Comment),
		LastUpdate:   m.LastUpdate,
		RawPayload:   copyJSON(m.RawPayload),
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
	}
}

type UserTradeModel struct {
	ID             int64          `gorm:"column:id"`
	UserID         string         `gorm:"column:user_id;not null"`
	Ticket         int64          `gorm:"column:ticket;not null"`
	PositionTicket *int64         `gorm:"column:position_ticket"`
	Platform       *string        `gorm:"column:platform"`
	Symbol         *string        `gorm:"column:symbol"`
	Side           *string        `gorm:"column:side"`
	EntryType      *string        `gorm:"column:entry_type"`
	Volume         float64        `gorm:"column:volume"`
	EntryTime      time.Time      `gorm:"column:entry_time"`
	EntryPrice     float64        `gorm:"column:entry_price"`
	StopLoss       float64        `gorm:"column:stop_loss"`
	TakeProfit     float64        `gorm:"column:take_profit"`
	Profit         float64        `gorm:"column:profit"`
	RiskPercent    float64        `gorm:"column:risk_percent"`
	TickValue      float64        `gorm:"column:tick_value"`
	TickSize       float64        `gorm:"column:tick_size"`
	Magic          int64          `gorm:"column:magic"`
	Reason         *string        `gorm:"column:reason"`
	Comment        *string        `gorm:"column:comment"`
	RawPayload     datatypes.JSON `gorm:"column:raw_payload;type:jsonb"`
	CreatedAt      time.Time      `gorm:"column:created_at"`
	UpdatedAt      time.Time      `gorm:"column:updated_at"`
}

func (UserTradeModel) TableName() string {
	return "user_trades"
}

func toUserTradeModel(trade domain.UserTrade) UserTradeModel {
	return UserTradeModel{
		UserID:         trade.UserID,
		Ticket:         trade.Ticket,
		PositionTicket: trade.PositionTicket,
		Platform:       stringPointerOrNil(trade.Platform),
		Symbol:         stringPointerOrNil(trade.Symbol),
		Side:           stringPointerOrNil(string(trade.Side)),
		EntryType:      stringPointerOrNil(string(trade.EntryType)),
		Volume:         trade.Volume,
		EntryTime:      trade.EntryTime,
		EntryPrice:     trade.EntryPrice,
		StopLoss:       trade.StopLoss,
		TakeProfit:     trade.TakeProfit,
		Profit:         trade.Profit,
		RiskPercent:    trade.RiskPercent,
		TickValue:      trade.TickValue,
		TickSize:       trade.TickSize,
		Magic:          trade.Magic,
		Reason:         stringPointerOrNil(trade.Reason),
		Comment:        stringPointerOrNil(trade.Comment),
		RawPayload:     jsonOrEmpty(trade.RawPayload),
	}
}

func (m UserTradeModel) toDomain() domain.UserTrade {
	return domain.UserTrade{
		UserID:         m.UserID,
		Ticket:         m.Ticket,
		PositionTicket: m.PositionTicket,
		Platform:       stringValueOrEmpty(m.Platform),
		Symbol:         stringValueOrEmpty(m.Symbol),
		Side:           domain.TradeSide(stringValueOrEmpty(m.Side)),
		EntryType:      domain.TradeEntryType(stringValueOrEmpty(m.EntryType)),
		Volume:         m.Volume,
		EntryTime:      m.EntryTime,
		EntryPrice:     m.EntryPrice,
		StopLoss:       m.StopLoss,
		TakeProfit:     m.TakeProfit,
		Profit:         m.Profit,
		RiskPercent:    m.RiskPercent,
		TickValue:      m.TickValue,
		TickSize:       m.TickSize,
		Magic:          m.Magic,
		Reason:         stringValueOrEmpty(m.Reason),
		Comment:        stringValueOrEmpty(m.Comment),
		RawPayload:     copyJSON(m.RawPayload),
		CreatedAt:      m.CreatedAt,
		UpdatedAt:      m.UpdatedAt,
	}
}

type UserModel struct {
	ID        int64          `gorm:"column:id"`
	UserID    string         `gorm:"column:user_id;primaryKey"`
	Platform  *string        `gorm:"column:platform"`
	Name      *string        `gorm:"column:name"`
	Login     *string        `gorm:"column:login"`
	Server    *string        `gorm:"column:server"`
	Broker    *string        `gorm:"column:broker"`
	Metadata  datatypes.JSON `gorm:"column:metadata;type:jsonb"`
	LastSeen  time.Time      `gorm:"column:last_seen"`
	CreatedAt time.Time      `gorm:"column:created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at"`
}

func (UserModel) TableName() string {
	return "users"
}

func toUserModel(user domain.User) UserModel {
	return UserModel{
		UserID:   user.UserID,
		Platform: stringPointerOrNil(user.Platform),
		Name:     stringPointerOrNil(user.Name),
		Login:    stringPointerOrNil(user.Login),
		Server:   stringPointerOrNil(user.Server),
		Broker:   stringPointerOrNil(user.Broker),
		Metadata: jsonOrEmpty(user.Metadata),
		LastSeen: user.LastSeen,
	}
}

func (m UserModel) toDomain() domain.User {
	return domain.User{
		UserID:   m.UserID,
		Platform: stringValueOrEmpty(m.Platform),
		Name:     stringValueOrEmpty(m.Name),
		Login:    stringValueOrEmpty(m.Login),
		Server:   stringValueOrEmpty(m.Server),
		Broker:   stringValueOrEmpty(m.Broker),
		Metadata: copyJSON(m.Metadata),
		LastSeen: m.LastSeen,
		Created:  m.CreatedAt,
		Updated:  m.UpdatedAt,
	}
}

func stringPointerOrNil(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

func stringValueOrEmpty(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func jsonOrEmpty(data []byte) datatypes.JSON {
	if len(data) == 0 {
		return datatypes.JSON([]byte("{}"))
	}
	return datatypes.JSON(append([]byte(nil), data...))
}

func copyJSON(data datatypes.JSON) []byte {
	if len(data) == 0 {
		return nil
	}
	cpy := make([]byte, len(data))
	copy(cpy, data)
	return cpy
}
