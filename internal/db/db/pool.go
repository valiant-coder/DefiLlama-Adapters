package db

import (
	"time"

	"gorm.io/gorm"
)

// Pool represents a trading pool in the DEX
type Pool struct {
	gorm.Model
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
