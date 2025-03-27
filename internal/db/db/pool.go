package db

import (
	"context"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

func init() {
	addMigrateFunc(func(r *Repo) error {
		return r.AutoMigrate(&Pool{})
	})
}

type Pool struct {
	gorm.Model
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
	Visible            bool            `gorm:"column:visible;type:tinyint(1);default:0"`
	UpdateBlockNum     uint64          `gorm:"column:update_block_num;type:bigint(20);default:0"`
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

func (r *Repo) UpdatePool(ctx context.Context, pool *Pool) error {
	return r.WithContext(ctx).Save(pool).Error
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

func (r *Repo) GetAllPools(ctx context.Context) ([]*Pool, error) {
	var pools []*Pool
	if err := r.WithContext(ctx).Find(&pools).Error; err != nil {
		return nil, err
	}
	return pools, nil
}

func (r *Repo) GetVisiblePools(ctx context.Context) ([]*Pool, error) {
	var pools []*Pool
	if err := r.WithContext(ctx).Where("visible = 1").Find(&pools).Error; err != nil {
		return nil, err
	}
	return pools, nil
}

func (r *Repo) GetVisiblePoolTokens(ctx context.Context) (map[string]string, error) {
	pools, err := r.GetVisiblePools(ctx)
	if err != nil {
		return nil, err
	}
	poolTokens := make(map[string]string)
	for _, pool := range pools {
		poolTokens[pool.BaseCoin] = pool.BaseSymbol
		poolTokens[pool.QuoteCoin] = pool.QuoteSymbol
	}
	return poolTokens, nil
}

func (r *Repo) GetPoolMaxUpdateBlockNum(ctx context.Context) (uint64, error) {
	var blockNum *uint64
	err := r.WithContext(ctx).Model(&Pool{}).Select("COALESCE(MAX(update_block_num), 0)").Scan(&blockNum).Error
	if err != nil {
		return 0, err
	}
	return *blockNum, nil
}
