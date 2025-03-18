package db

import (
	"context"
	"exapp-go/pkg/queryparams"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
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
	TxID               string          `gorm:"column:tx_id;type:varchar(255)"`
	App                string          `gorm:"column:app;type:varchar(255)"`
	CreatedAt          time.Time       `gorm:"column:created_at;type:timestamp;index:idx_created_at;index:idx_pool_created_at"`
	BlockNumber        uint64          `gorm:"column:block_number;type:bigint(20);index:idx_block_number"`
	OrderID            uint64          `gorm:"uniqueIndex:idx_order_id_pool_id_is_bid;column:order_id;type:bigint(20)"`
	PoolID             uint64          `gorm:"uniqueIndex:idx_order_id_pool_id_is_bid;index:idx_pool_id_is_bid;column:pool_id;type:bigint(20);index:idx_pool_created_at"`
	PoolSymbol         string          `gorm:"column:pool_symbol;type:varchar(255)"`
	PoolBaseCoin       string          `gorm:"column:pool_base_coin;type:varchar(255)"`
	PoolQuoteCoin      string          `gorm:"column:pool_quote_coin;type:varchar(255)"`
	ClientOrderID      string          `gorm:"column:order_cid;type:varchar(255)"`
	Trader             string          `gorm:"index:idx_trader;type:varchar(255)"`
	Permission         string          `gorm:"column:permission;type:varchar(255);default:'active'"`
	Type               OrderType       `gorm:"column:type;type:tinyint(4)"`
	Price              decimal.Decimal `gorm:"type:Decimal(36,18)"`
	IsBid              bool            `gorm:"column:is_bid;type:tinyint(1);uniqueIndex:idx_order_id_pool_id_is_bid;index:idx_pool_id_is_bid"`
	OriginalQuantity   decimal.Decimal `gorm:"type:Decimal(36,18)"`
	ExecutedQuantity   decimal.Decimal `gorm:"type:Decimal(36,18)"`
	QuoteCoinPrecision uint8           `gorm:"column:quote_coin_precision;type:tinyint(4)"`
	BaseCoinPrecision  uint8           `gorm:"column:base_coin_precision;type:tinyint(4);default:0"`
	Status             OrderStatus     `gorm:"column:status;type:tinyint(4)"`
}

// TableName overrides the table name
func (OpenOrder) TableName() string {
	return "open_orders"
}

func (o *OpenOrder) OrderTag() string {
	var side = "1"
	if !o.IsBid {
		side = "0"
	}
	return fmt.Sprintf("%d-%d-%s", o.PoolID, o.OrderID, side)
}

func (r *Repo) InsertOpenOrder(ctx context.Context, order *OpenOrder) error {
	return r.WithContext(ctx).Create(order).Error
}

func (r *Repo) DeleteOpenOrder(ctx context.Context, poolID uint64, orderID uint64, isBid bool) error {
	return r.WithContext(ctx).Where("order_id = ? and pool_id = ? and is_bid = ?", orderID, poolID, isBid).Delete(&OpenOrder{}).Error
}

func (r *Repo) UpdateOpenOrder(ctx context.Context, order *OpenOrder) error {
	return r.WithContext(ctx).Model(&OpenOrder{}).Where("order_id = ? and pool_id = ? and is_bid = ?", order.OrderID, order.PoolID, order.IsBid).Updates(order).Error
}

