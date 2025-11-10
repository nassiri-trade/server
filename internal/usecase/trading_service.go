package usecase

import (
	"context"
	"errors"
	"math"
	"sort"
	"time"

	"trading_server/internal/domain"
)

type TradingService struct {
	positionRepo domain.PositionRepository
	tradeRepo    domain.TradeRepository
	userRepo     domain.UserRepository
}

func NewTradingService(positionRepo domain.PositionRepository, tradeRepo domain.TradeRepository, userRepo domain.UserRepository) (*TradingService, error) {
	if positionRepo == nil {
		return nil, errors.New("position repository required")
	}
	if tradeRepo == nil {
		return nil, errors.New("trade repository required")
	}
	return &TradingService{
		positionRepo: positionRepo,
		tradeRepo:    tradeRepo,
		userRepo:     userRepo,
	}, nil
}

func (s *TradingService) UpsertPosition(ctx context.Context, position domain.UserPosition) error {
	if position.UserID == "" {
		return errors.New("user id required")
	}
	if position.Ticket == 0 {
		return errors.New("ticket required")
	}
	if position.Status == "" {
		position.Status = domain.PositionStatusOpen
	}
	if position.LastUpdate.IsZero() {
		position.LastUpdate = time.Now().UTC()
	}

	if err := s.ensureUserFromPosition(ctx, position); err != nil {
		return err
	}

	return s.positionRepo.UpsertPosition(ctx, position)
}

func (s *TradingService) RecordTrade(ctx context.Context, trade domain.UserTrade) error {
	if trade.UserID == "" {
		return errors.New("user id required")
	}
	if trade.Ticket == 0 {
		return errors.New("ticket required")
	}
	if err := s.ensureUserFromTrade(ctx, trade); err != nil {
		return err
	}
	return s.tradeRepo.AddTrade(ctx, trade)
}

func (s *TradingService) GenerateReport(ctx context.Context, userID string, limit int) (domain.UserPerformanceReport, error) {
	if limit <= 0 {
		limit = 1000
	}

	openPositions, err := s.positionRepo.ListOpenPositions(ctx, userID)
	if err != nil {
		return domain.UserPerformanceReport{}, err
	}

	trades, err := s.tradeRepo.ListTrades(ctx, userID, limit)
	if err != nil {
		return domain.UserPerformanceReport{}, err
	}

	sort.Slice(trades, func(i, j int) bool {
		return trades[i].EntryTime.Before(trades[j].EntryTime)
	})

	report := domain.UserPerformanceReport{
		UserID:          userID,
		AsOf:            time.Now().UTC(),
		OpenPositions:   enrichOpenPositions(openPositions),
		ClosedPositions: trades,
	}

	tradeMetrics := computeTradeMetrics(trades)
	riskMetrics, balanceCurve, equityCurve, realized, returns := computeRiskMetrics(trades)
	exposure := computeExposureMetrics(trades, openPositions)

	unrealized := 0.0
	for _, pos := range openPositions {
		unrealized += pos.Profit
	}

	report.TradeMetrics = tradeMetrics
	report.RiskMetrics = riskMetrics
	report.ExposureMetrics = exposure
	report.BalanceCurve = balanceCurve
	report.EquityCurve = equityCurve
	report.RealizedProfit = realized
	report.UnrealizedProfit = unrealized
	report.TotalReturnPercent = computeTotalReturnPercent(realized, returns)
	report.DerivedInsights = buildInsights(tradeMetrics, riskMetrics, exposure)

	return report, nil
}

func enrichOpenPositions(positions []domain.UserPosition) []domain.UserPositionSnapshot {
	out := make([]domain.UserPositionSnapshot, 0, len(positions))
	for _, p := range positions {
		out = append(out, domain.UserPositionSnapshot{
			UserPosition: p,
			UnrealizedPL: p.Profit,
		})
	}
	return out
}

func (s *TradingService) ensureUserFromPosition(ctx context.Context, position domain.UserPosition) error {
	if s.userRepo == nil || position.UserID == "" {
		return nil
	}

	lastSeen := position.LastUpdate
	if lastSeen.IsZero() {
		lastSeen = time.Now().UTC()
	}

	user := domain.User{
		UserID:   position.UserID,
		Platform: position.Platform,
		Name:     position.AccountName,
		Login:    position.AccountLogin,
		Server:   position.BrokerServer,
		Broker:   position.BrokerName,
		Metadata: position.RawPayload,
		LastSeen: lastSeen,
	}
	return s.userRepo.UpsertUser(ctx, user)
}

func (s *TradingService) ensureUserFromTrade(ctx context.Context, trade domain.UserTrade) error {
	if s.userRepo == nil || trade.UserID == "" {
		return nil
	}

	lastSeen := trade.EntryTime
	if lastSeen.IsZero() {
		lastSeen = time.Now().UTC()
	}

	user := domain.User{
		UserID:   trade.UserID,
		Platform: trade.Platform,
		Metadata: trade.RawPayload,
		LastSeen: lastSeen,
	}
	return s.userRepo.UpsertUser(ctx, user)
}

