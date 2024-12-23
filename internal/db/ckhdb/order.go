package ckhdb

import (
	"context"
	"exapp-go/pkg/queryparams"
	"time"

	"github.com/shopspring/decimal"
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
	PoolID           uint64          `json:"pool_id"`
	OrderID          uint64          `json:"order_id"`
	ClientOrderID    string          `json:"order_cid"`
	Trader           string          `json:"trader"`
	Type             OrderType       `json:"type"`
	Price            uint64          `json:"price"`
	IsBid            bool            `json:"is_bid"`
	OriginalQuantity decimal.Decimal `json:"original_quantity" gorm:"type:Decimal(36,18)"`
	ExecutedQuantity decimal.Decimal `json:"executed_quantity" gorm:"type:Decimal(36,18)"`
	Status           OrderStatus     `json:"status"`
	IsMarket         bool            `json:"is_market"`
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
}

// TableName overrides the table name
func (Order) TableName() string {
	return "orders"
}

func (r *ClickHouseRepo) QueryOrders(ctx context.Context, query *queryparams.QueryParams) ([]Order, int64, error) {
	orders := []Order{}
	total, err := r.Query(ctx, &orders, query, "pool_id", "trader")
	if err != nil {
		return nil, 0, err
	}
	return orders, total, nil
}

type OrderWithTrades struct {
	Order
	Trades []Trade `json:"trades"`
}

func (r *ClickHouseRepo) GetOrder(ctx context.Context, orderID uint64) (*OrderWithTrades, error) {
	order := Order{}
	err := r.DB.WithContext(ctx).Where("order_id = ?", orderID).First(&order).Error
	if err != nil {
		return nil, err
	}
	trades, err := r.GetTrades(ctx, orderID)
	if err != nil {
		return nil, err
	}
	return &OrderWithTrades{
		Order:  order,
		Trades: trades,
	}, nil
}




