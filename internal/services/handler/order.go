package handler

import (
	"context"
	"encoding/json"
	"exapp-go/internal/db/ckhdb"
	"exapp-go/internal/db/db"
	"exapp-go/pkg/hyperion"
	"fmt"
	"log"
	"time"

	eosgo "github.com/eoscanada/eos-go"
	"github.com/shopspring/decimal"
)

func eosAssetToDecimal(a string) (decimal.Decimal, error) {
	asset, err := eosgo.NewAssetFromString(a)
	if err != nil {
		return decimal.Decimal{}, err
	}
	return decimal.New(int64(asset.Amount), int32(asset.Symbol.Precision)), nil
}

func (s *Service) handleCreateOrder(action hyperion.Action) error {
	var newOrder struct {
		PoolID           uint64 `json:"pool_id"`
		OrderID          uint64 `json:"order_id"`
		OrderCID         string `json:"order_cid"`
		IsBid            bool   `json:"is_bid"`
		Trader           string `json:"trader"`
		ExecutedQuantity string `json:"executed_quantity"`
		PlacedQuantity   string `json:"placed_quantity"`
		Price            uint64 `json:"price"`
		Status           uint8  `json:"status"`
		IsMarket         bool   `json:"is_market"`
		IsInserted       bool   `json:"is_inserted"`
		Time             string `json:"time"`
	}

	if err := json.Unmarshal(action.Act.Data, &newOrder); err != nil {
		log.Printf("unmarshal create order data failed: %v", err)
		return nil
	}

	ctx := context.Background()

	pool, ok := s.poolCache[newOrder.PoolID]
	if !ok {
		pool, err := s.ckhRepo.GetPool(ctx, newOrder.PoolID)
		if err != nil {
			log.Printf("get pool failed: %v", err)
			return nil
		}
		s.poolCache[newOrder.PoolID] = pool
	}

	placedQuantity, err := eosAssetToDecimal(newOrder.PlacedQuantity)
	if err != nil {
		log.Printf("new asset from string failed: %v", err)
		return nil
	}

	executedQuantity, err := eosAssetToDecimal(newOrder.ExecutedQuantity)
	if err != nil {
		log.Printf("new asset from string failed: %v", err)
		return nil
	}

	originalQuantity := placedQuantity.Add(executedQuantity)

	time, err := time.Parse(time.RFC3339, newOrder.Time)
	if err != nil {
		log.Printf("parse action time failed: %v", err)
		return nil
	}

	if newOrder.IsInserted {
		order := db.OpenOrder{
			TxID:             action.TrxID,
			CreatedAt:        time,
			BlockNumber:      action.BlockNum,
			OrderID:          newOrder.OrderID,
			PoolID:           newOrder.PoolID,
			ClientOrderID:    newOrder.OrderCID,
			Trader:           newOrder.Trader,
			Price:            decimal.New(int64(newOrder.Price), -int32(pool.PricePrecision)),
			IsBid:            newOrder.IsBid,
			OriginalQuantity: originalQuantity,
			ExecutedQuantity: executedQuantity,
			Status:           db.OrderStatus(newOrder.Status),
			InOrderBook:      true,
		}
		err := s.repo.InsertOpenOrder(ctx, &order)
		if err != nil {
			log.Printf("insert open order failed: %v", err)
			return nil
		}
		err = s.repo.UpdateDepth(ctx, []db.UpdateDepthParams{
			{
				PoolID: newOrder.PoolID,
				Price:  order.Price.InexactFloat64(),
				Amount: placedQuantity.InexactFloat64(),
				IsBuy:  newOrder.IsBid,
			},
		})
		if err != nil {
			log.Printf("update depth failed: :%v", err)
			return nil
		}
		return nil

	} else {

		order := ckhdb.HistoryOrder{
			CreateTxID:       action.TrxID,
			CreateBlockNum:   action.BlockNum,
			PoolID:           newOrder.PoolID,
			OrderID:          newOrder.OrderID,
			ClientOrderID:    newOrder.OrderCID,
			Trader:           newOrder.Trader,
			Price:            decimal.New(int64(newOrder.Price), -int32(pool.PricePrecision)),
			IsBid:            newOrder.IsBid,
			OriginalQuantity: originalQuantity,
			ExecutedQuantity: executedQuantity,
			Status:           ckhdb.OrderStatus(newOrder.Status),
			IsMarket:         newOrder.IsMarket,
			CreatedAt:        time,
		}
		err = s.ckhRepo.InsertHistoryOrder(ctx, &order)
		if err != nil {
			log.Printf("insert history order failed: %v", err)
			return nil
		}

	}

	return nil
}

