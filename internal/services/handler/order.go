package handler

import (
	"context"
	"encoding/json"
	"exapp-go/internal/db/ckhdb"
	"exapp-go/internal/db/db"
	"exapp-go/pkg/hyperion"
	"exapp-go/pkg/utils"
	"fmt"
	"log"
	"time"

	eosgo "github.com/eoscanada/eos-go"
	"github.com/shopspring/decimal"
	"github.com/spf13/cast"
)

func eosAssetToDecimal(a string) (decimal.Decimal, error) {
	asset, err := eosgo.NewAssetFromString(a)
	if err != nil {
		return decimal.Decimal{}, err
	}
	return decimal.New(int64(asset.Amount), -int32(asset.Symbol.Precision)), nil
}

func (s *Service) handleCreateOrder(action hyperion.Action) error {
	start := time.Now()
	defer func() {
		log.Printf("[Performance Log] handleCreateOrder total time: %v", time.Since(start))
	}()

	var newOrder struct {
		EV struct {
			App      string `json:"app"`
			PoolID   string `json:"pool_id"`
			OrderID  string `json:"order_id"`
			OrderCID string `json:"order_cid"`
			IsBid    bool   `json:"is_bid"`
			Trader   struct {
				Actor      string `json:"actor"`
				Permission string `json:"permission"`
			} `json:"trader"`
			ExecutedQuantity string `json:"executed_quantity"`
			PlacedQuantity   string `json:"placed_quantity"`
			Price            string `json:"price"`
			Status           uint8  `json:"status"`
			IsMarket         bool   `json:"is_market"`
			IsInserted       bool   `json:"is_inserted"`
			Time             string `json:"time"`
		} `json:"ev"`
	}

	if err := json.Unmarshal(action.Act.Data, &newOrder); err != nil {
		log.Printf("unmarshal create order data failed: %v", err)
		return nil
	}

	log.Printf("newOrder: %v-%v-%v-%v,global_sequence: %v", newOrder.EV.PoolID, newOrder.EV.OrderID, newOrder.EV.IsBid, newOrder.EV.IsInserted, action.GlobalSequence)

	ctx := context.Background()
	poolID := cast.ToUint64(newOrder.EV.PoolID)
	orderID := cast.ToUint64(newOrder.EV.OrderID)

	var err error
	pool, ok := s.poolCache[poolID]
	if !ok {
		pool, err = s.repo.GetPoolByID(ctx, poolID)
		if err != nil {
			log.Printf("get pool failed: %v", err)
			return nil
		}
		s.poolCache[poolID] = pool
	}

	placedQuantity, err := eosAssetToDecimal(newOrder.EV.PlacedQuantity)
	if err != nil {
		log.Printf("new asset from string failed: %v", err)
		return nil
	}

	executedQuantity, err := eosAssetToDecimal(newOrder.EV.ExecutedQuantity)
	if err != nil {
		log.Printf("new asset from string failed: %v", err)
		return nil
	}

	originalQuantity := placedQuantity.Add(executedQuantity)

	createTime, err := utils.ParseTime(newOrder.EV.Time)
	if err != nil {
		log.Printf("parse action time failed: %v", err)
		return nil
	}

	if newOrder.EV.IsInserted {
		bufferStart := time.Now()
		order := db.OpenOrder{
			App:                newOrder.EV.App,
			TxID:               action.TrxID,
			CreatedAt:          createTime,
			BlockNumber:        action.BlockNum,
			OrderID:            orderID,
			PoolID:             poolID,
			PoolSymbol:         pool.Symbol,
			PoolBaseCoin:       pool.BaseCoin,
			PoolQuoteCoin:      pool.QuoteCoin,
			ClientOrderID:      newOrder.EV.OrderCID,
			Trader:             newOrder.EV.Trader.Actor,
			Price:              decimal.New(cast.ToInt64(newOrder.EV.Price), -int32(pool.PricePrecision)),
			IsBid:              newOrder.EV.IsBid,
			OriginalQuantity:   originalQuantity,
			ExecutedQuantity:   executedQuantity,
			QuoteCoinPrecision: pool.QuoteCoinPrecision,
			BaseCoinPrecision:  pool.BaseCoinPrecision,
			Status:             db.OrderStatus(newOrder.EV.Status),
		}
		s.openOrderBuffer.Add(&order)
		log.Printf("[Performance Log] openOrderBuffer add time: %v", time.Since(bufferStart))

		depthStart := time.Now()
		s.depthBuffer.Add(db.UpdateDepthParams{
			PoolID: poolID,
			UniqID: cast.ToString(action.GlobalSequence),
			Price:  order.Price,
			Amount: placedQuantity,
			IsBuy:  newOrder.EV.IsBid,
		})
		log.Printf("[Performance Log] depthBuffer add time: %v", time.Since(depthStart))

	} else {
		tradeStart := time.Now()
		var avgPrice, price decimal.Decimal
		if executedQuantity.GreaterThan(decimal.Zero) {
			var orderTag string
			if newOrder.EV.IsBid {
				orderTag = fmt.Sprintf("%d-%d-%d", poolID, orderID, 0)
			} else {
				orderTag = fmt.Sprintf("%d-%d-%d", poolID, orderID, 1)
			}

			var trades []ckhdb.Trade
			if s.tradeCache != nil {
				if cachedTrades, ok := s.tradeCache[orderTag]; ok {
					trades = make([]ckhdb.Trade, len(cachedTrades))
					for i, t := range cachedTrades {
						trades[i] = *t
					}
				}
				if len(trades) != 0 {
					delete(s.tradeCache, orderTag)
				}

			}

			if len(trades) == 0 {
				var err error
				trades, err = s.ckhRepo.GetTrades(ctx, orderTag)
				if err != nil {
					log.Printf("get trades failed: %v", err)
					return nil
				}
				log.Printf("[Performance Log] get trades time: %v", time.Since(tradeStart))
			}

			var totalQuoteQuantity, totalBaseQuantity decimal.Decimal
			for _, trade := range trades {
				totalQuoteQuantity = totalQuoteQuantity.Add(trade.QuoteQuantity)
				totalBaseQuantity = totalBaseQuantity.Add(trade.BaseQuantity)
			}
			avgPrice = totalQuoteQuantity.Div(totalBaseQuantity).Round(int32(pool.PricePrecision))
			price = decimal.New(cast.ToInt64(newOrder.EV.Price), -int32(pool.PricePrecision))
			if newOrder.EV.IsMarket {
				originalQuantity = totalBaseQuantity
				executedQuantity = totalBaseQuantity
			}
		} else {
			avgPrice = decimal.New(cast.ToInt64(newOrder.EV.Price), -int32(pool.PricePrecision))
			price = avgPrice
		}
		order := ckhdb.HistoryOrder{
			App:                newOrder.EV.App,
			CreateTxID:         action.TrxID,
			CreateBlockNum:     action.BlockNum,
			PoolID:             poolID,
			PoolSymbol:         pool.Symbol,
			PoolBaseCoin:       pool.BaseCoin,
			PoolQuoteCoin:      pool.QuoteCoin,
			OrderID:            orderID,
			ClientOrderID:      newOrder.EV.OrderCID,
			Trader:             newOrder.EV.Trader.Actor,
			Price:              price,
			AvgPrice:           avgPrice,
			IsBid:              newOrder.EV.IsBid,
			OriginalQuantity:   originalQuantity,
			ExecutedQuantity:   executedQuantity,
			Status:             ckhdb.OrderStatus(newOrder.EV.Status),
			IsMarket:           newOrder.EV.IsMarket,
			CreatedAt:          createTime,
			BaseCoinPrecision:  pool.BaseCoinPrecision,
			QuoteCoinPrecision: pool.QuoteCoinPrecision,
		}
		bufferStart := time.Now()
		s.historyOrderBuffer.Add(&order)
		log.Printf("[Performance Log] orderBuffer add time: %v", time.Since(bufferStart))

	}

	go s.updateUserTokenBalance(newOrder.EV.Trader.Actor)
	go s.publisher.PublishOrderUpdate(newOrder.EV.Trader.Actor, fmt.Sprintf("%d-%d-%s", poolID, orderID, map[bool]string{true: "0", false: "1"}[newOrder.EV.IsBid]))
	return nil
}

