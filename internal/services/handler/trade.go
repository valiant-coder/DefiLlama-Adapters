package handler

import (
	"context"
	"encoding/json"
	"exapp-go/internal/db/ckhdb"
	"exapp-go/internal/entity"
	"fmt"
	"log"
	"time"

	"github.com/shopspring/decimal"
)

const (
	redisLatestKlineKeyPrefix = "latest_kline"
)

func getLatestKlineRedisKey(poolID uint64, interval ckhdb.KlineInterval) string {
	return fmt.Sprintf("{%s:%d}:%s", redisLatestKlineKeyPrefix, poolID, interval)
}

func (s *Service) newTrade(ctx context.Context, trade *ckhdb.Trade) error {

	// Reduce lock scope
	func() {
		s.mu.Lock()
		defer s.mu.Unlock()
		s.lastTrades[trade.PoolID] = trade
	}()

	s.tradeBuffer.Add(trade)
	orderTag := fmt.Sprintf("%d-%d-%d", trade.PoolID, trade.TakerOrderID, map[bool]int{true: 0, false: 1}[trade.TakerIsBid])
	if s.tradeCache == nil {
		s.tradeCache = make(map[string][]*ckhdb.Trade)
	}
	s.tradeCache[orderTag] = append(s.tradeCache[orderTag], trade)

	// Pre-calculate buyer and seller info
	var buyer, seller string
	if trade.TakerIsBid {
		buyer = trade.Taker
		seller = trade.Maker
	} else {
		buyer = trade.Maker
		seller = trade.Taker
	}

	// Asynchronously publish trade update
	go s.publisher.PublishTradeUpdate(entity.Trade{
		PoolID:   trade.PoolID,
		Buyer:    buyer,
		Seller:   seller,
		Quantity: trade.BaseQuantity.String(),
		Price:    trade.Price.String(),
		TradedAt: entity.Time(trade.Time),
		Side:     entity.TradeSide(map[bool]string{true: "buy", false: "sell"}[trade.TakerIsBid]),
	})

	// Get or initialize kline map, use function scope to limit lock range
	klineMap := func() map[ckhdb.KlineInterval]*ckhdb.Kline {
		s.mu.Lock()
		defer s.mu.Unlock()
		if s.klineCache == nil {
			s.klineCache = make(map[uint64]map[ckhdb.KlineInterval]*ckhdb.Kline)
		}
		klineMap, ok := s.klineCache[trade.PoolID]
		if !ok {
			klineMap = make(map[ckhdb.KlineInterval]*ckhdb.Kline, 8) // Pre-allocate capacity
			s.klineCache[trade.PoolID] = klineMap
		}
		return klineMap
	}()

	// If cache is empty, get data from database
	if len(klineMap) == 0 {
		klines, err := s.ckhRepo.GetLatestTwoKlines(ctx, trade.PoolID)
		if err != nil {
			log.Printf("get latest kline failed: %v", err)
			return nil
		}

		// Use pre-allocated map to store kline data
		tmpKlineMap := make(map[ckhdb.KlineInterval][]*ckhdb.Kline, len(klines))
		for _, kline := range klines {
			tmpKlineMap[kline.Interval] = append(tmpKlineMap[kline.Interval], kline)
		}

		for interval, intervalKlines := range tmpKlineMap {
			if len(intervalKlines) > 0 {
				if len(intervalKlines) == 2 {
					intervalKlines[0].Open = intervalKlines[1].Close
				}
				klineMap[interval] = intervalKlines[0]
			}
		}
	}

	// Pre-define all intervals
	intervals := []ckhdb.KlineInterval{
		ckhdb.KlineInterval1m,
		ckhdb.KlineInterval5m,
		ckhdb.KlineInterval15m,
		ckhdb.KlineInterval30m,
		ckhdb.KlineInterval1h,
		ckhdb.KlineInterval4h,
		ckhdb.KlineInterval1d,
	}

	// Cache trade time to avoid repeated retrieval
	tradeTime := trade.Time

	// Batch process kline updates
	updates := make([]entity.Kline, 0, len(intervals))
	redisUpdates := make(map[string]*ckhdb.Kline, len(intervals))

	for _, interval := range intervals {
		latestKline, exists := klineMap[interval]
		if !exists {
			// Create new kline
			intervalStart := s.getIntervalStart(tradeTime, interval)
			latestKline = &ckhdb.Kline{
				PoolID:        trade.PoolID,
				IntervalStart: intervalStart,
				Interval:      interval,
				Open:          trade.Price,
				High:          trade.Price,
				Low:           trade.Price,
				Close:         trade.Price,
				Volume:        trade.BaseQuantity,
				QuoteVolume:   trade.QuoteQuantity,
				Trades:        1,
				UpdateTime:    tradeTime,
			}
			klineMap[interval] = latestKline
			redisUpdates[string(interval)] = latestKline
		} else {
			var periodEnd time.Time
			if interval == ckhdb.KlineInterval1M {
				periodEnd = latestKline.IntervalStart.Add(s.getMonthDuration(latestKline.IntervalStart))
			} else {
				periodEnd = latestKline.IntervalStart.Add(s.getIntervalDuration(interval))
			}

			if tradeTime.Before(periodEnd) {
				// Update current kline
				latestKline.High = decimal.Max(latestKline.High, trade.Price)
				latestKline.Low = decimal.Min(latestKline.Low, trade.Price)
				latestKline.Close = trade.Price
				latestKline.Volume = latestKline.Volume.Add(trade.BaseQuantity)
				latestKline.QuoteVolume = latestKline.QuoteVolume.Add(trade.QuoteQuantity)
				latestKline.Trades++
				latestKline.UpdateTime = tradeTime
				redisUpdates[string(interval)] = latestKline
			} else {
				// Create new kline period
				intervalStart := s.getIntervalStart(tradeTime, interval)
				newKline := &ckhdb.Kline{
					PoolID:        trade.PoolID,
					IntervalStart: intervalStart,
					Interval:      interval,
					Open:          latestKline.Close,
					High:          trade.Price,
					Low:           trade.Price,
					Close:         trade.Price,
					Volume:        trade.BaseQuantity,
					QuoteVolume:   trade.QuoteQuantity,
					Trades:        1,
					UpdateTime:    tradeTime,
				}
				klineMap[interval] = newKline
				redisUpdates[string(interval)] = newKline
			}
		}

		// Collect kline updates
		updates = append(updates, entity.DbKlineToEntity(klineMap[interval]))
	}

	for interval, kline := range redisUpdates {
		interval, kline := interval, kline
		go func() {
			redisKey := getLatestKlineRedisKey(kline.PoolID, ckhdb.KlineInterval(interval))
			klineData, err := json.Marshal(kline)
			if err != nil {
				log.Printf("marshal kline data failed: %v", err)
				return
			}
			expiration := s.getIntervalDuration(ckhdb.KlineInterval(interval)) * 2
			if err := s.repo.Redis().Set(context.Background(), redisKey, klineData, expiration).Err(); err != nil {
				log.Printf("set redis kline data failed: %v", err)
			}
		}()
	}

	// Batch publish kline updates
	for _, update := range updates {
		update := update
		go func() {
			s.publisher.PublishKlineUpdate(update)
		}()
	}
	return nil
}

