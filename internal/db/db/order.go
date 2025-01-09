package db

import (
	"context"
	"exapp-go/pkg/queryparams"
	"sort"
	"time"

	"github.com/shopspring/decimal"
)

func init() {
	addMigrateFunc(func(r *Repo) error {
		return r.AutoMigrate(&OpenOrder{})
	})
}

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
	TxID             string          `gorm:"column:tx_id;type:varchar(255)"`
	App              string          `gorm:"column:app;type:varchar(255)"`
	CreatedAt        time.Time       `gorm:"column:created_at;type:timestamp"`
	BlockNumber      uint64          `gorm:"column:block_number;type:bigint(20)"`
	PoolID           uint64          `gorm:"uniqueIndex:idx_pool_id_order_id_is_bid;column:pool_id;type:bigint(20)"`
	OrderID          uint64          `gorm:"uniqueIndex:idx_pool_id_order_id_is_bid;column:order_id;type:bigint(20)"`
	PoolSymbol       string          `gorm:"column:pool_symbol;type:varchar(255)"`
	PoolBaseCoin     string          `gorm:"column:pool_base_coin;type:varchar(255)"`
	PoolQuoteCoin    string          `gorm:"column:pool_quote_coin;type:varchar(255)"`
	ClientOrderID    string          `gorm:"column:order_cid;type:varchar(255)"`
	Trader           string          `gorm:"index:idx_trader;type:varchar(255)"`
	Type             OrderType       `gorm:"column:type;type:tinyint(4)"`
	Price            decimal.Decimal `gorm:"type:Decimal(36,18)"`
	IsBid            bool            `gorm:"column:is_bid;type:tinyint(1);uniqueIndex:idx_pool_id_order_id_is_bid"`
	OriginalQuantity decimal.Decimal `gorm:"type:Decimal(36,18)"`
	ExecutedQuantity decimal.Decimal `gorm:"type:Decimal(36,18)"`
	Status           OrderStatus     `gorm:"column:status;type:tinyint(4)"`
}

// TableName overrides the table name
func (OpenOrder) TableName() string {
	return "open_orders"
}

func (r *Repo) InsertOpenOrderIfNotExist(ctx context.Context, order *OpenOrder) error {
	return r.WithContext(ctx).Model(&OpenOrder{}).Where("pool_id = ? and order_id = ? and is_bid = ?", order.PoolID, order.OrderID, order.IsBid).FirstOrCreate(order).Error
}

func (r *Repo) DeleteOpenOrder(ctx context.Context, poolID uint64, orderID uint64, isBid bool) error {
	return r.WithContext(ctx).Where("pool_id = ? and order_id = ? and is_bid = ?", poolID, orderID, isBid).Delete(&OpenOrder{}).Error
}

func (r *Repo) UpdateOpenOrder(ctx context.Context, order *OpenOrder) error {
	return r.WithContext(ctx).Model(&OpenOrder{}).Where("pool_id = ? and order_id = ? and is_bid = ?", order.PoolID, order.OrderID, order.IsBid).Updates(order).Error
}

func (r *Repo) GetOpenOrder(ctx context.Context, poolID uint64, orderID uint64, isBid bool) (*OpenOrder, error) {
	order := OpenOrder{}
	err := r.WithContext(ctx).Where("pool_id = ? and order_id = ? and is_bid = ?", poolID, orderID, isBid).First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *Repo) GetOpenOrders(ctx context.Context, queryParams *queryparams.QueryParams) ([]OpenOrder, int64, error) {
	queryParams.Order = "created_at desc"
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


func (r *Repo) GetOpenOrderByTrader(ctx context.Context, trader string) ([]*OpenOrder, error) {
	var orders []*OpenOrder
	err := r.WithContext(ctx).Where("trader = ?", trader).Find(&orders).Error
	if err != nil {
		return nil, err
	}
	return orders, nil
}


type OrderBook struct {
	PoolID uint64      `json:"pool_id"`
	Bids   []OpenOrder `json:"bids"`
	Asks   []OpenOrder `json:"asks"`
}

func (r *Repo) GetOrderBook(ctx context.Context, poolID uint64, limit int) (*OrderBook, error) {
	orders := []OpenOrder{}
	err := r.WithContext(ctx).Where("pool_id = ?", poolID).Find(&orders).Error
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
