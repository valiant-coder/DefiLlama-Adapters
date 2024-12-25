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

// HistoryOrder represents a trading order in the DEX
type HistoryOrder struct {
	CreateTxID       string          `json:"create_tx_id"`
	CancelTxID       string          `json:"cancel_tx_id"`
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
	CreateTime       time.Time       `json:"create_time"`
	CancelTime       time.Time       `json:"cancel_time"`
}

// TableName overrides the table name
func (HistoryOrder) TableName() string {
	return "history_orders"
}


func (r *ClickHouseRepo) InsertHistoryOrder(ctx context.Context, order *HistoryOrder) error {
	return r.DB.WithContext(ctx).Create(order).Error
}


func (r *ClickHouseRepo) QueryOrders(ctx context.Context, query *queryparams.QueryParams) ([]HistoryOrder, int64, error) {
	orders := []HistoryOrder{}
	total, err := r.Query(ctx, &orders, query, "pool_id", "trader")
	if err != nil {
		return nil, 0, err
	}
	return orders, total, nil
}

type OrderWithTrades struct {
	HistoryOrder
	Trades []Trade `json:"trades"`
}

func (r *ClickHouseRepo) GetOrder(ctx context.Context, orderID uint64) (*OrderWithTrades, error) {
	order := HistoryOrder{}
	err := r.DB.WithContext(ctx).Where("order_id = ?", orderID).First(&order).Error
	if err != nil {
		return nil, err
	}
	trades, err := r.GetTrades(ctx, orderID)
	if err != nil {
		return nil, err
	}
	return &OrderWithTrades{
		HistoryOrder: order,
		Trades:       trades,
	}, nil
}
