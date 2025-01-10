package handler

import (
	"context"
	"exapp-go/internal/db/ckhdb"
	"exapp-go/internal/entity"
	"log"
)

func (s *Service) newTrade(ctx context.Context, trade *ckhdb.Trade) error {
	err := s.ckhRepo.InsertTradeIfNotExist(ctx, trade)
	if err != nil {
		log.Printf("insert trade failed: %v", err)
		return nil
	}
	go s.publisher.PublishTradeUpdate(entity.Trade{
		PoolID:   trade.PoolID,
		Buyer:    trade.Taker,
		Seller:   trade.Maker,
		Quantity: trade.BaseQuantity.String(),
		Price:    trade.Price.String(),
		TradedAt: entity.Time(trade.Time),
		Side:     entity.TradeSide(map[bool]string{true: "buy", false: "sell"}[trade.TakerIsBid]),
	})

	klines, err := s.ckhRepo.GetLatestKlines(ctx, trade.PoolID)
	if err != nil {
		log.Printf("get latest kline failed: %v", err)
		return nil
	}
	for _, kline := range klines {
		err = s.publisher.PublishKlineUpdate(entity.DbKlineToEntity(kline))
		if err != nil {
			log.Printf("publish kline update failed: %v", err)
		}
	}

	return nil

}
