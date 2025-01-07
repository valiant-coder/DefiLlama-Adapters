package handler

import (
	"context"
	"encoding/json"
	"exapp-go/internal/db/db"
	"exapp-go/pkg/cdex"
	"exapp-go/pkg/hyperion"
	"fmt"
	"log"
	"time"

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

	cdexClient := cdex.NewClient(s.eosCfg.NodeURL, s.cdexCfg.DexContract, s.cdexCfg.PoolContract)
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

	askingTime, err := time.Parse(time.RFC3339, pool.AskingTime)
	if err != nil {
		return err
	}
	tradingTime, err := time.Parse(time.RFC3339, pool.TradingTime)
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

	err = s.repo.CreatePool(ctx, &db.Pool{
		PoolID:             pool.ID,
		BaseSymbol:         baseSymbol,
		BaseContract:       pool.Base.Contract,
		BaseCoin:           pool.Base.Symbol,
		BaseCoinPrecision:  basePrecision,
		QuoteSymbol:        quoteSymbol,
		QuoteContract:      pool.Quote.Contract,
		QuoteCoin:          pool.Quote.Symbol,
		QuoteCoinPrecision: quotePrecision,
		Symbol:             fmt.Sprintf("%s-%s-%s-%s", pool.Base.Contract, pool.Base.Symbol, pool.Quote.Contract, pool.Quote.Symbol),
		AskingTime:         askingTime,
		TradingTime:        tradingTime,
		MaxFluctuation:     pool.MaxFlct,
		PricePrecision:     pool.PricePrecision,
		TakerFeeRate:       takerFeeRate,
		MakerFeeRate:       makerFeeRate,
		Status:             db.PoolStatus(pool.Status),
	})
	if err != nil {
		log.Printf("failed to create db pool: %v, pool: %+v", err, pool)
		return err
	}

	return nil
}
