package ckhdb

import (
	"context"
	"exapp-go/pkg/queryparams"
	"fmt"
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
	App              string          `gorm:"column:app;type:varchar(255)"`
	CreateTxID       string          `gorm:"column:create_tx_id;type:varchar(255)"`
	CreateBlockNum   uint64          `gorm:"column:create_block_num;type:bigint(20)"`
	CancelTxID       string          `gorm:"column:cancel_tx_id;type:varchar(255)"`
	CancelBlockNum   uint64          `gorm:"column:cancel_block_num;type:bigint(20)"`
	PoolID           uint64          `gorm:"column:pool_id;type:bigint(20)"`
	PoolSymbol       string          `gorm:"column:pool_symbol;type:varchar(255)"`
	PoolBaseCoin     string          `gorm:"column:pool_base_coin;type:varchar(255)"`
	PoolQuoteCoin    string          `gorm:"column:pool_quote_coin;type:varchar(255)"`
	OrderID          uint64          `gorm:"column:order_id;type:bigint(20)"`
	ClientOrderID    string          `gorm:"column:order_cid;type:varchar(255)"`
	Trader           string          `gorm:"column:trader;type:varchar(255)"`
	Type             OrderType       `gorm:"column:type;type:tinyint(4)"`
	Price            decimal.Decimal `gorm:"type:Decimal(36,18)"`
	AvgPrice         decimal.Decimal `gorm:"type:Decimal(36,18)"`
	IsBid            bool            `gorm:"column:is_bid;type:tinyint(1)"`
	OriginalQuantity decimal.Decimal `gorm:"type:Decimal(36,18)"`
	ExecutedQuantity decimal.Decimal `gorm:"type:Decimal(36,18)"`
	Status           OrderStatus     `gorm:"column:status;type:tinyint(4)"`
	IsMarket         bool            `gorm:"column:is_market;type:tinyint(1)"`
	CreatedAt        time.Time       `gorm:"column:created_at;type:datetime"`
	CanceledAt       time.Time       `gorm:"column:canceled_at;type:datetime"`
}

// TableName overrides the table name
func (HistoryOrder) TableName() string {
	return "history_orders"
}

func (r *ClickHouseRepo) InsertOrderIfNotExist(ctx context.Context, order *HistoryOrder) error {

	return r.DB.WithContext(ctx).Model(&HistoryOrder{}).Where("pool_id = ? and order_id = ? and is_bid = ?", order.PoolID, order.OrderID, order.IsBid).FirstOrCreate(order).Error
}

func (r *ClickHouseRepo) QueryHistoryOrders(ctx context.Context, queryParams *queryparams.QueryParams) ([]HistoryOrder, int64, error) {
	queryParams.TableName = "history_orders"
	queryParams.Order = "created_at desc"
	side := queryParams.Get("side")
	if side == "0" {
		queryParams.Add("is_bid", "true")
	} else if side == "1" {
		queryParams.Add("is_bid", "false")
	}
	queryParams.Del("side")

	orderType := queryParams.Get("type")
	if orderType == "0" {
		queryParams.Add("is_market", "true")
	} else if orderType == "1" {
		queryParams.Add("is_market", "false")
	}
	queryParams.Del("type")

	orders := []HistoryOrder{}
	total, err := r.Query(ctx, &orders, queryParams, "pool_id", "trader", "status", "is_bid")
	if err != nil {
		return nil, 0, err
	}
	return orders, total, nil
}

type OrderWithTrades struct {
	HistoryOrder
	Trades []Trade `json:"trades"`
}

func (r *ClickHouseRepo) GetOrder(ctx context.Context, poolID uint64, orderID uint64, isBid bool) (*OrderWithTrades, error) {
	order := HistoryOrder{}
	err := r.DB.WithContext(ctx).Where("pool_id = ? and order_id = ? and is_bid = ?", poolID, orderID, isBid).First(&order).Error
	if err != nil {
		return nil, err
	}
	var side uint8
	if isBid {
		side = 0
	} else {
		side = 1
	}
	orderTag := fmt.Sprintf("%d-%d-%d", poolID, orderID, side)
	trades, err := r.GetTrades(ctx, orderTag)
	if err != nil {
		return nil, err
	}
	return &OrderWithTrades{
		HistoryOrder: order,
		Trades:       trades,
	}, nil
}
