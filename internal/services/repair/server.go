package repair

import (
	"context"
	"exapp-go/config"
	"exapp-go/internal/db/db"
	"exapp-go/pkg/cdex"
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

	// 1. Get buy/sell orders in parallel
	type orderResult struct {
		orders []cdex.Order
		err    error
		isBuy  bool
	}
	orderChan := make(chan orderResult, 2)

	go func() {
		bids, err := s.client.GetOrders(ctx, poolID, true)
		orderChan <- orderResult{orders: bids, err: err, isBuy: true}
	}()

	go func() {
		asks, err := s.client.GetOrders(ctx, poolID, false)
		orderChan <- orderResult{orders: asks, err: err, isBuy: false}
	}()

	var bids, asks []cdex.Order
	for i := 0; i < 2; i++ {
		result := <-orderChan
		if result.err != nil {
			log.Printf("get orders error, poolID: %d, isBuy: %v, err: %v", poolID, result.isBuy, result.err)
			return result.err
		}
		if result.isBuy {
			bids = result.orders
		} else {
			asks = result.orders
		}
	}

	// 2. Clear old data (batch operation)
	if err := s.repo.ClearOpenOrders(ctx, poolID); err != nil {
		log.Printf("clear open orders error, poolID: %d, err: %v", poolID, err)
		return err
	}
	if err := s.repo.ClearDepthsV2(ctx, poolID); err != nil {
		log.Printf("clear depths error, poolID: %d, err: %v", poolID, err)
		return err
	}

	// 3. Prepare data for batch insert
	openOrders := make([]*db.OpenOrder, 0, len(bids)+len(asks))
	depthUpdates := make([]db.UpdateDepthParams, 0, len(bids)+len(asks))

	// Process buy orders
	for _, bid := range bids {
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
			Permission:         string(bid.Trader.Permission),
			Price:              decimal.RequireFromString(bid.Price).Shift(-int32(pool.PricePrecision)),
			IsBid:              true,
			OriginalQuantity:   decimal.New(int64(bid.Quantity.Amount), -int32(pool.BaseCoinPrecision)),
			ExecutedQuantity:   decimal.New(int64(bid.Filled.Amount), -int32(pool.BaseCoinPrecision)),
			QuoteCoinPrecision: pool.QuoteCoinPrecision,
			BaseCoinPrecision:  pool.BaseCoinPrecision,
			Status:             db.OrderStatusOpen,
		}
		openOrders = append(openOrders, openOrder)

		remainingQty := openOrder.OriginalQuantity.Sub(openOrder.ExecutedQuantity)
		depthUpdates = append(depthUpdates, db.UpdateDepthParams{
			PoolID: poolID,
			IsBuy:  true,
			Price:  openOrder.Price,
			Amount: remainingQty,
		})
	}

	// Process sell orders
	for _, ask := range asks {
		openOrder := &db.OpenOrder{
			TxID:               "",
			App:                string(ask.App),
			CreatedAt:          time.Now(),
			PoolID:             poolID,
			OrderID:            ask.ID,
			ClientOrderID:      ask.CID,
			Trader:             string(ask.Trader.Actor),
			Permission:         string(ask.Trader.Permission),
			PoolSymbol:         pool.Symbol,
			PoolBaseCoin:       pool.BaseCoin,
			PoolQuoteCoin:      pool.QuoteCoin,
			Price:              decimal.RequireFromString(ask.Price).Shift(-int32(pool.PricePrecision)),
			IsBid:              false,
			OriginalQuantity:   decimal.New(int64(ask.Quantity.Amount), -int32(pool.BaseCoinPrecision)),
			ExecutedQuantity:   decimal.New(int64(ask.Filled.Amount), -int32(pool.BaseCoinPrecision)),
			QuoteCoinPrecision: pool.QuoteCoinPrecision,
			BaseCoinPrecision:  pool.BaseCoinPrecision,
			Status:             db.OrderStatusOpen,
		}
		openOrders = append(openOrders, openOrder)

		remainingQty := openOrder.OriginalQuantity.Sub(openOrder.ExecutedQuantity)
		depthUpdates = append(depthUpdates, db.UpdateDepthParams{
			PoolID: poolID,
			IsBuy:  false,
			Price:  openOrder.Price,
			Amount: remainingQty,
		})
	}

	// 4. Batch insert orders
	if len(openOrders) > 0 {
		if err := s.repo.BatchInsertOpenOrders(ctx, openOrders); err != nil {
			log.Printf("batch insert open orders error, poolID: %d, err: %v", poolID, err)
			return err
		}
	}

	// 5. Batch update depth to Redis
	if len(depthUpdates) > 0 {
		if _, err := s.repo.UpdateDepthV2(ctx, depthUpdates); err != nil {
			log.Printf("update depth error, poolID: %d, err: %v", poolID, err)
			return err
		}
	}

	return nil
}
