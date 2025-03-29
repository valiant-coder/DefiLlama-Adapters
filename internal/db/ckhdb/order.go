package ckhdb

import (
	"context"
	"exapp-go/pkg/queryparams"
	"strings"
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
	App                string          `gorm:"column:app;type:varchar(255)"`
	CreateTxID         string          `gorm:"column:create_tx_id;type:varchar(255)"`
	CreateBlockNum     uint64          `gorm:"column:create_block_num;type:bigint(20)"`
	CancelTxID         string          `gorm:"column:cancel_tx_id;type:varchar(255)"`
	CancelBlockNum     uint64          `gorm:"column:cancel_block_num;type:bigint(20)"`
	PoolID             uint64          `gorm:"column:pool_id;type:bigint(20)"`
	PoolSymbol         string          `gorm:"column:pool_symbol;type:varchar(255)"`
	PoolBaseCoin       string          `gorm:"column:pool_base_coin;type:varchar(255)"`
	PoolQuoteCoin      string          `gorm:"column:pool_quote_coin;type:varchar(255)"`
	OrderID            uint64          `gorm:"column:order_id;type:bigint(20)"`
	ClientOrderID      string          `gorm:"column:order_cid;type:varchar(255)"`
	Trader             string          `gorm:"column:trader;type:varchar(255)"`
	Permission         string          `gorm:"column:permission;type:varchar(255);default:'active'"`
	Type               OrderType       `gorm:"column:type;type:tinyint(4)"`
	Price              decimal.Decimal `gorm:"type:Decimal(36,18)"`
	AvgPrice           decimal.Decimal `gorm:"type:Decimal(36,18)"`
	IsBid              bool            `gorm:"column:is_bid;type:tinyint(1)"`
	OriginalQuantity   decimal.Decimal `gorm:"type:Decimal(36,18)"`
	ExecutedQuantity   decimal.Decimal `gorm:"type:Decimal(36,18)"`
	Status             OrderStatus     `gorm:"column:status;type:tinyint(4)"`
	IsMarket           bool            `gorm:"column:is_market;type:tinyint(1)"`
	CreatedAt          time.Time       `gorm:"column:created_at;type:datetime"`
	CanceledAt         time.Time       `gorm:"column:canceled_at;type:datetime"`
	BaseCoinPrecision  uint8           `gorm:"column:base_coin_precision;type:tinyint(4);default:0"`
	QuoteCoinPrecision uint8           `gorm:"column:quote_coin_precision;type:tinyint(4);default:0"`
}

// TableName overrides the table name
func (HistoryOrder) TableName() string {
	return "history_orders"
}

func (r *ClickHouseRepo) BatchInsertOrders(ctx context.Context, orders []*HistoryOrder) error {
	return r.DB.WithContext(ctx).CreateInBatches(orders, 100).Error
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
	total, err := r.Query(ctx, &orders, queryParams, "pool_id", "trader", "status", "is_bid", "permission")
	if err != nil {
		return nil, 0, err
	}
	return orders, total, nil
}

