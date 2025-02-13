package marketplace

import (
	"context"
	"errors"
	"exapp-go/internal/db/ckhdb"
	"exapp-go/internal/entity"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
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
		return make([]entity.Kline, 0), err
	}

	lastKline, err := s.repo.GetLastKlineBefore(ctx, poolID, interval, start)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			lastKline = nil
		} else {
			return make([]entity.Kline, 0), err
		}
	}

	// Create a map to store existing kline data
	klineMap := make(map[int64]*ckhdb.Kline)
	for _, k := range klines {
		klineMap[k.IntervalStart.Unix()] = k
	}

	// Calculate time interval
	duration := getIntervalDuration(interval)

	// Generate complete time series
	completeKlines := make([]*ckhdb.Kline, 0)
	currentTime := start
	var lastValidKline *ckhdb.Kline
	if lastKline != nil {
		lastValidKline = lastKline
	}

	// Adjust start time to next complete cycle based on interval
	if interval == "1M" {
		// For monthly kline, adjust to beginning of next month
		if !start.Equal(time.Date(start.Year(), start.Month(), 1, 0, 0, 0, 0, start.Location())) {
			currentTime = time.Date(
				start.Year(),
				start.Month()+1,
				1,
				0, 0, 0, 0,
				start.Location(),
			)
		}
	} else {
		// Calculate start time of next complete cycle
		remainder := start.UnixNano() % duration.Nanoseconds()
		if remainder != 0 {
			currentTime = start.Add(duration - time.Duration(remainder))
		}
	}

	for currentTime.Before(end) {
		if k, exists := klineMap[currentTime.Unix()]; exists {
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
	entityKlines = make([]entity.Kline, 0, len(completeKlines))
	for _, kline := range completeKlines {
		entityKlines = append(entityKlines, entity.DbKlineToEntity(kline))
	}
	return entityKlines, nil
}
