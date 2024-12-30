package marketplace

import (
	"context"
	"exapp-go/internal/db/ckhdb"
	"exapp-go/internal/entity"
	"time"
)

func NewKlineService() *KlineService {
	return &KlineService{repo: ckhdb.New()}
}

type KlineService struct {
	repo *ckhdb.ClickHouseRepo
}

func (s *KlineService) GetKline(ctx context.Context, poolID uint64, interval string, start time.Time, end time.Time) ([]entity.Kline, error) {
	var klines []ckhdb.Kline
	var err error
	if klines, err = s.repo.GetKline(ctx, poolID, interval, start, end); err != nil {
		return nil, err
	}
	var entityKlines []entity.Kline
	for _, kline := range klines {
		entityKlines = append(entityKlines, entity.DbKlineToEntity(kline))
	}
	return entityKlines, nil
}
