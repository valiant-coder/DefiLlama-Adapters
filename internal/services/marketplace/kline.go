package marketplace

import (
	"context"
	"encoding/json"
	"errors"
	"exapp-go/internal/db/ckhdb"
	"exapp-go/internal/db/db"
	"exapp-go/internal/entity"
	"fmt"
	"sync"
	"time"

	"github.com/shopspring/decimal"
	"golang.org/x/sync/singleflight"
	"gorm.io/gorm"
)

const (
	redisLatestKlineKeyPrefix = "latest_kline"
)

var klineCache *KlineCache
var klineCacheOnce sync.Once

// KlineCache defines the cache structure
type KlineCache struct {
	sync.RWMutex
	data map[string]*CacheItem
}

// CacheItem defines the cache item structure
type CacheItem struct {
	Klines         []entity.Kline
	ExpireTime     time.Time
	LastUpdateTime time.Time // Last update time
}

// NewKlineCache creates a new cache instance
func NewKlineCache() *KlineCache {
	klineCacheOnce.Do(func() {
		klineCache = &KlineCache{
			data: make(map[string]*CacheItem),
		}
		// Start a goroutine to periodically clean expired cache
		go klineCache.cleanExpired()
	})
	return klineCache
}

// cleanExpired cleans expired cache data
func (c *KlineCache) cleanExpired() {
	ticker := time.NewTicker(5 * time.Minute)
	for range ticker.C {
		c.Lock()
		now := time.Now()
		for key, item := range c.data {
			if item.ExpireTime.Before(now) {
				delete(c.data, key)
			}
		}
		c.Unlock()
	}
}

// Set sets cache
func (c *KlineCache) Set(key string, klines []entity.Kline, expiration time.Duration) {
	c.Lock()
	defer c.Unlock()
	c.data[key] = &CacheItem{
		Klines:         klines,
		ExpireTime:     time.Now().Add(expiration),
		LastUpdateTime: time.Now(),
	}
}

// Get gets cache
func (c *KlineCache) Get(key string) ([]entity.Kline, bool) {
	c.RLock()
	defer c.RUnlock()
	if item, exists := c.data[key]; exists && time.Now().Before(item.ExpireTime) {
		return item.Klines, true
	}
	return nil, false
}

// normalizeTimeByInterval normalizes time based on interval
func normalizeTimeByInterval(t time.Time, interval string) time.Time {
	switch interval {
	case "1m":
		return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), 0, 0, t.Location())
	case "5m":
		minute := (t.Minute() / 5) * 5
		return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), minute, 0, 0, t.Location())
	case "15m":
		minute := (t.Minute() / 15) * 15
		return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), minute, 0, 0, t.Location())
	case "30m":
		minute := (t.Minute() / 30) * 30
		return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), minute, 0, 0, t.Location())
	case "1h":
		return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), 0, 0, 0, t.Location())
	case "4h":
		hour := (t.Hour() / 4) * 4
		return time.Date(t.Year(), t.Month(), t.Day(), hour, 0, 0, 0, t.Location())
	case "1d":
		return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	case "1w":
		year, _ := t.ISOWeek()
		// Find Monday of this week
		for t.Weekday() != time.Monday {
			t = t.AddDate(0, 0, -1)
		}
		return time.Date(year, t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	case "1M":
		return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
	default:
		return t
	}
}

// generateCacheKey generates cache key
func generateCacheKey(poolID uint64, interval string, t time.Time) string {
	normalizedTime := normalizeTimeByInterval(t, interval)
	return fmt.Sprintf("kline:%d:%s:%d", poolID, interval, normalizedTime.Unix())
}

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

func getLatestKlineRedisKey(poolID uint64, interval string) string {
	return fmt.Sprintf("{%s:%d}:%s", redisLatestKlineKeyPrefix, poolID, interval)
}

func NewKlineService() *KlineService {
	return &KlineService{
		repo:    db.New(),
		chkRepo: ckhdb.New(),
	}
}

type KlineService struct {
	repo    *db.Repo
	chkRepo *ckhdb.ClickHouseRepo
	sfGroup singleflight.Group // Prevent cache breakdown
}

// validateKlineParams validates input parameters for GetKline
func (s *KlineService) validateKlineParams(interval string, start, end time.Time) error {
	if start.After(end) {
		return fmt.Errorf("start time cannot be later than end time: start=%v, end=%v", start, end)
	}
	if !isValidInterval(interval) {
		return fmt.Errorf("invalid interval: %s", interval)
	}
	return nil
}

// isValidInterval checks if the interval is supported
func isValidInterval(interval string) bool {
	validIntervals := map[string]bool{
		"1m": true, "5m": true, "15m": true, "30m": true,
		"1h": true, "4h": true, "1d": true, "1w": true, "1M": true,
	}
	return validIntervals[interval]
}

