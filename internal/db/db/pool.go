package db

import (
	"context"
	"time"
)

func init() {
	addMigrateFunc(func(r *Repo) error {
		return r.AutoMigrate(&Pool{})
	})
}

type Pool struct {
	PoolID             uint64     `json:"pool_id"`
	BaseSymbol         string     `json:"base_symbol"`
	BaseContract       string     `json:"base_contract"`
	BaseCoin           string     `json:"base_coin"`
	BaseCoinPrecision  uint8      `json:"base_coin_precision"`
	QuoteSymbol        string     `json:"quote_symbol"`
	QuoteContract      string     `json:"quote_contract"`
	QuoteCoin          string     `json:"quote_coin"`
	QuoteCoinPrecision uint8      `json:"quote_coin_precision"`
	Symbol             string     `json:"symbol"`
	AskingTime         time.Time  `json:"asking_time"`
	TradingTime        time.Time  `json:"trading_time"`
	MaxFluctuation     uint64     `json:"max_flct"`
	PricePrecision     uint8      `json:"price_precision"`
	TakerFeeRate       float64    `json:"taker_fee_rate"`
	MakerFeeRate       float64    `json:"maker_fee_rate"`
	Status             PoolStatus `json:"status"`
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

func (r *Repo) CreatePool(ctx context.Context, pool *Pool) error {
	return r.WithContext(ctx).Create(pool).Error
}

func (r *Repo) GetPoolBySymbol(ctx context.Context, symbol string) (*Pool, error) {
	var pool Pool
	if err := r.WithContext(ctx).Where("symbol = ?", symbol).First(&pool).Error; err != nil {
		return nil, err
	}
	return &pool, nil
}

func (r *Repo) GetPoolByID(ctx context.Context, poolID uint64) (*Pool, error) {
	var pool Pool
	if err := r.WithContext(ctx).Where("pool_id = ?", poolID).First(&pool).Error; err != nil {
		return nil, err
	}
	return &pool, nil
}

func (r *Repo) GetPoolSymbolsByIDs(ctx context.Context, poolID []uint64) (map[uint64]string, error) {
	var pools []Pool
	if err := r.WithContext(ctx).Where("pool_id IN (?)", poolID).Find(&pools).Error; err != nil {
		return nil, err
	}
	poolSymbols := make(map[uint64]string)
	for _, pool := range pools {
		poolSymbols[pool.PoolID] = pool.Symbol
	}
	return poolSymbols, nil
}
