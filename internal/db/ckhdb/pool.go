package ckhdb

import (
	"context"
	"exapp-go/pkg/queryparams"
	"time"

	"github.com/shopspring/decimal"
)

type PoolStats struct {
	PoolID      uint64          `json:"pool_id"`
	BaseCoin    string          `json:"base_coin"`
	QuoteCoin   string          `json:"quote_coin"`
	Symbol      string          `json:"symbol"`
	LastPrice   decimal.Decimal `json:"last_price" gorm:"type:Decimal(36,18)"`
	Change      decimal.Decimal `json:"change" gorm:"type:Decimal(36,18)"`
	ChangeRate  float64         `json:"change_rate"`
	High        decimal.Decimal `json:"high" gorm:"type:Decimal(36,18)"`
	Low         decimal.Decimal `json:"low" gorm:"type:Decimal(36,18)"`
	Volume      decimal.Decimal `json:"volume" gorm:"type:Decimal(36,18)"`
	QuoteVolume decimal.Decimal `json:"quote_volume" gorm:"type:Decimal(36,18)"`
	Trades      uint64          `json:"trades"`
	Timestamp   time.Time       `json:"timestamp"`
}

func (PoolStats) TableName() string {
	return "pool_stats"
}

func (r *ClickHouseRepo) QueryPoolStats(ctx context.Context, queryParams *queryparams.QueryParams) ([]*PoolStats, int64, error) {
	queryParams.TableName = "pool_stats"
	queryParams.Order = "pool_id desc"
	pools := []*PoolStats{}
	total, err := r.Query(ctx, &pools, queryParams, "base_coin", "quote_coin")
	if err != nil {
		return nil, 0, err
	}
	return pools, total, nil
}

func (r *ClickHouseRepo) UpdatePoolStats(ctx context.Context) error {
	query := `
INSERT INTO pool_stats (
    pool_id,
    base_coin,
    quote_coin,
    symbol,
    high,
    low,
    trades,
    volume,
    quote_volume,
    last_price,
    change,
    change_rate,
    timestamp
)
WITH 
last_trade AS (
    SELECT 
        pool_id,
        base_coin,
        quote_coin,
        symbol,
        price as last_known_price
    FROM trades
    WHERE (pool_id, global_sequence) IN (
        SELECT pool_id, MAX(global_sequence)
        FROM trades
        GROUP BY pool_id
    )
),
aggregated AS (
    SELECT
        lt.pool_id,
        lt.base_coin,
        lt.quote_coin,
        lt.symbol,
        MAX(t.price) as high,
        MIN(t.price) as low,
        COUNT(t.tx_id) as trades,
        SUM(COALESCE(t.base_quantity, 0)) as volume,
        SUM(COALESCE(t.quote_quantity, 0)) as quote_volume,
        argMax(t.price, global_sequence) as last_price,
        argMin(t.price, global_sequence) as first_price,
        lt.last_known_price
    FROM last_trade lt
    LEFT JOIN trades t ON lt.pool_id = t.pool_id 
        AND t.time >= NOW() - INTERVAL 24 HOUR
    GROUP BY lt.pool_id, lt.base_coin, lt.quote_coin, lt.symbol, lt.last_known_price
)
SELECT
    pool_id,
    base_coin,
    quote_coin,
    symbol,
    COALESCE(high, last_known_price) as high,
    COALESCE(low, last_known_price) as low,
    trades,
    COALESCE(volume, 0) as volume,
    COALESCE(quote_volume, 0) as quote_volume,
    COALESCE(last_price, last_known_price) as last_price,
    CASE 
        WHEN first_price IS NOT NULL THEN (COALESCE(last_price, last_known_price) - first_price)
        ELSE 0
    END as change,
    CASE 
        WHEN first_price != 0 THEN (toFloat64(COALESCE(last_price, last_known_price)) / toFloat64(first_price) - 1)
        ELSE 0
    END as change_rate,
    now() as timestamp
FROM aggregated;
    `
	return r.DB.WithContext(ctx).Exec(query).Error
}

func (r *ClickHouseRepo) CreatePoolStats(ctx context.Context, poolStats *PoolStats) error {
	return r.DB.WithContext(ctx).Create(poolStats).Error
}

/*
OPTIMIZE TABLE pool_stats FINAL;
*/
func (r *ClickHouseRepo) OptimizePoolStats(ctx context.Context) error {
	query := `
		OPTIMIZE TABLE pool_stats FINAL;
	`
	return r.DB.WithContext(ctx).Exec(query).Error
}

func (r *ClickHouseRepo) GetPoolStats(ctx context.Context, poolID uint64) (*PoolStats, error) {
	var stats PoolStats
	err := r.DB.WithContext(ctx).
		Where("pool_id = ?", poolID).
		First(&stats).Error
	return &stats, err
}

func (r *ClickHouseRepo) ListPoolStats(ctx context.Context) ([]*PoolStats, error) {
	var pools []*PoolStats
	err := r.DB.WithContext(ctx).Find(&pools).Error
	return pools, err
}
