package marketplace

import (
	"context"
	"errors"
	"exapp-go/internal/db/ckhdb"
	"exapp-go/internal/db/db"
	"exapp-go/internal/entity"
	"exapp-go/pkg/queryparams"
	"strings"

	"github.com/spf13/cast"
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
		return make([]entity.OpenOrder, 0), 0, err
	}
	result := make([]entity.OpenOrder, 0, len(orders))
	for _, order := range orders {
		result = append(result, entity.OpenOrderFromDB(order))
	}

	return result, total, nil
}

func (s *OrderService) GetHistoryOrders(ctx context.Context, queryParams *queryparams.QueryParams) ([]entity.HistoryOrder, int64, error) {
	orders, total, err := s.ckhdbRepo.QueryHistoryOrders(ctx, queryParams)
	if err != nil {
		return make([]entity.HistoryOrder, 0), 0, err
	}
	result := make([]entity.HistoryOrder, 0, len(orders))
	for _, order := range orders {
		result = append(result, entity.HistoryOrderFromDB(order))
	}
	return result, total, nil
}

func (s *OrderService) GetHistoryOrderDetail(ctx context.Context, id string) (entity.HistoryOrderDetail, error) {
	params := strings.Split(id, "-")
	if len(params) != 3 {
		return entity.HistoryOrderDetail{}, errors.New("invalid id")
	}
	order, err := s.ckhdbRepo.GetOrder(ctx, cast.ToUint64(params[0]), cast.ToUint64(params[1]), params[2] == "0")
	if err != nil {
		return entity.HistoryOrderDetail{}, err
	}
	return entity.HistoryOrderDetailFromDB(order), nil
}