func (s *Service) handleMatchOrder(action hyperion.Action) error {
	var data struct {
		PoolID        uint64 `json:"pool_id"`
		Taker         string `json:"taker"`
		Maker         string `json:"maker"`
		MakerOrderID  uint64 `json:"maker_order_id"`
		MakerOrderCID string `json:"maker_order_cid"`
		TakerOrderID  uint64 `json:"taker_order_id"`
		TakerOrderCID string `json:"taker_order_cid"`
		Price         uint64 `json:"price"`
		TakerIsBid    bool   `json:"taker_is_bid"`
		BaseQuantity  string `json:"base_quantity"`
		QuoteQuantity string `json:"quote_quantity"`

		TakerFee             string `json:"taker_fee"`
		MakerFee             string `json:"maker_fee"`
		Time                 string `json:"time"`
		MakerOrderStatus     uint8  `json:"maker_order_status"`
		MakerOrderIsInserted bool   `json:"maker_order_is_inserted"`
	}
	if err := json.Unmarshal(action.Act.Data, &data); err != nil {
		return fmt.Errorf("unmarshal match order data failed: %w", err)
	}

	ctx := context.Background()
	pool, ok := s.poolCache[data.PoolID]
	if !ok {
		pool, err := s.ckhRepo.GetPool(ctx, data.PoolID)
		if err != nil {
			log.Printf("get pool failed: %v", err)
			return nil
		}
		s.poolCache[data.PoolID] = pool
	}

	baseQuantity, err := eosAssetToDecimal(data.BaseQuantity)
	if err != nil {
		log.Printf("new asset from string failed: %v", err)
		return nil
	}
	quoteQuantity, err := eosAssetToDecimal(data.QuoteQuantity)
	if err != nil {
		log.Printf("new asset from string failed: %v", err)
		return nil
	}
	takerFee, err := eosAssetToDecimal(data.TakerFee)
	if err != nil {
		log.Printf("new asset from string failed: %v", err)
		return nil
	}
	makerFee, err := eosAssetToDecimal(data.MakerFee)
	if err != nil {
		log.Printf("new asset from string failed: %v", err)
		return nil
	}

	timestamp, err := time.Parse(time.RFC3339, data.Time)
	if err != nil {
		log.Printf("parse action time failed: %v", err)
		return nil
	}

	trade := ckhdb.Trade{
		TxID:           action.TrxID,
		PoolID:         data.PoolID,
		Price:          decimal.New(int64(data.Price), -int32(pool.PricePrecision)),
		Timestamp:      timestamp,
		BlockNumber:    action.BlockNum,
		GlobalSequence: action.GlobalSequence,
		Taker:          data.Taker,
		Maker:          data.Maker,
		MakerOrderID:   data.MakerOrderID,
		MakerOrderCID:  data.MakerOrderCID,
		TakerOrderID:   data.TakerOrderID,
		TakerOrderCID:  data.TakerOrderCID,
		BaseQuantity:   baseQuantity,
		QuoteQuantity:  quoteQuantity,
		TakerFee:       takerFee,
		MakerFee:       makerFee,
		TakerIsBid:     data.TakerIsBid,
		CreatedAt:      time.Now(),
	}
	err = s.ckhRepo.InsertTrade(ctx, &trade)
	if err != nil {
		log.Printf("insert trade failed: %v", err)
		return nil
	}

	// update depth
	if data.TakerIsBid {
		// Taker is buyer, decrease sell depth
		err = s.repo.UpdateDepth(ctx, []db.UpdateDepthParams{
			{
				PoolID: data.PoolID,
				Price:  trade.Price.InexactFloat64(),
				Amount: -baseQuantity.InexactFloat64(),
				IsBuy:  false,
			},
		})
		if err != nil {
			log.Printf("update depth failed: :%v", err)
			return nil
		}
	} else {
		// Taker is seller, decrease buy depth
		err = s.repo.UpdateDepth(ctx, []db.UpdateDepthParams{
			{
				PoolID: data.PoolID,
				Price:  trade.Price.InexactFloat64(),
				Amount: -baseQuantity.InexactFloat64(),
				IsBuy:  true,
			},
		})
		if err != nil {
			log.Printf("update depth failed: :%v", err)
			return nil
		}
	}

	makerOrder, err := s.repo.GetOpenOrder(ctx, data.MakerOrderID)
	if err != nil {
		log.Printf("get maker order failed: %v", err)
		return nil
	}

	makerOrder.ExecutedQuantity = makerOrder.ExecutedQuantity.Add(quoteQuantity)
	makerOrder.Status = db.OrderStatus(data.MakerOrderStatus)

	// update maker order
	if data.MakerOrderIsInserted {
		err = s.repo.UpdateOpenOrder(ctx, makerOrder)
		if err != nil {
			log.Printf("update open order failed: %v", err)
			return nil
		}
	} else {
		err = s.repo.DeleteOpenOrder(ctx, data.MakerOrderID)
		if err != nil {
			log.Printf("delete open order failed: %v", err)
			return nil
		}

		historyOrder := ckhdb.HistoryOrder{
			PoolID:           makerOrder.PoolID,
			OrderID:          makerOrder.OrderID,
			ClientOrderID:    makerOrder.ClientOrderID,
			Trader:           makerOrder.Trader,
			Price:            makerOrder.Price,
			IsBid:            makerOrder.IsBid,
			OriginalQuantity: makerOrder.OriginalQuantity,
			ExecutedQuantity: makerOrder.ExecutedQuantity,
			Status:           ckhdb.OrderStatus(makerOrder.Status),
			IsMarket:         false,
			CreateTxID:       makerOrder.TxID,
			CreateBlockNum:   makerOrder.BlockNumber,
		}
		err = s.ckhRepo.InsertHistoryOrder(ctx, &historyOrder)
		if err != nil {
			log.Printf("insert history order failed: %v", err)
			return nil
		}

	}

	return nil
}

