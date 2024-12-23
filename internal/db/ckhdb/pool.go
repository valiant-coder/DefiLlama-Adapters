package ckhdb

import (
	"context"
	"time"
)

// Pool represents a trading pool in the DEX
type Pool struct {
	PoolID         uint64     `json:"pool_id"`
	BaseSymbol     string     `json:"base_symbol"`
	BaseContract   string     `json:"base_contract"`
	QuoteSymbol    string     `json:"quote_symbol"`
	QuoteContract  string     `json:"quote_contract"`
	Symbol         string     `json:"symbol"`
	AskingTime     time.Time  `json:"asking_time"`
	TradingTime    time.Time  `json:"trading_time"`
	MaxFluctuation uint64     `json:"max_flct"`
	PricePrecision uint8      `json:"price_precision"`
	TakerFeeRate   uint64     `json:"taker_fee_rate"`
	MakerFeeRate   uint64     `json:"maker_fee_rate"`
	Status         PoolStatus `json:"status"`
}

type PoolStatus uint8

const (
	PoolStatusClosed PoolStatus = 0
	PoolStatusOpen   PoolStatus = 1
)

// TableName overrides the table name
func (Pool) TableName() string {
	return "pools"
}

func (r *ClickHouseRepo) GetPool(ctx context.Context, poolID uint64) (*Pool, error) {
	pool := Pool{}
	err := r.DB.WithContext(ctx).Where("pool_id = ?", poolID).First(&pool).Error
	if err != nil {
		return nil, err
	}
	return &pool, nil
}