func computeTradeMetrics(trades []domain.UserTrade) domain.TradeMetrics {
	if len(trades) == 0 {
		return domain.TradeMetrics{}
	}

	var winCount, lossCount int
	var sumWin, sumLoss, best, worst float64
	best = math.Inf(-1)
	worst = math.Inf(1)
	var totalDuration float64

	for _, trade := range trades {
		profit := trade.Profit
		if profit > 0 {
			winCount++
			sumWin += profit
		} else if profit < 0 {
			lossCount++
			sumLoss += profit
		}
		if profit > best {
			best = profit
		}
		if profit < worst {
			worst = profit
		}

		if !trade.EntryTime.IsZero() {
			totalDuration += 0 // placeholder until exit time available
		}
	}

	totalTrades := len(trades)
	winRate := 0.0
	if totalTrades > 0 {
		winRate = float64(winCount) / float64(totalTrades)
	}

	avgWin := safeDivide(sumWin, float64(winCount))
	avgLoss := safeDivide(sumLoss, float64(lossCount))
	profitFactor := safeDivide(sumWin, math.Abs(sumLoss))
	payoffRatio := safeDivide(avgWin, math.Abs(avgLoss))

	expectancy := winRate*avgWin + (1-winRate)*avgLoss
	avgDuration := safeDivide(totalDuration, float64(totalTrades))

	return domain.TradeMetrics{
		TotalTrades:        totalTrades,
		WinningTrades:      winCount,
		LosingTrades:       lossCount,
		WinRate:            winRate,
		AverageWin:         avgWin,
		AverageLoss:        avgLoss,
		Expectancy:         expectancy,
		ProfitFactor:       profitFactor,
		PayoffRatio:        payoffRatio,
		AverageRMultiple:   computeAverageR(trades),
		BestTrade:          best,
		WorstTrade:         worst,
		AverageDurationSec: avgDuration,
	}
}

func computeAverageR(trades []domain.UserTrade) float64 {
	var total float64
	var count int
	for _, trade := range trades {
		denominator := math.Abs(trade.RiskPercent)
		if denominator <= 0 {
			continue
		}
		total += trade.Profit / denominator
		count++
	}
	return safeDivide(total, float64(count))
}

func computeRiskMetrics(trades []domain.UserTrade) (domain.RiskMetrics, []domain.EquityPoint, []domain.EquityPoint, float64, []float64) {
	if len(trades) == 0 {
		return domain.RiskMetrics{}, nil, nil, 0, nil
	}

	var equity []domain.EquityPoint
	var balance []domain.EquityPoint
	var returns []float64

	cumProfit := 0.0
	maxEquity := math.Inf(-1)
	peak := 0.0
	maxDrawdown := 0.0
	maxDrawdownPct := 0.0
	var sumReturns float64

	for _, trade := range trades {
		cumProfit += trade.Profit
		if trade.EntryTime.IsZero() {
			continue
		}

		point := domain.EquityPoint{
			Time:    trade.EntryTime,
			Balance: cumProfit,
		}
		equity = append(equity, point)
		balance = append(balance, point)

		if cumProfit > peak {
			peak = cumProfit
		}
		drawdown := peak - cumProfit
		if drawdown > maxDrawdown {
			maxDrawdown = drawdown
			if peak != 0 {
				maxDrawdownPct = drawdown / math.Abs(peak)
			}
		}

		returns = append(returns, trade.Profit)
		sumReturns += trade.Profit
		if cumProfit > maxEquity {
			maxEquity = cumProfit
		}
	}

	meanReturn := safeDivide(sumReturns, float64(len(returns)))
	stdDev := sampleStdDev(returns, meanReturn)
	downsideDev := downsideDeviation(returns, meanReturn)
	sort.Float64s(returns)

	var95 := percentile(returns, 5)
	cvar95 := conditionalValueAtRisk(returns, 5)
	sharpe := safeDivide(meanReturn, stdDev)
	sortino := safeDivide(meanReturn, downsideDev)
	calmar := safeDivide(sumReturns, maxDrawdown)
	recovery := safeDivide(sumReturns, maxDrawdown)
	ulcer := ulcerIndex(equity)
	stability := equityStability(equity)

	return domain.RiskMetrics{
			MaxDrawdown:         maxDrawdown,
			MaxDrawdownPercent:  maxDrawdownPct,
			SharpeRatio:         sharpe,
			SortinoRatio:        sortino,
			CalmarRatio:         calmar,
			RecoveryFactor:      recovery,
			Volatility:          stdDev,
			UlcerIndex:          ulcer,
			Stability:           stability,
			DownsideDeviation:   downsideDev,
			AverageDownsideRisk: downsideDev,
			Skewness:            skewness(returns, meanReturn, stdDev),
			Kurtosis:            kurtosis(returns, meanReturn, stdDev),
			ValueAtRisk95:       var95,
			ConditionalVaR95:    cvar95,
		},
		balance,
		equity,
		cumProfit,
		returns
}

