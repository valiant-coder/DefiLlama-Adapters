package marketplace

import (
	"context"
	"exapp-go/internal/db/ckhdb"
	"exapp-go/internal/entity"
)

func NewTradeService() *TradeService {
	return &TradeService{ckhdb: ckhdb.New()}
}

type TradeService struct {
	ckhdb *ckhdb.ClickHouseRepo
}


func (s *TradeService) GetLatestTrades(ctx context.Context, poolID uint64, limit int) ([]entity.Trade, error) {
	trades, err := s.ckhdb.GetLatestTrades(ctx, poolID, limit)
	if err != nil {
		return nil, err
	}
	tradeList := make([]entity.Trade, 0, len(trades))
	for _, trade := range trades {
		tradeList = append(tradeList, entity.DbTradeToTrade(trade))
	}

	return tradeList, nil
}
