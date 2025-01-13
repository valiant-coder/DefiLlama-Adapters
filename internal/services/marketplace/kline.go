package marketplace

import (
	"context"
	"exapp-go/internal/db/ckhdb"
	"exapp-go/internal/entity"
	"time"

	"github.com/shopspring/decimal"
)

// getIntervalDuration returns the time duration based on the interval string
func getIntervalDuration(interval string) time.Duration {
	switch interval {
	case "1m":
		return time.Minute
	case "5m":
		return 5 * time.Minute
	case "15m":
		return 15 * time.Minute
	case "30m":
		return 30 * time.Minute
	case "1h":
		return time.Hour
	case "4h":
		return 4 * time.Hour
	case "1d":
		return 24 * time.Hour
	case "1w":
		return 7 * 24 * time.Hour
	default:
		return time.Minute
	}
}

func NewKlineService() *KlineService {
	return &KlineService{repo: ckhdb.New()}
}

type KlineService struct {
	repo *ckhdb.ClickHouseRepo
}

func (s *KlineService) GetKline(ctx context.Context, poolID uint64, interval string, start time.Time, end time.Time) ([]entity.Kline, error) {
	var klines []*ckhdb.Kline
	var err error

	if klines, err = s.repo.GetKline(ctx, poolID, interval, start, end); err != nil {
		return nil, err
	}

	lastKline, err := s.repo.GetLastKlineBefore(ctx, poolID, interval, start)
	if err != nil {
		return nil, err
	}

	// Create a map to store existing kline data
	klineMap := make(map[time.Time]*ckhdb.Kline)
	for _, k := range klines {
		klineMap[k.IntervalStart] = k
	}

	// Calculate time interval
	duration := getIntervalDuration(interval)

	// Generate complete time series
	var completeKlines []*ckhdb.Kline
	currentTime := start
	var lastValidKline *ckhdb.Kline
	if lastKline != nil {
		lastValidKline = lastKline
	}

	for currentTime.Before(end) {
		if k, exists := klineMap[currentTime]; exists {
			// If data exists for this timestamp, use actual data
			if lastValidKline == nil {
				k.Open = k.Close
			} else {
				k.Open = lastValidKline.Close
			}
			lastValidKline = k
			completeKlines = append(completeKlines, k)
		} else {
			// If no data exists for this timestamp, create a new kline
			if lastValidKline != nil {
				newKline := &ckhdb.Kline{
					PoolID:        poolID,
					IntervalStart: currentTime,
					Interval:      ckhdb.KlineInterval(interval),
					Open:          lastValidKline.Close,
					High:          lastValidKline.Close,
					Low:           lastValidKline.Close,
					Close:         lastValidKline.Close,
					Volume:        decimal.Zero,
					QuoteVolume:   decimal.Zero,
					Trades:        0,
				}
				completeKlines = append(completeKlines, newKline)
			}
		}

		if interval == "1M" {
			currentTime = time.Date(
				currentTime.Year(),
				currentTime.Month()+1,
				1,
				0, 0, 0, 0,
				currentTime.Location(),
			)
		} else {
			currentTime = currentTime.Add(duration)
		}
	}

	// Handle price data continuity
	for i := 0; i < len(completeKlines); i++ {
		if i == 0 && lastKline != nil {
			completeKlines[i].Open = lastKline.Close
		} else if i > 0 {
			completeKlines[i].Open = completeKlines[i-1].Close
		}
		if completeKlines[i].Low.GreaterThan(completeKlines[i].Open) {
			completeKlines[i].Low = completeKlines[i].Open
		}
		if completeKlines[i].High.LessThan(completeKlines[i].Open) {
			completeKlines[i].High = completeKlines[i].Open
		}
	}

	var entityKlines []entity.Kline
	for _, kline := range completeKlines {
		entityKlines = append(entityKlines, entity.DbKlineToEntity(kline))
	}
	return entityKlines, nil
}