func (s *Service) handleCancelOrder(action hyperion.Action) error {
	var data struct {
		PoolID               uint64 `json:"pool_id"`
		OrderID              uint64 `json:"order_id"`
		OrderCID             string `json:"order_cid"`
		IsBid                bool   `json:"is_bid"`
		Trader               string `json:"trader"`
		OriginalQuantity     string `json:"original_quantity"`
		CanceledBaseQuantity string `json:"canceled_base_quantity"`
		Time                 string `json:"time"`
	}

	if err := json.Unmarshal(action.Act.Data, &data); err != nil {
		return fmt.Errorf("unmarshal cancel order data failed: %w", err)
	}

	ctx := context.Background()
	canceledQuantity, err := eosAssetToDecimal(data.CanceledBaseQuantity)
	if err != nil {
		log.Printf("new asset from string failed: %v", err)
		return nil
	}

	order, err := s.repo.GetOpenOrder(ctx, data.OrderID)
	if err != nil {
		log.Printf("get open order failed: %v", err)
		return nil
	}

	// update depth
	if data.IsBid {
		err = s.repo.UpdateDepth(ctx, []db.UpdateDepthParams{
			{
				PoolID: data.PoolID,
				Price:  order.Price.InexactFloat64(),
				Amount: -canceledQuantity.InexactFloat64(),
				IsBuy:  true,
			},
		})
		if err != nil {
			log.Printf("update depth failed: :%v", err)
			return nil
		}
	} else {
		err = s.repo.UpdateDepth(ctx, []db.UpdateDepthParams{
			{
				PoolID: data.PoolID,
				Price:  order.Price.InexactFloat64(),
				Amount: -canceledQuantity.InexactFloat64(),
				IsBuy:  false,
			},
		})
		if err != nil {
			log.Printf("update depth failed: :%v", err)
			return nil
		}
	}
	// delete open order
	err = s.repo.DeleteOpenOrder(ctx, data.OrderID)
	if err != nil {
		log.Printf("delete open order failed: %v", err)
		return nil
	}
	time, err := time.Parse(time.RFC3339, data.Time)
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
		PoolID:           order.PoolID,
		OrderID:          order.OrderID,
		ClientOrderID:    order.ClientOrderID,
		Trader:           order.Trader,
		Price:            order.Price,
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
		CanceledAt:       time,
		Type:             ckhdb.OrderType(order.Type),
	}
	err = s.ckhRepo.InsertHistoryOrder(ctx, &historyOrder)
	if err != nil {
		log.Printf("insert history order failed: %v", err)
		return nil
	}

	return nil
}
