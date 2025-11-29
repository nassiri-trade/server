package domain

import "time"

type PositionStatus string

const (
	PositionStatusOpen   PositionStatus = "open"
	PositionStatusClosed PositionStatus = "closed"
)

type PositionSide string

const (
	PositionSideBuy  PositionSide = "DEAL_TYPE_BUY"
	PositionSideSell PositionSide = "DEAL_TYPE_SELL"
)

type TradeEntryType string

type TradeSide string

type UserPosition struct {
	UserID       string
	Ticket       int64
	Platform     string
	AccountName  string
	AccountLogin string
	BrokerServer string
	BrokerName   string
	Balance      float64
	Status       PositionStatus
	Symbol       string
	Side         PositionSide
	Volume       float64
	EntryTime    time.Time
	EntryPrice   float64
	StopLoss     float64
	TakeProfit   float64
	Profit       float64
	RiskPercent  float64
	TickValue    float64
	TickSize     float64
	Magic        int64
	Reason       string
	Comment      string
	LastUpdate   time.Time
	RawPayload   []byte
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type UserPositionSnapshot struct {
	UserPosition
	UnrealizedPL float64
}

type UserTrade struct {
	UserID         string
	Ticket         int64
	PositionTicket *int64
	Platform       string
	Symbol         string
	Side           TradeSide
	EntryType      TradeEntryType
	Volume         float64
	EntryTime      time.Time
	EntryPrice     float64
	StopLoss       float64
	TakeProfit     float64
	Profit         float64
	RiskPercent    float64
	TickValue      float64
	TickSize       float64
	Magic          int64
	Reason         string
	Comment        string
	RawPayload     []byte
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type UserPerformanceReport struct {
	UserID             string
	AsOf               time.Time
	BalanceCurve       []EquityPoint
	EquityCurve        []EquityPoint
	TradeMetrics       TradeMetrics
	RiskMetrics        RiskMetrics
	ExposureMetrics    ExposureMetrics
	DerivedInsights    []string
	OpenPositions      []UserPositionSnapshot
	ClosedPositions    []UserTrade
	RealizedProfit     float64
	UnrealizedProfit   float64
	TotalReturnPercent float64
}

type EquityPoint struct {
	Time    time.Time
	Balance float64
}

type TradeMetrics struct {
	TotalTrades        int
	WinningTrades      int
	LosingTrades       int
	WinRate            float64
	AverageWin         float64
	AverageLoss        float64
	Expectancy         float64
	ProfitFactor       float64
	PayoffRatio        float64
	AverageRMultiple   float64
	BestTrade          float64
	WorstTrade         float64
	AverageDurationSec float64
}

type RiskMetrics struct {
	MaxDrawdown         float64
	MaxDrawdownPercent  float64
	SharpeRatio         float64
	SortinoRatio        float64
	CalmarRatio         float64
	RecoveryFactor      float64
	Volatility          float64
	UlcerIndex          float64
	Stability           float64
	DownsideDeviation   float64
	AverageDownsideRisk float64
	Skewness            float64
	Kurtosis            float64
	ValueAtRisk95       float64
	ConditionalVaR95    float64
}

type ExposureMetrics struct {
	AverageExposureTime float64
	MaxExposureTime     float64
	AverageLeverage     float64
	ExposureRatio       float64
}

type User struct {
	UserID   string
	Platform string
	Name     string
	Login    string
	Server   string
	Broker   string
	Metadata []byte
	LastSeen time.Time
	Created  time.Time
	Updated  time.Time
}

type Passkey struct {
	ID        int64
	PasskeyID string
	Enabled   bool
	CreatedAt time.Time
	UpdatedAt time.Time
}