// calculateNextKlineTime calculates the next kline time based on interval
func calculateNextKlineTime(current time.Time, interval string) time.Time {
	if interval == "1M" {
		return time.Date(
			current.Year(),
			current.Month()+1,
			1, 0, 0, 0, 0,
			current.Location(),
		)
	}
	return current.Add(getIntervalDuration(interval))
}

// fillMissingKlines fills missing klines with previous close price
func (s *KlineService) fillMissingKlines(
	poolID uint64,
	interval string,
	currentTime time.Time,
	lastValidKline *ckhdb.Kline,
	klineMap map[int64]*ckhdb.Kline,
) *ckhdb.Kline {
	if k, exists := klineMap[currentTime.Unix()]; exists {
		if lastValidKline != nil {
			k.Open = lastValidKline.Close
		}
		return k
	}

	if lastValidKline != nil {
		return &ckhdb.Kline{
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
	}

	return nil
}

// GetLatestKlines gets the latest kline data
func (s *KlineService) GetLatestKlines(ctx context.Context, poolID uint64, interval string, count int) ([]entity.Kline, error) {
	if err := s.validateKlineParams(interval, time.Now().Add(-24*time.Hour), time.Now()); err != nil {
		return nil, fmt.Errorf("parameter validation failed: %w", err)
	}

	now := time.Now()
	normalizedNow := normalizeTimeByInterval(now, interval)
	cacheKey := generateCacheKey(poolID, interval, normalizedNow)

	// Get current kline start time
	currentKlineStart := normalizeTimeByInterval(now, interval)
	duration := getIntervalDuration(interval)

	// Try to get data from cache
	if klines, ok := klineCache.Get(cacheKey); ok {
		if len(klines) >= count {
			// If the last kline is the current ongoing kline, need to update it
			if len(klines) > 0 {
				lastKline := klines[len(klines)-1]
				// Get the start time of the last kline
				lastKlineTime := time.Time(lastKline.Timestamp)
				if lastKlineTime.Equal(currentKlineStart) {
					redisKey := getLatestKlineRedisKey(poolID, interval)
					latestKlineData, err := s.repo.Redis().Get(ctx, redisKey).Bytes()
					if err == nil {
						var latestKline ckhdb.Kline
						if err := json.Unmarshal(latestKlineData, &latestKline); err == nil {
							klines[len(klines)-1] = entity.DbKlineToEntity(&latestKline)
							klineCache.Set(cacheKey, klines, s.getCacheExpiration(interval))
						}
					}
				}
			}
			return klines[len(klines)-count:], nil
		}
	}

	// Calculate start time
	start := normalizedNow.Add(-duration * time.Duration(count+5)) // Get 5 more to ensure data completeness
	end := now

	// Use singleflight to prevent cache breakdown
	result, err, _ := s.sfGroup.Do(cacheKey, func() (interface{}, error) {
		// Get kline data
		klines, err := s.chkRepo.GetKline(ctx, poolID, interval, start, end)
		if err != nil {
			return nil, fmt.Errorf("failed to get kline data: %w", err)
		}

		// Get the last kline before start time
		lastKline, err := s.chkRepo.GetLastKlineBefore(ctx, poolID, interval, start)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("failed to get previous kline data: %w", err)
		}

		// Create kline data mapping
		klineMap := make(map[int64]*ckhdb.Kline, len(klines))
		for _, k := range klines {
			klineMap[k.IntervalStart.Unix()] = k
		}

		// Estimate number of klines to generate
		estimatedSize := int(end.Sub(start)/duration) + 1
		completeKlines := make([]*ckhdb.Kline, 0, estimatedSize)

		// Generate complete time series
		var lastValidKline *ckhdb.Kline = lastKline
		for currentTime := start; currentTime.Before(end) || currentTime.Equal(end); currentTime = calculateNextKlineTime(currentTime, interval) {
			if newKline := s.fillMissingKlines(poolID, interval, currentTime, lastValidKline, klineMap); newKline != nil {
				completeKlines = append(completeKlines, newKline)
				lastValidKline = newKline
			}
		}

		// Handle price continuity
		for i := 0; i < len(completeKlines); i++ {
			if i == 0 && lastKline != nil {
				completeKlines[i].Open = lastKline.Close
			} else if i > 0 {
				completeKlines[i].Open = completeKlines[i-1].Close
			}

			// Ensure High/Low reasonability
			if completeKlines[i].Low.GreaterThan(completeKlines[i].Open) {
				completeKlines[i].Low = completeKlines[i].Open
			}
			if completeKlines[i].High.LessThan(completeKlines[i].Open) {
				completeKlines[i].High = completeKlines[i].Open
			}
		}

		// Convert to entity objects
		entityKlines := make([]entity.Kline, 0, len(completeKlines))
		for _, kline := range completeKlines {
			entityKlines = append(entityKlines, entity.DbKlineToEntity(kline))
		}

		// Set cache
		expiration := s.getCacheExpiration(interval)
		klineCache.Set(cacheKey, entityKlines, expiration)

		return entityKlines, nil
	})

	if err != nil {
		return nil, err
	}

	klines := result.([]entity.Kline)
	if len(klines) >= count {
		return klines[len(klines)-count:], nil
	}
	return klines, nil
}