func (r *Repo) GetOpenOrder(ctx context.Context, poolID uint64, orderID uint64, isBid bool) (*OpenOrder, error) {
	order := OpenOrder{}
	err := r.WithContext(ctx).Where("order_id = ? and pool_id = ? and is_bid = ?", orderID, poolID, isBid).First(&order).Error
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
	total, err := r.Query(ctx, &orders, queryParams, "is_bid", "trader", "pool_id", "permission")
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

func (r *Repo) GetOpenOrderMaxBlockNumber(ctx context.Context) (uint64, error) {
	var blockNumber *uint64
	err := r.WithContext(ctx).Model(&OpenOrder{}).Select("COALESCE(MAX(block_number), 0)").Scan(&blockNumber).Error
	if err != nil {
		return 0, err
	}
	if blockNumber == nil {
		return 0, nil
	}
	return *blockNumber, nil
}

func (r *Repo) GetOpenOrdersByPriceRange(ctx context.Context, poolID uint64, isBid bool, minPrice, maxPrice decimal.Decimal) ([]*OpenOrder, error) {
	var orders []*OpenOrder
	result := r.DB.WithContext(ctx).Where(
		"pool_id = ? AND is_bid = ? AND price > ? AND price < ?",
		poolID, isBid, minPrice, maxPrice,
	).Find(&orders)

	if result.Error != nil {
		return nil, fmt.Errorf("select open orders error: %w", result.Error)
	}
	return orders, nil
}

type OrderBook struct {
	PoolID uint64      `json:"pool_id"`
	Bids   []OpenOrder `json:"bids"`
	Asks   []OpenOrder `json:"asks"`
}

func (r *Repo) GetOrderBook(ctx context.Context, poolID uint64, limit int) (*OrderBook, error) {
	book := &OrderBook{PoolID: poolID}

	err := r.WithContext(ctx).
		Where("pool_id = ? AND is_bid = ?", poolID, true).
		Order("price DESC").
		Limit(limit).
		Find(&book.Bids).Error
	if err != nil {
		return nil, err
	}

	err = r.WithContext(ctx).
		Where("pool_id = ? AND is_bid = ?", poolID, false).
		Order("price ASC").
		Limit(limit).
		Find(&book.Asks).Error
	if err != nil {
		return nil, err
	}

	return book, nil
}

func (r *Repo) ClearOpenOrders(ctx context.Context, poolID uint64) error {
	return r.WithContext(ctx).Where("pool_id = ?", poolID).Delete(&OpenOrder{}).Error
}

func (r *Repo) BatchInsertOpenOrders(ctx context.Context, orders []*OpenOrder) error {
	if len(orders) == 0 {
		return nil
	}
	return r.WithContext(ctx).CreateInBatches(orders, 100).Error
}

func (r *Repo) BatchUpdateOpenOrders(ctx context.Context, orders []*OpenOrder) error {
	if len(orders) == 0 {
		return nil
	}

	tx := r.WithContext(ctx).Begin()
	if tx.Error != nil {
		return tx.Error
	}

	for _, order := range orders {
		err := tx.Model(&OpenOrder{}).
			Where("order_id = ? AND pool_id = ? AND is_bid = ?", order.OrderID, order.PoolID, order.IsBid).
			Updates(map[string]interface{}{
				"executed_quantity": order.ExecutedQuantity,
				"status":            order.Status,
			}).Error
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit().Error
}

func (r *Repo) BatchDeleteOpenOrders(ctx context.Context, orders []*OpenOrder) error {
	if len(orders) == 0 {
		return nil
	}

	tx := r.WithContext(ctx).Begin()
	if tx.Error != nil {
		return tx.Error
	}

	for _, order := range orders {
		err := tx.Where("order_id = ? AND pool_id = ? AND is_bid = ?", order.OrderID, order.PoolID, order.IsBid).Delete(&OpenOrder{}).Error
		if err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit().Error
}

// Get Redis key for unread orders
func getUnreadOrdersKey(trader string) string {
	return fmt.Sprintf("unread_filled_orders:%s", trader)
}

// Add unread order
func (r *Repo) AddUnreadOrder(ctx context.Context, trader string, orderID string) error {
	key := getUnreadOrdersKey(trader)
	return r.Redis().SAdd(ctx, key, orderID).Err()
}

// Mark order as read
func (r *Repo) MarkOrderAsRead(ctx context.Context, trader string, orderID string) error {
	key := getUnreadOrdersKey(trader)
	return r.Redis().SRem(ctx, key, orderID).Err()
}

// Check if order is unread
func (r *Repo) IsOrderUnread(ctx context.Context, trader string, orderID string) (bool, error) {
	key := getUnreadOrdersKey(trader)
	return r.Redis().SIsMember(ctx, key, orderID).Result()
}

// Check if user has any unread orders
func (r *Repo) HasUnreadOrders(ctx context.Context, trader string) (bool, error) {
	key := getUnreadOrdersKey(trader)
	count, err := r.Redis().SCard(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return false, nil
		}
		return false, err
	}
	return count > 0, nil
}

// ClearUnreadOrders clears all unread orders for a trader
func (r *Repo) ClearUnreadOrders(ctx context.Context, trader string) error {
	key := getUnreadOrdersKey(trader)
	return r.Redis().Del(ctx, key).Err()
}