func (r *ClickHouseRepo) GetOrder(ctx context.Context, poolID uint64, orderID uint64, isBid bool) (*HistoryOrder, error) {
	order := HistoryOrder{}
	err := r.DB.WithContext(ctx).Where("pool_id = ? and order_id = ? and is_bid = ?", poolID, orderID, isBid).First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

type HistoryOrderForm struct {
	PoolID           uint64          `gorm:"column:pool_id"`
	OrderID          uint64          `gorm:"column:order_id"`
	IsBid            bool            `gorm:"column:is_bid"`
	CreatedAt        time.Time       `gorm:"column:created_at"`
	CompleteAt       time.Time       `gorm:"column:complete_at"`
	PoolBaseCoin     string          `gorm:"column:pool_base_coin"`
	Trader           string          `gorm:"column:trader"`
	Fee              decimal.Decimal `gorm:"column:fee"`
	PoolSymbol       string          `gorm:"column:pool_symbol"`
	App              string          `gorm:"column:app"`
	ExecutedQuantity decimal.Decimal `gorm:"column:executed_quantity"`
	Price            decimal.Decimal `gorm:"column:price"`
	TxIDs            string          `gorm:"column:tx_ids"`
}

func (r *ClickHouseRepo) QueryHistoryOrdersList(ctx context.Context, params *queryparams.QueryParams) ([]*HistoryOrderForm, int64, error) {
	var whereClauses []string
	var args []interface{}

	if poolBaseCoin, ok := params.CustomQuery["pool_base_coin"]; ok {
		whereClauses = append(whereClauses, "pool_base_coin = ?")
		args = append(args, poolBaseCoin[0].(string))
	}
	if poolSymbol, ok := params.CustomQuery["pool_symbol"]; ok {
		whereClauses = append(whereClauses, "pool_symbol = ?")
		args = append(args, poolSymbol[0].(string))
	}
	if app, ok := params.CustomQuery["app"]; ok {
		whereClauses = append(whereClauses, "app = ?")
		args = append(args, app[0].(string))
	}
	if trader, ok := params.CustomQuery["trader"]; ok {
		whereClauses = append(whereClauses, "trader = ?")
		args = append(args, trader[0].(string))
	}
	if startTime, ok := params.CustomQuery["start_time"]; ok {
		whereClauses = append(whereClauses, "created_at >= ?")
		args = append(args, startTime[0].(string))
	}
	if endTime, ok := params.CustomQuery["end_time"]; ok {
		whereClauses = append(whereClauses, "created_at <= ?")
		args = append(args, endTime[0].(string))
	}

	whereClause := ""
	if len(whereClauses) > 0 {
		whereClause = " AND " + strings.Join(whereClauses, " AND ")
	}

	orders := []*HistoryOrderForm{}
	query := `
    SELECT 
        pool_id,
        order_id,
        is_bid,
        created_at,
        pool_base_coin,
        trader,
        pool_symbol,
        app,
        executed_quantity,
        price,
		MIN(trade_created_at) AS complete_at,
		arrayStringConcat(groupArray(tx_id), ',') AS tx_ids,
		SUM(fee) AS fee
    FROM (
        SELECT 
            o.pool_id,
            o.order_id,
            o.is_bid,
            o.created_at,
            o.pool_base_coin,
            o.trader,
            o.pool_symbol,
            o.app,
            o.executed_quantity,
            o.price,
            t.time AS trade_created_at,
            t.tx_id,
            t.maker_fee AS fee
        FROM history_orders AS o
        LEFT JOIN trades AS t ON 
            t.maker_order_tag = concat(toString(o.pool_id), '-', toString(o.order_id), '-', if(o.is_bid = 1, '1', '0'))
        WHERE o.is_market = true
        ` + whereClause + `

        UNION ALL

        SELECT 
            o.pool_id,
            o.order_id,
            o.is_bid,
            o.created_at,
            o.pool_base_coin,
            o.trader,
            o.pool_symbol,
            o.app,
            o.executed_quantity,
            o.price,
            t.time AS trade_created_at,
            t.tx_id,
            t.taker_fee AS fee
        FROM history_orders AS o
        LEFT JOIN trades AS t ON 
            t.taker_order_tag = concat(toString(o.pool_id), '-', toString(o.order_id), '-', if(o.is_bid = 1, '1', '0'))
        WHERE o.is_market = false
        ` + whereClause + `
    ) AS combined
    GROUP BY 
        pool_id, order_id, is_bid, created_at, pool_base_coin,
        trader, pool_symbol, app, executed_quantity, price
    ORDER BY created_at DESC
	LIMIT ?, ?
    `

	values := args
	args = append(args, args...)
	args = append(args, params.Offset, params.Limit)

	err := r.DB.Raw(query, args...).Scan(&orders).Error
	if err != nil {
		return nil, 0, err
	}

	var total int64
	if whereClause != "" {
		whereClause = strings.TrimPrefix(whereClause, " AND")
	}
	err = r.DB.Table("history_orders").Where(whereClause, values...).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	return orders, total, nil
}

func (r *ClickHouseRepo) GetOrdersCoinTotal(ctx context.Context, startTime, endTime string) ([]*HistoryOrder, error) {
	var orders []*HistoryOrder

	err := r.DB.Raw(`SELECT pool_base_coin, SUM(executed_quantity) AS executed_quantity 
		FROM history_orders
		WHERE created_at BETWEEN ? AND ?
		GROUP BY pool_base_coin
		ORDER BY executed_quantity DESC`, startTime, endTime).Scan(&orders).Error
	if err != nil {
		return nil, err
	}

	return orders, nil
}

func (r *ClickHouseRepo) GetOrdersSymbolTotal(ctx context.Context, startTime, endTime string) ([]*HistoryOrder, error) {
	var orders []*HistoryOrder

	err := r.DB.Raw(`SELECT pool_symbol, SUM(executed_quantity) AS executed_quantity 
		FROM history_orders
		WHERE created_at BETWEEN ? AND ?
		GROUP BY pool_symbol
		ORDER BY executed_quantity DESC`, startTime, endTime).Scan(&orders).Error
	if err != nil {
		return nil, err
	}

	return orders, nil
}

func (r *ClickHouseRepo) GetOrdersCoinFee(ctx context.Context, startTime, endTime string) ([]*HistoryOrderForm, error) {
	var orders []*HistoryOrderForm

	err := r.DB.Raw(`SELECT base_coin as pool_base_coin, SUM(taker_fee + maker_fee) AS fee 
		FROM trades
		WHERE created_at BETWEEN ? AND ?
		GROUP BY pool_base_coin
		ORDER BY fee DESC`, startTime, endTime).Scan(&orders).Error
	if err != nil {
		return nil, err
	}

	return orders, nil
}

func (r *ClickHouseRepo) GetOrdersSymbolFee(ctx context.Context, startTime, endTime string) ([]*HistoryOrderForm, error) {
	var orders []*HistoryOrderForm

	err := r.DB.Raw(`SELECT symbol AS pool_symbol, SUM(taker_fee + maker_fee) AS fee 
		FROM trades
		WHERE created_at BETWEEN ? AND ?
		GROUP BY pool_symbol
		ORDER BY fee DESC`, startTime, endTime).Scan(&orders).Error
	if err != nil {
		return nil, err
	}

	return orders, nil
}