func (s *Service) handleMatchOrder(action hyperion.Action) error {
	start := time.Now()
	defer func() {
		log.Printf("[Performance Log] handleMatchOrder total time: %v", time.Since(start))
	}()

	var data struct {
		EV struct {
			MakerApp string `json:"maker_app"`
			TakerApp string `json:"taker_app"`
			PoolID   string `json:"pool_id"`
			Taker    struct {
				Actor      string `json:"actor"`
				Permission string `json:"permission"`
			} `json:"taker"`
			Maker struct {
				Actor      string `json:"actor"`
				Permission string `json:"permission"`
			} `json:"maker"`
			MakerOrderID  string `json:"maker_order_id"`
			MakerOrderCID string `json:"maker_order_cid"`
			TakerOrderID  string `json:"taker_order_id"`
			TakerOrderCID string `json:"taker_order_cid"`
			Price         string `json:"price"`
			TakerIsBid    bool   `json:"taker_is_bid"`
			BaseQuantity  string `json:"base_quantity"`
			QuoteQuantity string `json:"quote_quantity"`
			TakerFee      struct {
				BaseFee string `json:"base_fee"`
				AppFee  string `json:"app_fee"`
			} `json:"taker_fee"`
			MakerFee struct {
				BaseFee string `json:"base_fee"`
				AppFee  string `json:"app_fee"`
			} `json:"maker_fee"`
			Time         string `json:"time"`
			MakerRemoved bool   `json:"maker_removed"`
			MakerStatus  uint8  `json:"maker_status"`
		} `json:"ev"`
	}
	if err := json.Unmarshal(action.Act.Data, &data); err != nil {
		log.Printf("unmarshal match order data failed: %v", err)
		return nil
	}

	log.Printf("match taker order: %v-%v-%v,global_sequence: %v", data.EV.PoolID, data.EV.TakerOrderID, data.EV.TakerIsBid, action.GlobalSequence)
	log.Printf("match maker order: %v-%v-%v-%v,global_sequence: %v", data.EV.PoolID, data.EV.MakerOrderID, !data.EV.TakerIsBid, data.EV.MakerRemoved, action.GlobalSequence)

	ctx := context.Background()
	var err error
	poolID := cast.ToUint64(data.EV.PoolID)
	pool, ok := s.poolCache[poolID]
	if !ok {
		pool, err = s.repo.GetPoolByID(ctx, poolID)
		if err != nil {
			log.Printf("get pool failed: %v", err)
			return nil
		}
		s.poolCache[poolID] = pool
	}

	baseQuantity, err := eosAssetToDecimal(data.EV.BaseQuantity)
	if err != nil {
		log.Printf("new asset from string failed: %v", err)
		return nil
	}
	quoteQuantity, err := eosAssetToDecimal(data.EV.QuoteQuantity)
	if err != nil {
		log.Printf("new asset from string failed: %v", err)
		return nil
	}
	takerFee, err := eosAssetToDecimal(data.EV.TakerFee.BaseFee)
	if err != nil {
		log.Printf("new asset from string failed: %v", err)
		return nil
	}
	makerFee, err := eosAssetToDecimal(data.EV.MakerFee.BaseFee)
	if err != nil {
		log.Printf("new asset from string failed: %v", err)
		return nil
	}
	takerAppFee, err := eosAssetToDecimal(data.EV.TakerFee.AppFee)
	if err != nil {
		log.Printf("new asset from string failed: %v", err)
		return nil
	}
	makerAppFee, err := eosAssetToDecimal(data.EV.MakerFee.AppFee)
	if err != nil {
		log.Printf("new asset from string failed: %v", err)
		return nil
	}

	tradeTime, err := utils.ParseTime(data.EV.Time)
	if err != nil {
		log.Printf("parse action time failed: %v", err)
		return nil
	}

	trade := ckhdb.Trade{
		TxID:           action.TrxID,
		PoolID:         poolID,
		BaseCoin:       pool.BaseCoin,
		QuoteCoin:      pool.QuoteCoin,
		Symbol:         pool.Symbol,
		Price:          decimal.New(int64(cast.ToInt64(data.EV.Price)), -int32(pool.PricePrecision)),
		Time:           tradeTime,
		BlockNumber:    action.BlockNum,
		GlobalSequence: action.GlobalSequence,
		Taker:          data.EV.Taker.Actor,
		Maker:          data.EV.Maker.Actor,
		MakerOrderID:   cast.ToUint64(data.EV.MakerOrderID),
		MakerOrderCID:  data.EV.MakerOrderCID,
		TakerOrderID:   cast.ToUint64(data.EV.TakerOrderID),
		TakerOrderCID:  data.EV.TakerOrderCID,
		MakerApp:       data.EV.MakerApp,
		TakerApp:       data.EV.TakerApp,
		BaseQuantity:   baseQuantity,
		QuoteQuantity:  quoteQuantity,
		TakerFee:       takerFee,
		MakerFee:       makerFee,
		TakerAppFee:    takerAppFee,
		MakerAppFee:    makerAppFee,
		TakerIsBid:     data.EV.TakerIsBid,
		CreatedAt:      time.Now(),
	}
	var makerSide, takerSide uint8
	if trade.TakerIsBid {
		takerSide = 0
		makerSide = 1
	} else {
		takerSide = 1
		makerSide = 0
	}
	trade.MakerOrderTag = fmt.Sprintf("%d-%d-%d", trade.PoolID, trade.MakerOrderID, makerSide)
	trade.TakerOrderTag = fmt.Sprintf("%d-%d-%d", trade.PoolID, trade.TakerOrderID, takerSide)
	tradeStart := time.Now()
	err = s.newTrade(ctx, &trade)
	if err != nil {
		log.Printf("new trade failed: %v", err)
	}
	log.Printf("[Performance Log] create new trade time: %v", time.Since(tradeStart))

	depthStart := time.Now()
	if data.EV.TakerIsBid {
		s.depthBuffer.Add(db.UpdateDepthParams{
			PoolID: poolID,
			UniqID: cast.ToString(action.GlobalSequence),
			Price:  trade.Price,
			Amount: baseQuantity.Neg(),
			IsBuy:  false,
		})
	} else {
		s.depthBuffer.Add(db.UpdateDepthParams{
			PoolID: poolID,
			UniqID: cast.ToString(action.GlobalSequence),
			Price:  trade.Price,
			Amount: baseQuantity.Neg(),
			IsBuy:  true,
		})
	}
	log.Printf("[Performance Log] depthBuffer update time: %v", time.Since(depthStart))

	orderStart := time.Now()
	makerOrder, err := s.openOrderBuffer.Get(poolID, cast.ToUint64(data.EV.MakerOrderID), !data.EV.TakerIsBid)
	if err != nil {
		log.Printf("get maker order failed: %v", err)
		return nil
	}
	log.Printf("[Performance Log] get maker order time: %v", time.Since(orderStart))

	makerOrder.ExecutedQuantity = makerOrder.ExecutedQuantity.Add(baseQuantity)
	makerOrder.Status = db.OrderStatus(data.EV.MakerStatus)
	if makerOrder.ExecutedQuantity.GreaterThan(makerOrder.OriginalQuantity) {
		errMsg := fmt.Sprintf("Invalid order execution: TxID=%v, OrderTxID=%v, Original=%v, Executed=%v",
			trade.TxID,
			makerOrder.TxID,
			makerOrder.OriginalQuantity,
			makerOrder.ExecutedQuantity)
		log.Printf(errMsg)
		makerOrder.ExecutedQuantity = makerOrder.OriginalQuantity
	}

	if !data.EV.MakerRemoved {
		s.openOrderBuffer.Update(makerOrder)
	} else {
		s.openOrderBuffer.Delete(poolID, cast.ToUint64(data.EV.MakerOrderID), makerOrder.IsBid)

		historyOrder := ckhdb.HistoryOrder{
			App:                makerOrder.App,
			PoolID:             makerOrder.PoolID,
			PoolSymbol:         pool.Symbol,
			PoolBaseCoin:       pool.BaseCoin,
			PoolQuoteCoin:      pool.QuoteCoin,
			OrderID:            makerOrder.OrderID,
			ClientOrderID:      makerOrder.ClientOrderID,
			Trader:             makerOrder.Trader,
			Price:              makerOrder.Price,
			AvgPrice:           makerOrder.Price,
			IsBid:              makerOrder.IsBid,
			OriginalQuantity:   makerOrder.OriginalQuantity,
			ExecutedQuantity:   makerOrder.ExecutedQuantity,
			Status:             ckhdb.OrderStatus(makerOrder.Status),
			IsMarket:           false,
			CreateTxID:         makerOrder.TxID,
			CreateBlockNum:     makerOrder.BlockNumber,
			CreatedAt:          makerOrder.CreatedAt,
			BaseCoinPrecision:  pool.BaseCoinPrecision,
			QuoteCoinPrecision: pool.QuoteCoinPrecision,
		}
		s.historyOrderBuffer.Add(&historyOrder)

	}

	go s.publisher.PublishOrderUpdate(data.EV.Maker.Actor, fmt.Sprintf("%d-%d-%s", poolID, cast.ToUint64(data.EV.MakerOrderID), map[bool]string{true: "0", false: "1"}[!data.EV.TakerIsBid]))
	go s.publisher.PublishOrderUpdate(data.EV.Taker.Actor, fmt.Sprintf("%d-%d-%s", poolID, cast.ToUint64(data.EV.TakerOrderID), map[bool]string{true: "0", false: "1"}[data.EV.TakerIsBid]))

	return nil
}

