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

	placedAsset, err := eosgo.NewAssetFromString(newOrder.PlacedQuantity)
	if err != nil {
		log.Printf("new asset from string failed: %v", err)
		return nil
	}

	executedAsset, err := eosgo.NewAssetFromString(newOrder.ExecutedQuantity)
	if err != nil {
		log.Printf("new asset from string failed: %v", err)
		return nil
	}

	executedQuantity := decimal.New(int64(executedAsset.Amount), int32(executedAsset.Symbol.Precision))
	placedQuantity := decimal.New(int64(placedAsset.Amount), int32(placedAsset.Symbol.Precision))
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
		MakerOrderID string `json:"maker_order_id"`
		TakerOrderID string `json:"taker_order_id"`
		Price        string `json:"price"`
		Amount       string `json:"amount"`
	}

	if err := json.Unmarshal(action.Act.Data, &data); err != nil {
		return fmt.Errorf("unmarshal match order data failed: %w", err)
	}

	return nil
}

func (s *Service) handleCancelOrder(action hyperion.Action) error {
	var data struct {
		OrderID string `json:"order_id"`
	}

	if err := json.Unmarshal(action.Act.Data, &data); err != nil {
		return fmt.Errorf("unmarshal cancel order data failed: %w", err)
	}

	return nil
}