func computeExposureMetrics(trades []domain.UserTrade, positions []domain.UserPosition) domain.ExposureMetrics {
	var exposureTimes []float64
	for range trades {
		// Without explicit exit times, assume trade duration equals zero exposure to avoid misleading metrics.
	}
	for _, pos := range positions {
		if pos.LastUpdate.IsZero() || pos.EntryTime.IsZero() {
			continue
		}
		duration := pos.LastUpdate.Sub(pos.EntryTime).Seconds()
		exposureTimes = append(exposureTimes, duration)
	}

	avgExposure := average(exposureTimes)
	maxExposure := maxFloat(exposureTimes)

	return domain.ExposureMetrics{
		AverageExposureTime: avgExposure,
		MaxExposureTime:     maxExposure,
		AverageLeverage:     0,
		ExposureRatio:       0,
	}
}

func computeTotalReturnPercent(realized float64, returns []float64) float64 {
	if len(returns) == 0 {
		return 0
	}
	initialEquity := 0.0
	finalEquity := realized
	if initialEquity == 0 {
		return 0
	}
	return (finalEquity - initialEquity) / math.Abs(initialEquity)
}

func buildInsights(tradeMetrics domain.TradeMetrics, riskMetrics domain.RiskMetrics, exposure domain.ExposureMetrics) []string {
	var insights []string
	if tradeMetrics.WinRate > 0.6 {
		insights = append(insights, "Consistent win rate above 60%.")
	}
	if riskMetrics.MaxDrawdownPercent > 0.2 {
		insights = append(insights, "Drawdown exceeds 20%; review risk practices.")
	}
	if riskMetrics.SharpeRatio > 1 {
		insights = append(insights, "Sharpe ratio indicates strong risk-adjusted performance.")
	}
	if exposure.MaxExposureTime > 0 {
		insights = append(insights, "Track exposure durations for risk control.")
	}
	return insights
}

func safeDivide(a, b float64) float64 {
	if math.IsNaN(a) || math.IsNaN(b) {
		return 0
	}
	if math.Abs(b) < 1e-9 {
		return 0
	}
	return a / b
}

func sampleStdDev(values []float64, mean float64) float64 {
	if len(values) < 2 {
		return 0
	}
	var sum float64
	for _, v := range values {
		diff := v - mean
		sum += diff * diff
	}
	return math.Sqrt(sum / float64(len(values)-1))
}

func downsideDeviation(values []float64, threshold float64) float64 {
	if len(values) == 0 {
		return 0
	}
	var sum float64
	var count float64
	for _, v := range values {
		if v < threshold {
			diff := threshold - v
			sum += diff * diff
			count++
		}
	}
	if count == 0 {
		return 0
	}
	return math.Sqrt(sum / count)
}

func percentile(values []float64, p float64) float64 {
	if len(values) == 0 {
		return 0
	}
	if p <= 0 {
		return values[0]
	}
	if p >= 100 {
		return values[len(values)-1]
	}
	index := (p / 100) * float64(len(values)-1)
	lower := int(math.Floor(index))
	upper := int(math.Ceil(index))
	if lower == upper {
		return values[lower]
	}
	weight := index - float64(lower)
	return values[lower]*(1-weight) + values[upper]*weight
}

func conditionalValueAtRisk(values []float64, p float64) float64 {
	if len(values) == 0 {
		return 0
	}
	cutoff := percentile(values, p)
	var sum float64
	var count int
	for _, v := range values {
		if v <= cutoff {
			sum += v
			count++
		}
	}
	return safeDivide(sum, float64(count))
}

func ulcerIndex(points []domain.EquityPoint) float64 {
	if len(points) == 0 {
		return 0
	}
	var peak float64
	var sum float64
	for i, p := range points {
		if i == 0 || p.Balance > peak {
			peak = p.Balance
		}
		drawdown := (p.Balance - peak) / peak
		sum += drawdown * drawdown
	}
	return math.Sqrt(sum / float64(len(points)))
}

func equityStability(points []domain.EquityPoint) float64 {
	if len(points) < 2 {
		return 0
	}
	var sumDiff float64
	for i := 1; i < len(points); i++ {
		diff := points[i].Balance - points[i-1].Balance
		sumDiff += math.Abs(diff)
	}
	totalChange := math.Abs(points[len(points)-1].Balance - points[0].Balance)
	return safeDivide(totalChange, sumDiff)
}

func skewness(values []float64, mean, std float64) float64 {
	if len(values) == 0 || std == 0 {
		return 0
	}
	var sum float64
	for _, v := range values {
		sum += math.Pow((v-mean)/std, 3)
	}
	return sum / float64(len(values))
}

func kurtosis(values []float64, mean, std float64) float64 {
	if len(values) == 0 || std == 0 {
		return 0
	}
	var sum float64
	for _, v := range values {
		sum += math.Pow((v-mean)/std, 4)
	}
	return sum/float64(len(values)) - 3
}

func average(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	var sum float64
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

func maxFloat(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	max := values[0]
	for _, v := range values {
		if v > max {
			max = v
		}
	}
	return max
}
