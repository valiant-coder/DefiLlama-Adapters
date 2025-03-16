package marketplace

import (
	"context"
	"errors"
	"exapp-go/config"
	"exapp-go/internal/db/ckhdb"
	"exapp-go/internal/db/db"
	"exapp-go/internal/entity"
	"exapp-go/internal/errno"
	"exapp-go/pkg/queryparams"
	"strconv"
	"sync"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

var poolCache *PoolCache
var poolCacheOnce sync.Once

type PoolCache struct {
	pools     []*entity.PoolStats
	total     int64
	timestamp time.Time
}

func getPoolCache() *PoolCache {
	poolCacheOnce.Do(func() {
		poolCache = &PoolCache{
			pools:     make([]*entity.PoolStats, 0),
			total:     0,
			timestamp: time.Now(),
		}
	})
	return poolCache
}

type PoolService struct {
	repo       *db.Repo
	ckhRepo    *ckhdb.ClickHouseRepo
	cache      *PoolCache
	cacheMutex sync.RWMutex
}

func NewPoolService() *PoolService {
	return &PoolService{
		repo:    db.New(),
		ckhRepo: ckhdb.New(),
		cache:   getPoolCache(),
	}
}

func (s *PoolService) GetPools(ctx context.Context, queryParams *queryparams.QueryParams) ([]*entity.PoolStats, int64, error) {
	s.cacheMutex.RLock()
	if !s.cache.timestamp.IsZero() && time.Since(s.cache.timestamp) < 5*time.Second {
		result := make([]*entity.PoolStats, len(s.cache.pools))
		copy(result, s.cache.pools)
		total := s.cache.total
		s.cacheMutex.RUnlock()
		return result, total, nil
	}
	s.cacheMutex.RUnlock()

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
	poolsMap := make(map[uint64]bool)
	for _, pool := range pools {
		if _, ok := poolsMap[pool.PoolID]; ok {
			continue
		}
		poolsMap[pool.PoolID] = true
		result = append(result, entity.PoolStatusFromDB(pool))
	}

	s.cacheMutex.Lock()
	s.cache.pools = make([]*entity.PoolStats, len(result))
	copy(s.cache.pools, result)
	s.cache.total = total
	s.cache.timestamp = time.Now()
	s.cacheMutex.Unlock()

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
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return entity.Pool{}, nil
		}
		return entity.Pool{}, err
	}
	if !pool.Visible {
		return entity.Pool{}, errno.DefaultParamsError("not found")
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
	cfg := config.Conf()
	appTakerFeeRate := cfg.Eos.OneDex.AppTakerFeeRate
	appMakerFeeRate := cfg.Eos.OneDex.AppMakerFeeRate
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
		TakerFeeRate:       pool.TakerFeeRate + appTakerFeeRate,
		MakerFeeRate:       pool.MakerFeeRate + appMakerFeeRate,
		MinAmount:          pool.MinAmount.String(),
		Status:             uint8(pool.Status),
		PoolStats:          *entity.PoolStatusFromDB(poolStats),
	}, nil
}
