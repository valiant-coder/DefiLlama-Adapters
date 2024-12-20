package db

import (
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// Trade represents a trade record in the DEX
type Trade struct {
	gorm.Model
	PoolID        uint64          `json:"pool_id"`
	Taker         string          `json:"taker" `
	Maker         string          `json:"maker"`
	MakerOrderID  uint64          `json:"maker_order_id" `
	MakerOrderCID string          `json:"maker_order_cid"`
	TakerOrderID  uint64          `json:"taker_order_id"`
	TakerOrderCID string          `json:"taker_order_cid"`
	Price         uint64          `json:"price"`
	TakerIsBid    bool            `json:"taker_is_bid"`
	BaseQuantity  decimal.Decimal `json:"base_quantity"`
	QuoteQuantity decimal.Decimal `json:"quote_quantity"`
	TakerFee      decimal.Decimal `json:"taker_fee"`
	MakerFee      decimal.Decimal `json:"maker_fee"`
	Timestamp     time.Time       `json:"time"`
}

// TableName overrides the table name
func (Trade) TableName() string {
	return "trades"
}
