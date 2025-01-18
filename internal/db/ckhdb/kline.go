package ckhdb

import (
	"context"
	"sync"
	"time"

	"github.com/shopspring/decimal"
)

type KlineInterval string

const (
	KlineInterval1m  KlineInterval = "1m"
	KlineInterval5m  KlineInterval = "5m"
	KlineInterval15m KlineInterval = "15m"
	KlineInterval30m KlineInterval = "30m"
	KlineInterval1h  KlineInterval = "1h"
	KlineInterval4h  KlineInterval = "4h"
	KlineInterval1d  KlineInterval = "1d"
	KlineInterval1w  KlineInterval = "1w"
	KlineInterval1M  KlineInterval = "1M"
)

type Kline struct {
	PoolID        uint64          `json:"pool_id"`
	IntervalStart time.Time       `json:"interval_start"`
	Interval      KlineInterval   `json:"interval"`
	Open          decimal.Decimal `json:"open" gorm:"type:Decimal(36,18)"`
	High          decimal.Decimal `json:"high" gorm:"type:Decimal(36,18)"`
	Low           decimal.Decimal `json:"low" gorm:"type:Decimal(36,18)"`
	Close         decimal.Decimal `json:"close" gorm:"type:Decimal(36,18)"`
	Volume        decimal.Decimal `json:"volume" gorm:"type:Decimal(36,18)"`
	QuoteVolume   decimal.Decimal `json:"quote_volume" gorm:"type:Decimal(36,18)"`
	Trades        uint64          `json:"trades"`
	UpdateTime    time.Time       `json:"update_time"`
}

func (Kline) TableName() string {
	return "klines_view"
}

func (r *ClickHouseRepo) GetKline(ctx context.Context, poolID uint64, interval string, start time.Time, end time.Time) ([]*Kline, error) {
	var klines []*Kline
	err := r.WithContext(ctx).Where("pool_id = ? AND interval = ? AND interval_start >= ? AND interval_start <= ?", poolID, interval, start, end).Find(&klines).Error
	return klines, err
}

func (r *ClickHouseRepo) GetLastKlineBefore(ctx context.Context, poolID uint64, interval string, start time.Time) (*Kline, error) {
	var kline *Kline
	err := r.WithContext(ctx).Where("pool_id = ? AND interval = ? AND interval_start < ?", poolID, interval, start).Order("interval_start DESC").Limit(1).First(&kline).Error
	return kline, err
}

type klineCache struct {
	data      []*Kline
	timestamp time.Time
}

var (
	latestKlinesCache sync.Map
	cacheDuration     = 100 * time.Millisecond
)

func (r *ClickHouseRepo) GetLatestKlines(ctx context.Context, poolID uint64) ([]*Kline, error) {
	if cached, ok := latestKlinesCache.Load(poolID); ok {
		cache := cached.(klineCache)
		if time.Since(cache.timestamp) < cacheDuration {
			return cache.data, nil
		}
	}

	var klines []*Kline
	err := r.WithContext(ctx).Raw(`
		WITH latest_intervals AS (
			SELECT 
				pool_id,
				interval,
				MAX(interval_start) as max_interval_start
			FROM klines_view
			WHERE pool_id = ?
			GROUP BY pool_id, interval
		)
		SELECT 
			k.*
		FROM klines_view k
		INNER JOIN latest_intervals li 
		ON k.pool_id = li.pool_id 
		AND k.interval = li.interval 
		AND k.interval_start = li.max_interval_start
		WHERE k.pool_id = ?
		ORDER BY k.interval;
	`, poolID, poolID).Scan(&klines).Error

	if err != nil {
		return nil, err
	}
	latestKlinesCache.Store(poolID, klineCache{
		data:      klines,
		timestamp: time.Now(),
	})

	return klines, nil
}

func (r *ClickHouseRepo) BatchGetLatestKlines(ctx context.Context, poolIDs []uint64) (map[uint64][]*Kline, error) {
	var klines []*Kline
	err := r.WithContext(ctx).Raw(`
		WITH latest_intervals AS (
			SELECT 
				pool_id,
				interval,
				MAX(interval_start) as max_interval_start
			FROM klines_view
			WHERE pool_id IN (?)
			GROUP BY pool_id, interval
		)
		SELECT 
			k.*
		FROM klines_view k
		INNER JOIN latest_intervals li 
		ON k.pool_id = li.pool_id 
		AND k.interval = li.interval 
		AND k.interval_start = li.max_interval_start
		ORDER BY k.pool_id, k.interval;
	`, poolIDs).
		Scan(&klines).Error

	if err != nil {
		return nil, err
	}

	result := make(map[uint64][]*Kline)
	for _, kline := range klines {
		result[kline.PoolID] = append(result[kline.PoolID], kline)
	}
	return result, nil
}


func (r *ClickHouseRepo) GetLatestTwoKlines(ctx context.Context, poolID uint64) ([]*Kline, error) {
	var klines []*Kline
	err := r.WithContext(ctx).Raw(`	
SELECT 
    pool_id,
    interval,
    interval_start,
    open,
    high,
    low,
    close,
    volume,
    quote_volume,
    trades,
    rank
FROM (
    SELECT 
        *,
        row_number() OVER (PARTITION BY pool_id, interval ORDER BY interval_start DESC) as rank
    FROM klines_view
    where pool_id = ?
)
WHERE rank <= 2
	`, poolID).Scan(&klines).Error
	return klines, err
}
