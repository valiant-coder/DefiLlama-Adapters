package handler

import (
	"context"
	"encoding/json"
	"exapp-go/internal/db/ckhdb"
	"exapp-go/internal/db/db"
	"exapp-go/pkg/cdex"
	"exapp-go/pkg/hyperion"
	"exapp-go/pkg/utils"
	"fmt"
	"log"
	"time"

	"github.com/shopspring/decimal"
	"github.com/spf13/cast"
)

func (s *Service) handleCreatePool(action hyperion.Action) error {
	ctx := context.Background()
	var createPool struct {
		Base struct {
			Sym      string `json:"sym"`
			Contract string `json:"contract"`
		} `json:"base"`
		Quote struct {
			Sym      string `json:"sym"`
			Contract string `json:"contract"`
		} `json:"quote"`
	}

	if err := json.Unmarshal(action.Act.Data, &createPool); err != nil {
		return err
	}

	cdexCfg := s.eosCfg.CdexConfig
	cdexClient := cdex.NewClient(s.eosCfg.NodeURL, cdexCfg.DexContract, cdexCfg.PoolContract)
	pools, err := cdexClient.GetPools(ctx)
	if err != nil {
		return err
	}

	var pool cdex.Pool
	var hasPool bool
	for _, p := range pools {
		if p.Base.Symbol == createPool.Base.Sym &&
			p.Base.Contract == createPool.Base.Contract &&
			p.Quote.Symbol == createPool.Quote.Sym &&
			p.Quote.Contract == createPool.Quote.Contract {
			pool = p
			hasPool = true
			break
		}
	}

	if !hasPool {
		return fmt.Errorf("pool not found")
	}

	baseSymbol, basePrecision := pool.Base.SymbolAndPrecision()
	quoteSymbol, quotePrecision := pool.Quote.SymbolAndPrecision()

	askingTime, err := utils.ParseTime(pool.AskingTime)
	if err != nil {
		return err
	}
	tradingTime, err := utils.ParseTime(pool.TradingTime)
	if err != nil {
		return err
	}
	var takerFeeRate, makerFeeRate float64
	if pool.TakerFeeRate == "18446744073709551615" {
		takerFeeRate = s.cdexCfg.DefaultPoolTakerFeeRate
	} else {
		takerFeeRate = cast.ToFloat64(pool.TakerFeeRate) / 10000
	}
	if pool.MakerFeeRate == "18446744073709551615" {
		makerFeeRate = s.cdexCfg.DefaultPoolMakerFeeRate
	} else {
		makerFeeRate = cast.ToFloat64(pool.MakerFeeRate) / 10000
	}

	newPool := &db.Pool{
		PoolID:             pool.ID,
		BaseSymbol:         baseSymbol,
		BaseContract:       pool.Base.Contract,
		BaseCoin:           fmt.Sprintf("%s-%s", pool.Base.Contract, baseSymbol),
		BaseCoinPrecision:  basePrecision,
		QuoteSymbol:        quoteSymbol,
		QuoteContract:      pool.Quote.Contract,
		QuoteCoin:          fmt.Sprintf("%s-%s", pool.Quote.Contract, quoteSymbol),
		QuoteCoinPrecision: quotePrecision,
		Symbol:             fmt.Sprintf("%s-%s-%s-%s", pool.Base.Contract, baseSymbol, pool.Quote.Contract, quoteSymbol),
		AskingTime:         askingTime,
		TradingTime:        tradingTime,
		MaxFluctuation:     pool.MaxFlct,
		PricePrecision:     pool.PricePrecision,
		TakerFeeRate:       takerFeeRate,
		MakerFeeRate:       makerFeeRate,
		Status:             db.PoolStatus(pool.Status),
		MinAmount:          decimal.New(int64(pool.MinAmount), -int32(basePrecision)),
		Visible:            true,
	}
	err = s.repo.CreatePoolIfNotExist(ctx, newPool)
	if err != nil {
		log.Printf("failed to create db pool: %v, pool: %+v", err, newPool)
		return err
	}

	err = s.ckhRepo.CreatePoolStats(ctx, &ckhdb.PoolStats{
		PoolID:      newPool.PoolID,
		Symbol:      newPool.Symbol,
		BaseCoin:    newPool.BaseCoin,
		QuoteCoin:   newPool.QuoteCoin,
		LastPrice:   decimal.NewFromInt(0),
		Change:      decimal.NewFromInt(0),
		ChangeRate:  0,
		High:        decimal.NewFromInt(0),
		Low:         decimal.NewFromInt(0),
		Volume:      decimal.NewFromInt(0),
		QuoteVolume: decimal.NewFromInt(0),
		Trades:      0,
		Timestamp:   time.Now(),
	})
	if err != nil {
		log.Printf("failed to create db pool stats: %v, pool: %+v", err, newPool)
		return err
	}

	return nil
}
