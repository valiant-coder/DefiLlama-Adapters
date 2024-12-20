package db

import (
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// OrderType represents the type of order
type OrderType uint8

const (
	OrderTypeNoRestriction OrderType = iota
	OrderTypeImmediateOrCancel
	OrderTypeFillOrKill
	OrderTypePostOnly
)

// OrderStatus represents the status of an order
type OrderStatus uint8

const (
	OrderStatusOpen OrderStatus = iota
	OrderStatusPartiallyFilled
	OrderStatusFilled
	OrderStatusCancelled
)

// Order represents a trading order in the DEX
type Order struct {
	gorm.Model
	PoolID                  uint64          `json:"pool_id"`
	ClientOrderID           string          `json:"order_cid"`
	Trader                  string          `json:"trader"`
	Type                    OrderType       `json:"type"`
	Price                   uint64          `json:"price"`
	IsBid                   bool            `json:"is_bid"`
	OriginalQuantity        decimal.Decimal `json:"original_quantity"`
	ExecutedQuantity        decimal.Decimal `json:"executed_quantity"`
	CumulativeQuoteQuantity decimal.Decimal `json:"cumulative_quote_quantity"`
	PaidFees                decimal.Decimal `json:"paid_fees"`
	Status                  OrderStatus     `json:"status"`
	IsMarket                bool            `json:"is_market"`
}

// TableName overrides the table name
func (Order) TableName() string {
	return "orders"
}
