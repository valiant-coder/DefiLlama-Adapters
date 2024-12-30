package ckhdb

import (
	"context"
	"exapp-go/pkg/queryparams"
	"time"

	"github.com/shopspring/decimal"
)

// Pool represents a trading pool in the DEX
type Pool struct {
	PoolID             uint64    `json:"pool_id"`
	BaseSymbol         string    `json:"base_symbol"`
	BaseContract       string    `json:"base_contract"`
	BaseCoin           string    `json:"base_coin"`
	BaseCoinPrecision  uint8     `json:"base_coin_precision"`
	QuoteSymbol        string    `json:"quote_symbol"`
	QuoteContract      string    `json:"quote_contract"`
	QuoteCoin          string    `json:"quote_coin"`
	QuoteCoinPrecision uint8     `json:"quote_coin_precision"`
	Symbol             string    `json:"symbol"`
	AskingTime         time.Time `json:"asking_time"`
	TradingTime        time.Time `json:"trading_time"`
	MaxFluctuation     uint64    `json:"max_flct"`
	PricePrecision     uint8     `json:"price_precision"`
	TakerFeeRate       uint64    `json:"taker_fee_rate"`
	MakerFeeRate       uint64    `json:"maker_fee_rate"`
	Status             uint8     `json:"status"`
}

type PoolStatus uint8

const (
	PoolStatusClosed PoolStatus = 0
	PoolStatusOpen   PoolStatus = 1
)

// TableName overrides the table name
func (Pool) TableName() string {
	return "pools"
}

func (r *ClickHouseRepo) GetPoolByID(ctx context.Context, poolID uint64) (*Pool, error) {
	pool := Pool{}
	err := r.DB.WithContext(ctx).Where("pool_id = ?", poolID).First(&pool).Error
	if err != nil {
		return nil, err
	}
	return &pool, nil
}

func (r *ClickHouseRepo) GetPoolBySymbol(ctx context.Context, symbol string) (*Pool, error) {
	pool := Pool{}
	err := r.DB.WithContext(ctx).Where("symbol = ?", symbol).First(&pool).Error
	if err != nil {
		return nil, err
	}
	return &pool, nil
}

type PoolStats struct {
	PoolID      uint64          `json:"pool_id"`
	PriceChange float64         `json:"price_change"`
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

type PoolStatsWithPool struct {
	PoolStats
	BaseCoin  string `json:"base_coin"`
	QuoteCoin string `json:"quote_coin"`
	Symbol    string `json:"symbol"`
}

func (r *ClickHouseRepo) QueryPoolStats(ctx context.Context, queryParams *queryparams.QueryParams) ([]PoolStatsWithPool, int64, error) {
	pools := []PoolStatsWithPool{}
	queryParams.Select = "pool_stats.*, pools.base_coin, pools.quote_coin, pools.symbol"
	queryParams.Joins = "left join pools on pool_stats.pool_id = pools.pool_id"
	total, err := r.Query(ctx, &pools, queryParams, "base_coin", "quote_coin")
	if err != nil {
		return nil, 0, err
	}
	return pools, total, nil
}

func (r *ClickHouseRepo) UpdatePoolStats(ctx context.Context) error {
	query := `
        INSERT INTO pool_stats (
            pool_id, high, low, trades, volume, quote_volume, 
            price_change, timestamp
        )
        SELECT
            t.pool_id,
            MAX(price) as high,
            MIN(price) as low,
            COUNT(*) as trades,
            SUM(base_quantity) as volume,
            SUM(quote_quantity) as quote_volume,
            ((LAST_VALUE(price) OVER w) / (FIRST_VALUE(price) OVER w) - 1)  as price_change,
            now() as timestamp
        FROM trades t
        WHERE time >= NOW() - INTERVAL 24 HOUR
        GROUP BY pool_id
        WINDOW w AS (PARTITION BY pool_id ORDER BY time)
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
