package db

import (
	"context"
	"sort"

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
)

// Order represents a trading order in the DEX
type OpenOrder struct {
	gorm.Model
	PoolID           uint64          `json:"pool_id"`
	ClientOrderID    string          `json:"order_cid"`
	Trader           string          `json:"trader"`
	Type             OrderType       `json:"type"`
	Price            uint64          `json:"price"`
	IsBid            bool            `json:"is_bid"`
	OriginalQuantity decimal.Decimal `json:"original_quantity" gorm:"type:Decimal(36,18)"`
	ExecutedQuantity decimal.Decimal `json:"executed_quantity" gorm:"type:Decimal(36,18)"`
	Status           OrderStatus     `json:"status"`
	InOrderBook      bool            `json:"in_order_book"`
}

// TableName overrides the table name
func (OpenOrder) TableName() string {
	return "open_orders"
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
		return book.Bids[i].Price > book.Bids[j].Price
	})
	sort.Slice(book.Asks, func(i, j int) bool {
		return book.Asks[i].Price < book.Asks[j].Price
	})
	return &book, nil
}


func (r *Repo) GetOpenOrders(ctx context.Context, poolID uint64, user string) ([]OpenOrder, error) {
	orders := []OpenOrder{}
	err := r.WithContext(ctx).Where("pool_id = ? AND trader = ? AND status = ?", poolID, user, OrderStatusOpen).Find(&orders).Error
	if err != nil {
		return nil, err
	}
	return orders, nil
}



