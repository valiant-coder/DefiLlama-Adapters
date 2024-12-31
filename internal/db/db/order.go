package db

import (
	"context"
	"exapp-go/pkg/queryparams"
	"sort"
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
)

// Order represents a trading order in the DEX
type OpenOrder struct {
	TxID             string          `json:"tx_id"`
	CreatedAt        time.Time       `json:"created_at"`
	BlockNumber      uint64          `json:"block_number"`
	PoolID           uint64          `json:"pool_id" gorm:"uniqueIndex:idx_pool_id_order_id"`
	OrderID          uint64          `json:"order_id" gorm:"uniqueIndex:idx_pool_id_order_id"`
	ClientOrderID    string          `json:"order_cid"`
	Trader           string          `json:"trader" gorm:"index:idx_trader"`
	Type             OrderType       `json:"type"`
	Price            decimal.Decimal `json:"price" gorm:"type:Decimal(36,18)"`
	IsBid            bool            `json:"is_bid"`
	OriginalQuantity decimal.Decimal `json:"original_quantity" gorm:"type:Decimal(36,18)"`
	ExecutedQuantity decimal.Decimal `json:"executed_quantity" gorm:"type:Decimal(36,18)"`
	Status           OrderStatus     `json:"status"`
}

// TableName overrides the table name
func (OpenOrder) TableName() string {
	return "open_orders"
}

func (r *Repo) InsertOpenOrder(ctx context.Context, order *OpenOrder) error {
	return r.WithContext(ctx).Create(order).Error
}

func (r *Repo) DeleteOpenOrder(ctx context.Context, orderID uint64) error {
	return r.WithContext(ctx).Where("order_id = ?", orderID).Delete(&OpenOrder{}).Error
}

func (r *Repo) UpdateOpenOrder(ctx context.Context, order *OpenOrder) error {
	return r.WithContext(ctx).Save(order).Error
}

func (r *Repo) GetOpenOrder(ctx context.Context, orderID uint64) (*OpenOrder, error) {
	order := OpenOrder{}
	err := r.WithContext(ctx).Where("order_id = ?", orderID).First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *Repo) GetOpenOrders(ctx context.Context, queryParams *queryparams.QueryParams) ([]OpenOrder, int64, error) {
	queryParams.Order = "order_id desc"
	side := queryParams.Get("side")
	if side == "0" {
		queryParams.Add("is_bid", "true")
	} else if side == "1" {
		queryParams.Add("is_bid", "false")
	}
	queryParams.Del("side")
	var orders []OpenOrder
	total, err := r.Query(ctx, &orders, queryParams, "is_bid", "trader", "pool_id")
	if err != nil {
		return nil, 0, err
	}
	return orders, total, nil
}




type OrderBook struct {
	PoolID uint64      `json:"pool_id"`
	Bids   []OpenOrder `json:"bids"`
	Asks   []OpenOrder `json:"asks"`
}

func (r *Repo) GetOrderBook(ctx context.Context, poolID uint64, limit int) (*OrderBook, error) {
	orders := []OpenOrder{}
	err := r.WithContext(ctx).Where("pool_id = ? AND in_order_book = true", poolID).Find(&orders).Error
	if err != nil {
		return nil, err
	}
	book := OrderBook{
		PoolID: poolID,
	}
	for _, order := range orders {
		if order.IsBid {
			book.Bids = append(book.Bids, order)
		} else {
			book.Asks = append(book.Asks, order)
		}
	}
	sort.Slice(book.Bids, func(i, j int) bool {
		return book.Bids[i].Price.GreaterThan(book.Bids[j].Price)
	})
	sort.Slice(book.Asks, func(i, j int) bool {
		return book.Asks[i].Price.LessThan(book.Asks[j].Price)
	})
	return &book, nil
}
