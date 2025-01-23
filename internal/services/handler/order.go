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
			Status:             db.OrderStatus(newOrder.EV.Status),
		}
		err := s.repo.InsertOpenOrderIfNotExist(ctx, &order)
		if err != nil {
			log.Printf("insert open order failed: %v", err)
			return nil
		}
		err = s.updateDepth(ctx, db.UpdateDepthParams{
			PoolID: poolID,
			UniqID: cast.ToString(action.GlobalSequence),
			Price:  order.Price,
			Amount: placedQuantity,
			IsBuy:  newOrder.EV.IsBid,
		})
		if err != nil {
			log.Printf("update depth failed: :%v", err)
		}

	} else {
		var avgPrice, price decimal.Decimal
		if executedQuantity.GreaterThan(decimal.Zero) {
			var orderTag string
			if newOrder.EV.IsBid {
				orderTag = fmt.Sprintf("%d-%d-%d", poolID, orderID, 0)
			} else {
				orderTag = fmt.Sprintf("%d-%d-%d", poolID, orderID, 1)
			}
			trades, err := s.ckhRepo.GetTrades(ctx, orderTag)
			if err != nil {
				log.Printf("get trades failed: %v", err)
				return nil
			}
			if len(trades) == 0 {
				log.Printf("no trades found for executed order: %v", orderTag)
				return s.publisher.DeferPublishCreateOrder(action)
			}
			var totalQuoteQuantity, totalBaseQuantity decimal.Decimal
			for _, trade := range trades {
				totalQuoteQuantity = totalQuoteQuantity.Add(trade.QuoteQuantity)
				totalBaseQuantity = totalBaseQuantity.Add(trade.BaseQuantity)
			}
			avgPrice = totalQuoteQuantity.Div(totalBaseQuantity).Round(int32(pool.PricePrecision))
			price = avgPrice
			if newOrder.EV.IsMarket {
				originalQuantity = totalBaseQuantity
				executedQuantity = totalBaseQuantity
			}
		} else {
			avgPrice = decimal.New(cast.ToInt64(newOrder.EV.Price), -int32(pool.PricePrecision))
			price = avgPrice
		}
		order := ckhdb.HistoryOrder{
			App:              newOrder.EV.App,
			CreateTxID:       action.TrxID,
			CreateBlockNum:   action.BlockNum,
			PoolID:           poolID,
			PoolSymbol:       pool.Symbol,
			PoolBaseCoin:     pool.BaseCoin,
			PoolQuoteCoin:    pool.QuoteCoin,
			OrderID:          orderID,
			ClientOrderID:    newOrder.EV.OrderCID,
			Trader:           newOrder.EV.Trader.Actor,
			Price:            price,
			AvgPrice:         avgPrice,
			IsBid:            newOrder.EV.IsBid,
			OriginalQuantity: originalQuantity,
			ExecutedQuantity: executedQuantity,
			Status:           ckhdb.OrderStatus(newOrder.EV.Status),
			IsMarket:         newOrder.EV.IsMarket,
			CreatedAt:        createTime,
		}
		err = s.ckhRepo.InsertOrderIfNotExist(ctx, &order)
		if err != nil {
			log.Printf("insert history order failed: %v", err)
			return nil
		}

	}

	go s.updateUserTokenBalance(newOrder.EV.Trader.Actor)
	go s.publisher.PublishOrderUpdate(newOrder.EV.Trader.Actor, fmt.Sprintf("%d-%d-%s", poolID, orderID, map[bool]string{true: "0", false: "1"}[newOrder.EV.IsBid]))
	return nil
}

