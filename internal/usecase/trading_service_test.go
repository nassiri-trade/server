package usecase

import (
	"testing"
	"time"

	"trading_server/internal/domain"
)

func TestComputeTradeMetrics(t *testing.T) {
	now := time.Now()
	trades := []domain.UserTrade{
		{Profit: 100, EntryTime: now},
		{Profit: -50, EntryTime: now.Add(time.Minute)},
		{Profit: 200, EntryTime: now.Add(2 * time.Minute)},
	}

	metrics := computeTradeMetrics(trades)

	if metrics.TotalTrades != 3 {
		t.Fatalf("expected 3 trades, got %d", metrics.TotalTrades)
	}
	if metrics.WinRate <= 0 {
		t.Fatalf("expected win rate > 0, got %f", metrics.WinRate)
	}
	if metrics.BestTrade != 200 {
		t.Fatalf("unexpected best trade: %f", metrics.BestTrade)
	}
	if metrics.WorstTrade != -50 {
		t.Fatalf("unexpected worst trade: %f", metrics.WorstTrade)
	}
	if metrics.ProfitFactor <= 1 {
		t.Fatalf("expected profit factor > 1, got %f", metrics.ProfitFactor)
	}
}

func TestComputeRiskMetrics(t *testing.T) {
	now := time.Now()
	trades := []domain.UserTrade{
		{Profit: 100, EntryTime: now},
		{Profit: -40, EntryTime: now.Add(time.Hour)},
		{Profit: 60, EntryTime: now.Add(2 * time.Hour)},
		{Profit: -20, EntryTime: now.Add(3 * time.Hour)},
	}

	risk, balance, equity, realized, returns := computeRiskMetrics(trades)

	if realized == 0 {
		t.Fatalf("expected realized profit")
	}
	if len(balance) == 0 || len(equity) == 0 {
		t.Fatalf("expected balance and equity curves")
	}
	if len(returns) != len(trades) {
		t.Fatalf("expected returns length %d, got %d", len(trades), len(returns))
	}
	if risk.MaxDrawdown < 0 {
		t.Fatalf("drawdown should be non-negative")
	}
}