func (s *KlineService) GetKline(ctx context.Context, poolID uint64, interval string, start time.Time, end time.Time) ([]entity.Kline, error) {
	// Validate input parameters
	if err := s.validateKlineParams(interval, start, end); err != nil {
		return nil, fmt.Errorf("parameter validation failed: %w", err)
	}

	// If requesting recent kline data, use optimized method
	duration := getIntervalDuration(interval)
	if end.Sub(start) <= duration*150 && time.Since(end) < duration {
		count := int(end.Sub(start)/duration) + 1
		return s.GetLatestKlines(ctx, poolID, interval, count)
	}

	// For historical data requests, use original logic
	normalizedStart := normalizeTimeByInterval(start, interval)
	normalizedEnd := normalizeTimeByInterval(end, interval)

	// Generate cache key
	cacheKey := fmt.Sprintf("kline_history:%d:%s:%d:%d", poolID, interval, normalizedStart.Unix(), normalizedEnd.Unix())

	// Try to get data from cache
	if klines, ok := klineCache.Get(cacheKey); ok {
		return klines, nil
	}

	// Use singleflight to prevent cache breakdown
	result, err, _ := s.sfGroup.Do(cacheKey, func() (interface{}, error) {
		// Get kline data
		klines, err := s.chkRepo.GetKline(ctx, poolID, interval, normalizedStart, normalizedEnd)
		if err != nil {
			return nil, fmt.Errorf("failed to get kline data: %w", err)
		}

		// Get the last kline before start time
		lastKline, err := s.chkRepo.GetLastKlineBefore(ctx, poolID, interval, normalizedStart)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("failed to get previous kline data: %w", err)
		}

		// Create kline data mapping
		klineMap := make(map[int64]*ckhdb.Kline, len(klines))
		for _, k := range klines {
			klineMap[k.IntervalStart.Unix()] = k
		}

		// Estimate number of klines to generate
		estimatedSize := int(normalizedEnd.Sub(normalizedStart)/duration) + 1
		completeKlines := make([]*ckhdb.Kline, 0, estimatedSize)

		// Generate complete time series
		var lastValidKline *ckhdb.Kline = lastKline
		for currentTime := normalizedStart; currentTime.Before(normalizedEnd) || currentTime.Equal(normalizedEnd); currentTime = calculateNextKlineTime(currentTime, interval) {
			if newKline := s.fillMissingKlines(poolID, interval, currentTime, lastValidKline, klineMap); newKline != nil {
				completeKlines = append(completeKlines, newKline)
				lastValidKline = newKline
			}
		}

		// Handle price continuity
		for i := 0; i < len(completeKlines); i++ {
			if i == 0 && lastKline != nil {
				completeKlines[i].Open = lastKline.Close
			} else if i > 0 {
				completeKlines[i].Open = completeKlines[i-1].Close
			}

			// Ensure High/Low reasonability
			if completeKlines[i].Low.GreaterThan(completeKlines[i].Open) {
				completeKlines[i].Low = completeKlines[i].Open
			}
			if completeKlines[i].High.LessThan(completeKlines[i].Open) {
				completeKlines[i].High = completeKlines[i].Open
			}
		}

		// Convert to entity objects
		entityKlines := make([]entity.Kline, 0, len(completeKlines))
		for _, kline := range completeKlines {
			entityKlines = append(entityKlines, entity.DbKlineToEntity(kline))
		}

		// Set cache, historical data can be cached longer
		expiration := s.getCacheExpiration(interval) * 2
		klineCache.Set(cacheKey, entityKlines, expiration)

		return entityKlines, nil
	})

	if err != nil {
		return nil, err
	}

	return result.([]entity.Kline), nil
}

// getCacheExpiration returns cache expiration time based on different intervals
func (s *KlineService) getCacheExpiration(interval string) time.Duration {
	switch interval {
	case "1m":
		return 1 * time.Minute
	case "5m":
		return 5 * time.Minute
	case "15m":
		return 15 * time.Minute
	case "30m":
		return 30 * time.Minute
	case "1h":
		return 1 * time.Hour
	case "4h":
		return 4 * time.Hour
	case "1d":
		return 24 * time.Hour
	case "1w":
		return 7 * 24 * time.Hour
	case "1M":
		return 30 * 24 * time.Hour
	default:
		return 5 * time.Minute
	}
}
