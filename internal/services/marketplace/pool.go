package marketplace

import (
	"context"
	"errors"
	"exapp-go/internal/db/ckhdb"
	"exapp-go/internal/db/db"
	"exapp-go/internal/entity"
	"exapp-go/pkg/queryparams"
	"strconv"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type PoolService struct {
	repo    *db.Repo
	ckhRepo *ckhdb.ClickHouseRepo
}

func NewPoolService() *PoolService {
	return &PoolService{
		repo:    db.New(),
		ckhRepo: ckhdb.New(),
	}
}

func (s *PoolService) GetPools(ctx context.Context, queryParams *queryparams.QueryParams) ([]*entity.PoolStats, int64, error) {
	visiblePools, err := s.repo.GetVisiblePools(ctx)
	if err != nil {
		return make([]*entity.PoolStats, 0), 0, err
	}
	visiblePoolIDs := make([]string, len(visiblePools))
	for i, pool := range visiblePools {
		visiblePoolIDs[i] = strconv.FormatUint(pool.PoolID, 10)
	}
	queryParams.AddCustomQuery("pool_id in (?)", visiblePoolIDs)
	pools, total, err := s.ckhRepo.QueryPoolStats(ctx, queryParams)
	if err != nil {
		return make([]*entity.PoolStats, 0), 0, err
	}
	result := make([]*entity.PoolStats, 0, len(pools))
	for _, pool := range pools {
		result = append(result, entity.PoolStatusFromDB(pool))
	}
	return result, total, nil
}

func (s *PoolService) GetPool(ctx context.Context, poolSymbolOrID string) (entity.Pool, error) {
	var poolID uint64
	var err error
	if poolSymbolOrID == "" {
		return entity.Pool{}, errors.New("pool symbol or id is required")
	}
	var pool *db.Pool
	if poolID, err = strconv.ParseUint(poolSymbolOrID, 10, 64); err != nil {
		pool, err = s.repo.GetPoolBySymbol(ctx, poolSymbolOrID)
	} else {
		pool, err = s.repo.GetPoolByID(ctx, poolID)
	}
	if err != nil {
		return entity.Pool{}, err
	}
	poolStats, err := s.ckhRepo.GetPoolStats(ctx, pool.PoolID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			poolStats = &ckhdb.PoolStats{
				PoolID:      pool.PoolID,
				Symbol:      pool.Symbol,
				BaseCoin:    pool.BaseCoin,
				QuoteCoin:   pool.QuoteCoin,
				LastPrice:   decimal.NewFromInt(0),
				Change:      decimal.NewFromInt(0),
				ChangeRate:  0,
				High:        decimal.NewFromInt(0),
				Low:         decimal.NewFromInt(0),
				Volume:      decimal.NewFromInt(0),
				QuoteVolume: decimal.NewFromInt(0),
				Trades:      0,
				Timestamp:   time.Now(),
			}

		} else {
			return entity.Pool{}, err
		}
	}
	return entity.Pool{
		PoolID:             pool.PoolID,
		Symbol:             pool.Symbol,
		BaseCoin:           pool.BaseCoin,
		QuoteCoin:          pool.QuoteCoin,
		BaseSymbol:         pool.BaseSymbol,
		QuoteSymbol:        pool.QuoteSymbol,
		BaseContract:       pool.BaseContract,
		QuoteContract:      pool.QuoteContract,
		BaseCoinPrecision:  pool.BaseCoinPrecision,
		QuoteCoinPrecision: pool.QuoteCoinPrecision,
		AskingTime:         entity.Time(pool.AskingTime),
		TradingTime:        entity.Time(pool.TradingTime),
		MaxFluctuation:     pool.MaxFluctuation,
		PricePrecision:     pool.PricePrecision,
		TakerFeeRate:       pool.TakerFeeRate,
		MakerFeeRate:       pool.MakerFeeRate,
		MinAmount:          pool.MinAmount.String(),
		Status:             uint8(pool.Status),
		PoolStats:          *entity.PoolStatusFromDB(poolStats),
	}, nil
}
