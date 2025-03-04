package handler

import (
	"context"
	"exapp-go/internal/db/ckhdb"
	"exapp-go/internal/entity"
	"fmt"
	"log"
	"time"

	"github.com/shopspring/decimal"
)

func (s *Service) newTrade(ctx context.Context, trade *ckhdb.Trade) error {
	go func() {
		totalCleaned, err := s.repo.CleanInvalidDepth(ctx, trade.PoolID, trade.Price, trade.TakerIsBid)
		if err != nil {
			log.Printf("clean invalid depth failed: %v", err)
		}
		log.Printf("cleaned %d invalid depth", totalCleaned)
	}()
	err := s.ckhRepo.InsertTradeIfNotExist(ctx, trade)
	if err != nil {
		log.Printf("insert trade failed: %v", err)
		return nil
	}

	orderTag := fmt.Sprintf("%d-%d-%d", trade.PoolID, trade.TakerOrderID, map[bool]int{true: 0, false: 1}[trade.TakerIsBid])
	if s.tradeCache == nil {
		s.tradeCache = make(map[string][]*ckhdb.Trade)
	}
	s.tradeCache[orderTag] = append(s.tradeCache[orderTag], trade)


	var buyer, seller string
	if trade.TakerIsBid {
		buyer = trade.Taker
		seller = trade.Maker
	} else {
		buyer = trade.Maker
		seller = trade.Taker
	}
	go s.publisher.PublishTradeUpdate(entity.Trade{
		PoolID:   trade.PoolID,
		Buyer:    buyer,
		Seller:   seller,
		Quantity: trade.BaseQuantity.String(),
		Price:    trade.Price.String(),
		TradedAt: entity.Time(trade.Time),
		Side:     entity.TradeSide(map[bool]string{true: "buy", false: "sell"}[trade.TakerIsBid]),
	})

	// Get kline data from cache
	klineMap, ok := s.klineCache[trade.PoolID]
	if !ok {
		// Initialize a new map if no data exists for this trading pair
		klineMap = make(map[ckhdb.KlineInterval]*ckhdb.Kline)
		s.klineCache[trade.PoolID] = klineMap
	}

	// Get data from database if cache is empty
	if len(klineMap) == 0 {
		klines, err := s.ckhRepo.GetLatestTwoKlines(ctx, trade.PoolID)
		if err != nil {
			log.Printf("get latest kline failed: %v", err)
			return nil
		}

		// Group kline data by interval
		tmpKlineMap := make(map[ckhdb.KlineInterval][]*ckhdb.Kline)
		for _, kline := range klines {
			tmpKlineMap[kline.Interval] = append(tmpKlineMap[kline.Interval], kline)
		}

		// Process klines for each interval
		for interval, intervalKlines := range tmpKlineMap {
			if len(intervalKlines) > 0 {
				// Set latest kline's open price to previous kline's close price if two klines exist
				if len(intervalKlines) == 2 {
					intervalKlines[0].Open = intervalKlines[1].Close
				}
				klineMap[interval] = intervalKlines[0] // Only cache the latest kline
			}
		}
	}

	// Update all kline intervals
	intervals := []ckhdb.KlineInterval{
		ckhdb.KlineInterval1m,
		ckhdb.KlineInterval5m,
		ckhdb.KlineInterval15m,
		ckhdb.KlineInterval30m,
		ckhdb.KlineInterval1h,
		ckhdb.KlineInterval4h,
		ckhdb.KlineInterval1d,
		ckhdb.KlineInterval1w,
		ckhdb.KlineInterval1M,
	}

	for _, interval := range intervals {
		latestKline, exists := klineMap[interval]
		if !exists {
			// Create new kline if no data exists for this interval
			latestKline = &ckhdb.Kline{
				PoolID:        trade.PoolID,
				IntervalStart: s.getIntervalStart(trade.Time, interval),
				Interval:      interval,
				Open:          trade.Price,
				High:          trade.Price,
				Low:           trade.Price,
				Close:         trade.Price,
				Volume:        trade.BaseQuantity,
				QuoteVolume:   trade.QuoteQuantity,
				Trades:        1,
				UpdateTime:    trade.Time,
			}
			klineMap[interval] = latestKline
		} else {
			// Check if trade is within current kline period
			var periodEnd time.Time
			if interval == ckhdb.KlineInterval1M {
				periodEnd = latestKline.IntervalStart.Add(s.getMonthDuration(latestKline.IntervalStart))
			} else {
				periodEnd = latestKline.IntervalStart.Add(s.getIntervalDuration(interval))
			}

			if trade.Time.Before(periodEnd) {
				// Update current kline
				latestKline.High = decimal.Max(latestKline.High, trade.Price)
				latestKline.Low = decimal.Min(latestKline.Low, trade.Price)
				latestKline.Close = trade.Price
				latestKline.Volume = latestKline.Volume.Add(trade.BaseQuantity)
				latestKline.QuoteVolume = latestKline.QuoteVolume.Add(trade.QuoteQuantity)
				latestKline.Trades++
				latestKline.UpdateTime = trade.Time
			} else {
				// Create new kline period
				// var high, low decimal.Decimal
				// if latestKline.High.GreaterThan(trade.Price) {
				// 	high = latestKline.High
				// } else {
				// 	high = trade.Price
				// }
				// if latestKline.Low.LessThan(trade.Price) {
				// 	low = latestKline.Low
				// } else {
				// 	low = trade.Price
				// }
				newKline := &ckhdb.Kline{
					PoolID:        trade.PoolID,
					IntervalStart: s.getIntervalStart(trade.Time, interval),
					Interval:      interval,
					Open:          latestKline.Close, // Use previous kline's close price as open price
					High:          trade.Price,
					Low:           trade.Price,
					Close:         trade.Price,
					Volume:        trade.BaseQuantity,
					QuoteVolume:   trade.QuoteQuantity,
					Trades:        1,
					UpdateTime:    trade.Time,
				}
				klineMap[interval] = newKline
			}
		}

		// Publish kline update
		err = s.publisher.PublishKlineUpdate(entity.DbKlineToEntity(klineMap[interval]))
		if err != nil {
			log.Printf("publish kline update failed: %v", err)
		}
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