func (s *Service) handleCancelOrder(action hyperion.Action) error {
	start := time.Now()
	defer func() {
		log.Printf("[Performance Log] handleCancelOrder total time: %v", time.Since(start))
	}()

	var data struct {
		EV struct {
			App      string `json:"app"`
			PoolID   string `json:"pool_id"`
			OrderID  string `json:"order_id"`
			OrderCID string `json:"order_cid"`
			IsBid    bool   `json:"is_bid"`
			Price    string `json:"price"`
			Trader   struct {
				Actor      string `json:"actor"`
				Permission string `json:"permission"`
			} `json:"trader"`
			OriginalQuantity     string `json:"original_quantity"`
			CanceledBaseQuantity string `json:"canceled_base_quantity"`
			Time                 string `json:"time"`
		} `json:"ev"`
	}

	if err := json.Unmarshal(action.Act.Data, &data); err != nil {
		log.Printf("unmarshal cancel order data failed: %v", err)
		return nil
	}

	log.Printf("cancel order: %v-%v-%v,global_sequence: %v", data.EV.PoolID, data.EV.OrderID, data.EV.IsBid, action.GlobalSequence)

	ctx := context.Background()
	canceledQuantity, err := eosAssetToDecimal(data.EV.CanceledBaseQuantity)
	if err != nil {
		log.Printf("new asset from string failed: %v", err)
		return nil
	}

	poolID := cast.ToUint64(data.EV.PoolID)
	pool, ok := s.poolCache[poolID]
	if !ok {
		pool, err = s.repo.GetPoolByID(ctx, poolID)
		if err != nil {
			log.Printf("get pool failed: %v", err)
			return nil
		}
		s.poolCache[poolID] = pool
	}

	price := decimal.New(cast.ToInt64(data.EV.Price), -int32(pool.PricePrecision))

	depthStart := time.Now()
	if data.EV.IsBid {
		s.depthBuffer.Add(db.UpdateDepthParams{
			PoolID: poolID,
			UniqID: cast.ToString(action.GlobalSequence),
			Price:  price,
			Amount: canceledQuantity.Neg(),
			IsBuy:  true,
		})
	} else {
		s.depthBuffer.Add(db.UpdateDepthParams{
			PoolID: poolID,
			UniqID: cast.ToString(action.GlobalSequence),
			Price:  price,
			Amount: canceledQuantity.Neg(),
			IsBuy:  false,
		})
	}
	log.Printf("[Performance Log] depthBuffer update time: %v", time.Since(depthStart))

	orderID := cast.ToUint64(data.EV.OrderID)
	orderStart := time.Now()
	order, err := s.openOrderBuffer.Get(poolID, orderID, data.EV.IsBid)
	if err != nil {
		log.Printf("get open order failed: %v", err)
		return nil
	}
	log.Printf("[Performance Log] get order time: %v", time.Since(orderStart))

	bufferStart := time.Now()
	s.openOrderBuffer.Delete(poolID, orderID, order.IsBid)
	log.Printf("[Performance Log] delete order time: %v", time.Since(bufferStart))

	canceledTime, err := utils.ParseTime(data.EV.Time)
	if err != nil {
		log.Printf("parse action time failed: %v", err)
		return nil
	}
	var status ckhdb.OrderStatus
	if order.Status == db.OrderStatusPartiallyFilled {
		status = ckhdb.OrderStatusPartiallyFilled
	} else {
		status = ckhdb.OrderStatusCancelled
	}

	historyStart := time.Now()
	historyOrder := ckhdb.HistoryOrder{
		App:                order.App,
		PoolID:             order.PoolID,
		PoolSymbol:         order.PoolSymbol,
		PoolBaseCoin:       order.PoolBaseCoin,
		PoolQuoteCoin:      order.PoolQuoteCoin,
		OrderID:            order.OrderID,
		ClientOrderID:      order.ClientOrderID,
		Trader:             order.Trader,
		Price:              order.Price,
		AvgPrice:           order.Price,
		IsBid:              order.IsBid,
		OriginalQuantity:   order.OriginalQuantity,
		ExecutedQuantity:   order.ExecutedQuantity,
		Status:             status,
		IsMarket:           false,
		CreateTxID:         order.TxID,
		CreatedAt:          order.CreatedAt,
		CreateBlockNum:     order.BlockNumber,
		CancelTxID:         action.TrxID,
		CancelBlockNum:     action.BlockNum,
		CanceledAt:         canceledTime,
		Type:               ckhdb.OrderType(order.Type),
		BaseCoinPrecision:  order.BaseCoinPrecision,
		QuoteCoinPrecision: order.QuoteCoinPrecision,
	}
	s.historyOrderBuffer.Add(&historyOrder)
	log.Printf("[Performance Log] add history order time: %v", time.Since(historyStart))

	go s.updateUserTokenBalance(data.EV.Trader.Actor)
	go s.publisher.PublishOrderUpdate(data.EV.Trader.Actor, fmt.Sprintf("%d-%d-%s", poolID, orderID, map[bool]string{true: "0", false: "1"}[data.EV.IsBid]))
	return nil
}
