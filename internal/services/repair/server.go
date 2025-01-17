package repair

import (
	"context"
	"exapp-go/config"
	"exapp-go/internal/db/db"
	"exapp-go/pkg/cdex"
	"fmt"
	"log"
	"time"

	"github.com/shopspring/decimal"
)

type Server struct {
	repo   *db.Repo
	client *cdex.Client
}

func NewRepairServer() *Server {
	return &Server{
		repo: db.New(),
		client: cdex.NewClient(
			config.Conf().Eos.NodeURL,
			config.Conf().Eos.CdexConfig.DexContract,
			config.Conf().Eos.CdexConfig.PoolContract,
		),
	}
}

// RepairPool repairs order and depth data for the specified pool
func (s *Server) RepairPool(ctx context.Context, poolID uint64) error {
	pool, err := s.repo.GetPoolByID(ctx, poolID)
	if err != nil {
		log.Printf("get pool error, poolID: %d, err: %v", poolID, err)
		return err
	}

	// 1. Get buy orders
	bids, err := s.client.GetOrders(ctx, poolID, true)
	if err != nil {
		log.Printf("get bids error, poolID: %d, err: %v", poolID, err)
		return err
	}

	// 2. Get sell orders
	asks, err := s.client.GetOrders(ctx, poolID, false)
	if err != nil {
		log.Printf("get asks error, poolID: %d, err: %v", poolID, err)
		return err
	}

	// 3.1 Clear old open orders
	err = s.repo.ClearOpenOrders(ctx, poolID)
	if err != nil {
		log.Printf("clear open orders error, poolID: %d, err: %v", poolID, err)
		return err
	}
	// 3. Clear old depth data
	err = s.repo.ClearDepths(ctx, poolID)
	if err != nil {
		log.Printf("clear depths error, poolID: %d, err: %v", poolID, err)
		return err
	}

	// 4. Update buy order data and depth
	for _, bid := range bids {
		// 4.1 Convert to OpenOrder
		openOrder := &db.OpenOrder{
			TxID:               "",
			App:                string(bid.App),
			CreatedAt:          time.Now(),
			PoolID:             poolID,
			OrderID:            bid.ID,
			ClientOrderID:      bid.CID,
			PoolSymbol:         pool.Symbol,
			PoolBaseCoin:       pool.BaseCoin,
			PoolQuoteCoin:      pool.QuoteCoin,
			Trader:             string(bid.Trader.Actor),
			Price:              decimal.RequireFromString(bid.Price).Shift(-int32(pool.PricePrecision)),
			IsBid:              true,
			OriginalQuantity:   decimal.New(int64(bid.Quantity.Amount), -int32(pool.QuoteCoinPrecision)),
			ExecutedQuantity:   decimal.New(int64(bid.Filled.Amount), -int32(pool.QuoteCoinPrecision)),
			QuoteCoinPrecision: pool.QuoteCoinPrecision,
			Status:             db.OrderStatusOpen,
		}

		fmt.Println(openOrder.Price.String())
		// 4.2 Insert OpenOrder
		err = s.repo.InsertOpenOrderIfNotExist(ctx, openOrder)
		if err != nil {
			log.Printf("insert open order error, poolID: %d, err: %v", poolID, err)
			return err
		}

		// 4.3 Update depth
		_, err = s.repo.UpdateDepth(ctx, []db.UpdateDepthParams{
			{
				PoolID: poolID,
				IsBuy:  true,
				Price:  openOrder.Price,
				Amount: openOrder.OriginalQuantity.Sub(openOrder.ExecutedQuantity),
			},
		})
		if err != nil {
			log.Printf("update depth error, poolID: %d, err: %v", poolID, err)
			return err
		}
	}

	// 5. Update sell order data and depth
	for _, ask := range asks {
		// 5.1 Convert to OpenOrder
		openOrder := &db.OpenOrder{
			TxID:               "",
			App:                string(ask.App),
			CreatedAt:          time.Now(),
			PoolID:             poolID,
			OrderID:            ask.ID,
			ClientOrderID:      ask.CID,
			Trader:             string(ask.Trader.Actor),
			Price:              decimal.RequireFromString(ask.Price).Shift(-int32(pool.PricePrecision)),
			IsBid:              false,
			OriginalQuantity:   decimal.New(int64(ask.Quantity.Amount), -int32(pool.QuoteCoinPrecision)),
			ExecutedQuantity:   decimal.New(int64(ask.Filled.Amount), -int32(pool.QuoteCoinPrecision)),
			QuoteCoinPrecision: pool.QuoteCoinPrecision,
			Status:             db.OrderStatusOpen,
		}

		// 5.2 Insert OpenOrder
		err = s.repo.InsertOpenOrderIfNotExist(ctx, openOrder)
		if err != nil {
			log.Printf("insert open order error, poolID: %d, err: %v", poolID, err)
			return err
		}

		// 5.3 Update depth
		_, err = s.repo.UpdateDepth(ctx, []db.UpdateDepthParams{
			{
				PoolID: poolID,
				IsBuy:  false,
				Price:  openOrder.Price,
				Amount: openOrder.OriginalQuantity.Sub(openOrder.ExecutedQuantity),
			},
		})
		if err != nil {
			log.Printf("update depth error, poolID: %d, err: %v", poolID, err)
			return err
		}
	}

	return nil
}
