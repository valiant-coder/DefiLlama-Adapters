package db

import (
	"context"
	"time"

	"github.com/shopspring/decimal"
)

func init() {
	addMigrateFunc(func(r *Repo) error {
		return r.AutoMigrate(&Pool{})
	})
}

type Pool struct {
	PoolID             uint64          `gorm:"column:pool_id;type:bigint(20);primaryKey;autoIncrement:false"`
	BaseSymbol         string          `gorm:"column:base_symbol;type:varchar(255)"`
	BaseContract       string          `gorm:"column:base_contract;type:varchar(255)"`
	BaseCoin           string          `gorm:"column:base_coin;type:varchar(255)"`
	BaseCoinPrecision  uint8           `gorm:"column:base_coin_precision;type:tinyint(4)"`
	QuoteSymbol        string          `gorm:"column:quote_symbol;type:varchar(255)"`
	QuoteContract      string          `gorm:"column:quote_contract;type:varchar(255)"`
	QuoteCoin          string          `gorm:"column:quote_coin;type:varchar(255)"`
	QuoteCoinPrecision uint8           `gorm:"column:quote_coin_precision;type:tinyint(4)"`
	Symbol             string          `gorm:"column:symbol;type:varchar(255);index:idx_symbol"`
	AskingTime         time.Time       `gorm:"column:asking_time;type:timestamp"`
	TradingTime        time.Time       `gorm:"column:trading_time;type:timestamp"`
	MaxFluctuation     uint64          `gorm:"column:max_flct;type:bigint(20)"`
	PricePrecision     uint8           `gorm:"column:price_precision;type:tinyint(4)"`
	TakerFeeRate       float64         `gorm:"column:taker_fee_rate;type:decimal(10,4)"`
	MakerFeeRate       float64         `gorm:"column:maker_fee_rate;type:decimal(10,4)"`
	Status             PoolStatus      `gorm:"column:status;type:tinyint(4)"`
	MinAmount          decimal.Decimal `gorm:"column:min_amount;type:decimal(36,18)"`
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

func (r *Repo) CreatePoolIfNotExist(ctx context.Context, pool *Pool) error {
	var existPool Pool
	if err := r.WithContext(ctx).Where("pool_id = ?", pool.PoolID).First(&existPool).Error; err != nil {
		return r.WithContext(ctx).Create(pool).Error
	}
	return nil
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
