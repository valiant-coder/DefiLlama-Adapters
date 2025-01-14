package ckhdb

import (
	"context"
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

func (r *ClickHouseRepo) GetLatestKlines(ctx context.Context, poolID uint64) ([]*Kline, error) {
	var klines []*Kline
	err := r.WithContext(ctx).Raw(`
		SELECT 
			k.*
		FROM klines_view k
		INNER JOIN (
			SELECT 
				pool_id,
				interval,
				MAX(interval_start) as max_interval_start
			FROM klines_view
			WHERE pool_id = ?
			GROUP BY pool_id, interval
		) latest 
		ON k.pool_id = latest.pool_id 
		AND k.interval = latest.interval 
		AND k.interval_start = latest.max_interval_start
		WHERE k.pool_id = ?
		ORDER BY k.interval;
	`, poolID, poolID).Scan(&klines).Error
	return klines, err
}

func (r *ClickHouseRepo) GetRecentClosePrice(ctx context.Context, poolID uint64, interval string, start time.Time) (decimal.Decimal, error) {
	var klines []Kline
	err := r.WithContext(ctx).Where("pool_id = ? AND interval = ? AND interval_start < ?", poolID, interval, start).Order("interval_start DESC").Limit(1).Find(&klines).Error
	if err != nil {
		return decimal.Decimal{}, err
	}
	if len(klines) == 0 {
		return decimal.Decimal{}, nil
	}
	return klines[0].Close, nil
}
