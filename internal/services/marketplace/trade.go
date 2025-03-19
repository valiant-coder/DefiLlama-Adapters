package marketplace

import (
	"context"
	"exapp-go/internal/db/ckhdb"
	"exapp-go/internal/db/db"
	"exapp-go/internal/entity"
)

func NewTradeService() *TradeService {
	return &TradeService{ckhRepo: ckhdb.New(), repo: db.New()}
}

type TradeService struct {
	ckhRepo *ckhdb.ClickHouseRepo
	repo    *db.Repo
}

func (s *TradeService) GetLatestTrades(ctx context.Context, poolID uint64, limit int) ([]entity.Trade, error) {
	trades, err := s.ckhRepo.GetLatestTradesByPool(ctx, poolID, limit)
	if err != nil {
		return make([]entity.Trade, 0), err
	}
	tradeList := make([]entity.Trade, 0, len(trades))
	for _, trade := range trades {
		tradeList = append(tradeList, entity.DbTradeToTrade(trade))
	}

	return tradeList, nil
}

func (s *TradeService) GetTradeCountAndVolume(ctx context.Context) (entity.SysTradeInfo, error) {
	tradeCount, tradeVolume, err := s.ckhRepo.GetTradeCountAndVolume(ctx)
	if err != nil {
		return entity.SysTradeInfo{}, err
	}
	totalUserCount, err := s.repo.GetTotalUserCount(ctx)
	if err != nil {
		return entity.SysTradeInfo{}, err
	}
	return entity.SysTradeInfo{
		TotalTrades:    tradeCount,
		TotalVolume:    tradeVolume,
		TotalUserCount: totalUserCount,
	}, nil
}
