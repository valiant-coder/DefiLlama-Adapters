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

func (r *ClickHouseRepo) QueryPoolStats(ctx context.Context, queryParams *queryparams.QueryParams) ([]PoolStats, int64, error) {
	queryParams.TableName = "pool_stats"
	queryParams.Order = "pool_id desc"
	pools := []PoolStats{}
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
WITH aggregated AS (
    SELECT
        t.pool_id,
        t.base_coin,
        t.quote_coin,
        t.symbol,
        MAX(price) as high,
        MIN(price) as low,
        COUNT(*) as trades,
        SUM(base_quantity) as volume,
        SUM(quote_quantity) as quote_volume,
        argMax(price, global_sequence) as last_price,
        argMin(price, global_sequence) as first_price
    FROM trades t
    WHERE time >= NOW() - INTERVAL 24 HOUR 
    GROUP BY pool_id, base_coin, quote_coin, symbol
)
SELECT
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
    (last_price - first_price) as change,
    (last_price / first_price - 1) as change_rate,
    now() as timestamp
FROM aggregated;
    `
	return r.DB.WithContext(ctx).Exec(query).Error
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
