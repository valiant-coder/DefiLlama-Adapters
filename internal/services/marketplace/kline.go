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
	var klines []*ckhdb.Kline
	var err error
	if klines, err = s.repo.GetKline(ctx, poolID, interval, start, end); err != nil {
		return nil, err
	}

	for i := 0; i < len(klines); i++ {
		if i > 0 {
			klines[i].Open = klines[i-1].Close

			if klines[i].Open.IsZero() {
				klines[i].Open = klines[i].Low
			}
			if klines[i].High.IsZero() {
				klines[i].High = klines[i].Open
			}
			if klines[i].Low.IsZero() {
				klines[i].Low = klines[i].Open
			}
			if klines[i].Close.IsZero() {
				klines[i].Close = klines[i].Open
			}
			if klines[i].Low.GreaterThan(klines[i].Open) {
				klines[i].Low = klines[i].Open
			}
			if klines[i].High.LessThan(klines[i].Open) {
				klines[i].High = klines[i].Open
			}
		}
	}


	var entityKlines []entity.Kline
	for _, kline := range klines {
		entityKlines = append(entityKlines, entity.DbKlineToEntity(kline))
	}
	return entityKlines, nil
}