func (s *Service) handleMatchOrder(action hyperion.Action) error {
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
	err = s.newTrade(ctx, &trade)
	if err != nil {
		log.Printf("new trade failed: %v", err)
	}

	// update depth
	if data.EV.TakerIsBid {
		// Taker is buyer, decrease sell depth
		err = s.updateDepth(ctx, db.UpdateDepthParams{
			PoolID: poolID,
			UniqID: cast.ToString(action.GlobalSequence),
			Price:  trade.Price,
			Amount: baseQuantity.Neg(),
			IsBuy:  false,
		})
		if err != nil {
			log.Printf("update depth failed: :%v", err)
		}
	} else {
		// Taker is seller, decrease buy depth
		err = s.updateDepth(ctx, db.UpdateDepthParams{
			PoolID: poolID,
			UniqID: cast.ToString(action.GlobalSequence),
			Price:  trade.Price,
			Amount: baseQuantity.Neg(),
			IsBuy:  true,
		})
		if err != nil {
			log.Printf("update depth failed: :%v", err)
		}
	}

	makerOrder, err := s.repo.GetOpenOrder(ctx, poolID, cast.ToUint64(data.EV.MakerOrderID), !data.EV.TakerIsBid)
	if err != nil {
		log.Printf("get maker order failed: %v", err)
		return nil
	}

	makerOrder.ExecutedQuantity = makerOrder.ExecutedQuantity.Add(baseQuantity)
	makerOrder.Status = db.OrderStatus(data.EV.MakerStatus)
	if makerOrder.ExecutedQuantity.GreaterThan(makerOrder.OriginalQuantity) {
		log.Printf("trade executed quantity greater than original quantity: %v,%v", trade.TxID,makerOrder.TxID)
		log.Panicf("marker order original quantity: %v, executed quantity: %v", makerOrder.OriginalQuantity,makerOrder.ExecutedQuantity)
		log.Printf("current trade base quantity: %v",trade.BaseQuantity)
	}


	
	// update maker order
	if !data.EV.MakerRemoved {
		err = s.repo.UpdateOpenOrder(ctx, makerOrder)
		if err != nil {
			log.Printf("update open order failed: %v", err)
		}
	} else {
		err = s.repo.DeleteOpenOrder(ctx, poolID, cast.ToUint64(data.EV.MakerOrderID), makerOrder.IsBid)
		if err != nil {
			log.Printf("delete open order failed: %v", err)
		}

		historyOrder := ckhdb.HistoryOrder{
			App:              makerOrder.App,
			PoolID:           makerOrder.PoolID,
			PoolSymbol:       pool.Symbol,
			PoolBaseCoin:     pool.BaseCoin,
			PoolQuoteCoin:    pool.QuoteCoin,
			OrderID:          makerOrder.OrderID,
			ClientOrderID:    makerOrder.ClientOrderID,
			Trader:           makerOrder.Trader,
			Price:            makerOrder.Price,
			AvgPrice:         makerOrder.Price,
			IsBid:            makerOrder.IsBid,
			OriginalQuantity: makerOrder.OriginalQuantity,
			ExecutedQuantity: makerOrder.ExecutedQuantity,
			Status:           ckhdb.OrderStatus(makerOrder.Status),
			IsMarket:         false,
			CreateTxID:       makerOrder.TxID,
			CreateBlockNum:   makerOrder.BlockNumber,
			CreatedAt:        makerOrder.CreatedAt,
		}
		err = s.ckhRepo.InsertOrderIfNotExist(ctx, &historyOrder)
		if err != nil {
			log.Printf("insert history order failed: %v", err)
		}

	}
	go s.updateUserTokenBalance(data.EV.Maker.Actor)
	go s.updateUserTokenBalance(data.EV.Taker.Actor)
	go s.publisher.PublishOrderUpdate(data.EV.Maker.Actor, fmt.Sprintf("%d-%d-%s", poolID, cast.ToUint64(data.EV.MakerOrderID), map[bool]string{true: "0", false: "1"}[!data.EV.TakerIsBid]))
	go s.publisher.PublishOrderUpdate(data.EV.Taker.Actor, fmt.Sprintf("%d-%d-%s", poolID, cast.ToUint64(data.EV.TakerOrderID), map[bool]string{true: "0", false: "1"}[!data.EV.TakerIsBid]))

	return nil
}

func (s *Service) handleCancelOrder(action hyperion.Action) error {
	var data struct {
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
			OriginalQuantity     string `json:"original_quantity"`
			CanceledBaseQuantity string `json:"canceled_base_quantity"`
			Time                 string `json:"time"`
		} `json:"ev"`
	}

	if err := json.Unmarshal(action.Act.Data, &data); err != nil {
		log.Printf("unmarshal cancel order data failed: %v", err)
		return nil
	}

	ctx := context.Background()
	canceledQuantity, err := eosAssetToDecimal(data.EV.CanceledBaseQuantity)
	if err != nil {
		log.Printf("new asset from string failed: %v", err)
		return nil
	}

	orderID := cast.ToUint64(data.EV.OrderID)
	poolID := cast.ToUint64(data.EV.PoolID)
	order, err := s.repo.GetOpenOrder(ctx, poolID, orderID, data.EV.IsBid)
	if err != nil {
		log.Printf("get open order failed: %v", err)
		return nil
	}

	// update depth
	if data.EV.IsBid {
		err = s.updateDepth(ctx, db.UpdateDepthParams{
			PoolID: poolID,
			UniqID: cast.ToString(action.GlobalSequence),
			Price:  order.Price,
			Amount: canceledQuantity.Neg(),
			IsBuy:  true,
		})
		if err != nil {
			log.Printf("update depth failed: :%v", err)
			return nil
		}
	} else {
		err = s.updateDepth(ctx, db.UpdateDepthParams{
			PoolID: poolID,
			UniqID: cast.ToString(action.GlobalSequence),
			Price:  order.Price,
			Amount: canceledQuantity.Neg(),
			IsBuy:  false,
		})
		if err != nil {
			log.Printf("update depth failed: :%v", err)
			return nil
		}
	}
	// delete open order
	err = s.repo.DeleteOpenOrder(ctx, poolID, orderID, order.IsBid)
	if err != nil {
		log.Printf("delete open order failed: %v", err)
		return nil
	}
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
	// insert history order
	historyOrder := ckhdb.HistoryOrder{
		App:              order.App,
		PoolID:           order.PoolID,
		PoolSymbol:       order.PoolSymbol,
		PoolBaseCoin:     order.PoolBaseCoin,
		PoolQuoteCoin:    order.PoolQuoteCoin,
		OrderID:          order.OrderID,
		ClientOrderID:    order.ClientOrderID,
		Trader:           order.Trader,
		Price:            order.Price,
		AvgPrice:         order.Price,
		IsBid:            order.IsBid,
		OriginalQuantity: order.OriginalQuantity,
		ExecutedQuantity: order.ExecutedQuantity,
		Status:           status,
		IsMarket:         false,
		CreateTxID:       order.TxID,
		CreatedAt:        order.CreatedAt,
		CreateBlockNum:   order.BlockNumber,
		CancelTxID:       action.TrxID,
		CancelBlockNum:   action.BlockNum,
		CanceledAt:       canceledTime,
		Type:             ckhdb.OrderType(order.Type),
	}
	err = s.ckhRepo.InsertOrderIfNotExist(ctx, &historyOrder)
	if err != nil {
		log.Printf("insert history order failed: %v", err)
		return nil
	}
	go s.updateUserTokenBalance(data.EV.Trader.Actor)
	go s.publisher.PublishOrderUpdate(data.EV.Trader.Actor, fmt.Sprintf("%d-%d-%s", poolID, orderID, map[bool]string{true: "0", false: "1"}[data.EV.IsBid]))
	return nil
}
