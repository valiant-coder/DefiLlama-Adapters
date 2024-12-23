package ckhdb

import (
	"context"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
)

type Kline struct {
	PoolID      uint64          `json:"pool_id"`
	Timestamp   time.Time       `json:"timestamp"`
	Open        uint64          `json:"open"`
	High        uint64          `json:"high"`
	Low         uint64          `json:"low"`
	Close       uint64          `json:"close"`
	Volume      decimal.Decimal `json:"volume" gorm:"type:Decimal(36,18)"`
	QuoteVolume decimal.Decimal `json:"quote_volume" gorm:"type:Decimal(36,18)"`
	Trades      uint64          `json:"trades"`
	UpdateTime  time.Time       `json:"update_time"`
}

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
	KlineInterval1y  KlineInterval = "1y"
)

func (r *ClickHouseRepo) GetKline(ctx context.Context, poolID uint64, interval KlineInterval, start time.Time, end time.Time) ([]Kline, error) {
	query := fmt.Sprintf("SELECT * FROM kline_%s WHERE pool_id = %d AND timestamp >= %d AND timestamp <= %d", interval, poolID, start.Unix(), end.Unix())
	rows, err := r.DB.WithContext(ctx).Raw(query).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var klines []Kline
	for rows.Next() {
		var kline Kline
		err := rows.Scan(&kline.PoolID, &kline.Timestamp, &kline.Open, &kline.High, &kline.Low, &kline.Close, &kline.Volume, &kline.QuoteVolume, &kline.Trades, &kline.UpdateTime)
		if err != nil {
			return nil, err
		}
		klines = append(klines, kline)
	}
	return klines, nil
}


