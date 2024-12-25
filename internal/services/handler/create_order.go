package handler

import (
	"context"
	"encoding/json"
	"exapp-go/internal/db/ckhdb"
	"exapp-go/internal/db/db"
	"exapp-go/pkg/hyperion"
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
	originalQuantity := decimal.New(int64(placedAsset.Amount), int32(placedAsset.Symbol.Precision)).Add(executedQuantity)

	ctx := context.Background()
	if newOrder.IsInserted {
		order := db.OpenOrder{
			TxID:             action.TrxID,
			OrderID:          newOrder.OrderID,
			PoolID:           newOrder.PoolID,
			ClientOrderID:    newOrder.OrderCID,
			Trader:           newOrder.Trader,
			Price:            newOrder.Price,
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

		// todo update depth
		return nil

	} else {

		time, err := time.Parse(time.RFC3339, newOrder.Time)
		if err != nil {
			log.Printf("parse time failed: %v", err)
			return nil
		}

		order := ckhdb.HistoryOrder{
			CreateTxID:       action.TrxID,
			PoolID:           newOrder.PoolID,
			OrderID:          newOrder.OrderID,
			ClientOrderID:    newOrder.OrderCID,
			Trader:           newOrder.Trader,
			Price:            newOrder.Price,
			IsBid:            newOrder.IsBid,
			OriginalQuantity: originalQuantity,
			ExecutedQuantity: executedQuantity,
			Status:           ckhdb.OrderStatus(newOrder.Status),
			IsMarket:         newOrder.IsMarket,
			CreateTime:       time,
		}
		err = s.ckhRepo.InsertHistoryOrder(ctx, &order)
		if err != nil {
			log.Printf("insert history order failed: %v", err)
			return nil
		}

	}

	return nil
}
