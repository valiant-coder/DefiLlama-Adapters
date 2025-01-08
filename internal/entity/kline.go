package entity

import (
	"exapp-go/internal/db/ckhdb"
)

type Kline struct {
	PoolID    uint64  `json:"pool_id"`
	Timestamp Time    `json:"timestamp"`
	Open      float64 `json:"open"`
	High      float64 `json:"high"`
	Low       float64 `json:"low"`
	Close     float64 `json:"close"`
	Volume    float64 `json:"volume"`
	Turnover  float64 `json:"turnover"`
	Count     int64   `json:"count"`
}

func DbKlineToEntity(kline *ckhdb.Kline) Kline {
	return Kline{
		PoolID:    kline.PoolID,
		Timestamp: Time(kline.IntervalStart),
		Open:      kline.Open.InexactFloat64(),
		High:      kline.High.InexactFloat64(),
		Low:       kline.Low.InexactFloat64(),
		Close:     kline.Close.InexactFloat64(),
		Volume:    kline.Volume.InexactFloat64(),
		Turnover:  kline.QuoteVolume.InexactFloat64(),
		Count:     int64(kline.Trades),
	}
}