// getIntervalDuration returns the duration of a kline interval
func (s *Service) getIntervalDuration(interval ckhdb.KlineInterval) time.Duration {
	switch interval {
	case ckhdb.KlineInterval1m:
		return time.Minute
	case ckhdb.KlineInterval5m:
		return 5 * time.Minute
	case ckhdb.KlineInterval15m:
		return 15 * time.Minute
	case ckhdb.KlineInterval30m:
		return 30 * time.Minute
	case ckhdb.KlineInterval1h:
		return time.Hour
	case ckhdb.KlineInterval4h:
		return 4 * time.Hour
	case ckhdb.KlineInterval1d:
		return 24 * time.Hour
	case ckhdb.KlineInterval1w:
		return 7 * 24 * time.Hour
	case ckhdb.KlineInterval1M:
		// For monthly klines, return actual days in month
		now := time.Now()
		return s.getMonthDuration(now)
	default:
		return time.Minute
	}
}

// getMonthDuration returns the duration of the month containing the given time
func (s *Service) getMonthDuration(t time.Time) time.Duration {
	year, month, _ := t.Date()
	firstDayNextMonth := time.Date(year, month+1, 1, 0, 0, 0, 0, t.Location())
	firstDayThisMonth := time.Date(year, month, 1, 0, 0, 0, 0, t.Location())
	return firstDayNextMonth.Sub(firstDayThisMonth)
}

// getIntervalStart returns the start time of a kline interval
func (s *Service) getIntervalStart(t time.Time, interval ckhdb.KlineInterval) time.Time {
	switch interval {
	case ckhdb.KlineInterval1m:
		return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), 0, 0, t.Location())
	case ckhdb.KlineInterval5m:
		return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute()-t.Minute()%5, 0, 0, t.Location())
	case ckhdb.KlineInterval15m:
		return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute()-t.Minute()%15, 0, 0, t.Location())
	case ckhdb.KlineInterval30m:
		return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute()-t.Minute()%30, 0, 0, t.Location())
	case ckhdb.KlineInterval1h:
		return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), 0, 0, 0, t.Location())
	case ckhdb.KlineInterval4h:
		return time.Date(t.Year(), t.Month(), t.Day(), t.Hour()-t.Hour()%4, 0, 0, 0, t.Location())
	case ckhdb.KlineInterval1d:
		return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	case ckhdb.KlineInterval1w:
		return time.Date(t.Year(), t.Month(), t.Day()-int(t.Weekday()), 0, 0, 0, 0, t.Location())
	case ckhdb.KlineInterval1M:
		// For monthly klines, return first day of month
		return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
	default:
		return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), 0, 0, t.Location())
	}
}
