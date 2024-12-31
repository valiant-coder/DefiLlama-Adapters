package marketplace

import (
	"context"
	"errors"
	"exapp-go/internal/db/ckhdb"
	"exapp-go/internal/db/db"
	"exapp-go/internal/entity"
	"exapp-go/pkg/queryparams"
	"strconv"
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

func (s *PoolService) GetPools(ctx context.Context, queryParams *queryparams.QueryParams) ([]entity.PoolStats, int64, error) {
	pools, total, err := s.ckhRepo.QueryPoolStats(ctx, queryParams)
	if err != nil {
		return nil, 0, err
	}
	var result []entity.PoolStats
	for _, pool := range pools {
		result = append(result, entity.PoolStats{
			PoolID:    pool.PoolID,
			BaseCoin:  pool.BaseCoin,
			QuoteCoin: pool.QuoteCoin,
			Symbol:    pool.Symbol,
			Change:    pool.PriceChange,
			High:      pool.High.String(),
			Low:       pool.Low.String(),
			Volume:    pool.Volume.String(),
			Turnover:  pool.QuoteVolume.String(),
			Trades:    pool.Trades,
			UpdatedAt: entity.Time(pool.Timestamp),
		})
	}
	return result, total, nil
}

func (s *PoolService) GetPool(ctx context.Context, poolSymbolOrID string) (entity.Pool, error) {
	var poolID uint64
	var err error
	if poolSymbolOrID == "" {
		return entity.Pool{}, errors.New("pool symbol or id is required")
	}
	var pool *ckhdb.Pool
	if poolID, err = strconv.ParseUint(poolSymbolOrID, 10, 64); err != nil {
		pool, err = s.ckhRepo.GetPoolBySymbol(ctx, poolSymbolOrID)
	} else {
		pool, err = s.ckhRepo.GetPoolByID(ctx, poolID)
	}
	if err != nil {
		return entity.Pool{}, err
	}
	poolStats, err := s.ckhRepo.GetPoolStats(ctx, pool.PoolID)
	if err != nil {
		return entity.Pool{}, err
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
		Status:             uint8(pool.Status),
		PoolStats: entity.PoolStats{
			PoolID:    pool.PoolID,
			Symbol:    pool.Symbol,
			BaseCoin:  pool.BaseCoin,
			QuoteCoin: pool.QuoteCoin,
			LastPrice: poolStats.LastPrice.String(),
			Change:    poolStats.PriceChange,
			High:      poolStats.High.String(),
			Low:       poolStats.Low.String(),
			Volume:    poolStats.Volume.String(),
			Turnover:  poolStats.QuoteVolume.String(),
			Trades:    poolStats.Trades,
			UpdatedAt: entity.Time(poolStats.Timestamp),
		},
	}, nil
}
