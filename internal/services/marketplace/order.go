package marketplace

import (
	"context"
	"exapp-go/internal/db/ckhdb"
	"exapp-go/internal/db/db"
	"exapp-go/internal/entity"
	"exapp-go/pkg/queryparams"
)

type OrderService struct {
	ckhdbRepo *ckhdb.ClickHouseRepo
	repo      *db.Repo
}

func NewOrderService() *OrderService {
	return &OrderService{
		ckhdbRepo: ckhdb.New(),
		repo:      db.New(),
	}
}

func (s *OrderService) GetOpenOrders(ctx context.Context, queryParams *queryparams.QueryParams) ([]entity.OpenOrder, int64, error) {
	orders, total, err := s.repo.GetOpenOrders(ctx, queryParams)
	if err != nil {
		return nil, 0, err
	}
	var result []entity.OpenOrder
	for _, order := range orders {
		result = append(result, entity.OpenOrderFromDB(order))
	}

	return result, total, nil
}

func (s *OrderService) GetHistoryOrders(ctx context.Context, queryParams *queryparams.QueryParams) ([]entity.HistoryOrder, int64, error) {
	orders, total, err := s.ckhdbRepo.QueryHistoryOrders(ctx, queryParams)
	if err != nil {
		return nil, 0, err
	}
	var result []entity.HistoryOrder
	for _, order := range orders {
		result = append(result, entity.HistoryOrderFromDB(order))
	}
	return result, total, nil
}

func (s *OrderService) GetHistoryOrderDetail(ctx context.Context, id uint64) (entity.HistoryOrderDetail, error) {
	order, err := s.ckhdbRepo.GetOrder(ctx, id)
	if err != nil {
		return entity.HistoryOrderDetail{}, err
	}
	return entity.HistoryOrderDetailFromDB(order), nil
}
