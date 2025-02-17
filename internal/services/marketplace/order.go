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
	"gorm.io/gorm"
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

func (s *OrderService) GetHistoryOrders(ctx context.Context, queryParams *queryparams.QueryParams) ([]entity.Order, int64, error) {
	orders, total, err := s.ckhdbRepo.QueryHistoryOrders(ctx, queryParams)
	if err != nil {
		return make([]entity.Order, 0), 0, err
	}
	result := make([]entity.Order, 0, len(orders))
	for _, order := range orders {
		result = append(result, entity.OrderFromHistoryDB(order))
	}
	return result, total, nil
}

func (s *OrderService) GetHistoryOrderDetail(ctx context.Context, id string) (entity.OrderDetail, error) {
	params := strings.Split(id, "-")
	if len(params) != 3 {
		return entity.OrderDetail{}, errors.New("invalid id")
	}
	poolID := cast.ToUint64(params[0])
	orderID := cast.ToUint64(params[1])
	isBid := params[2] == "0"

	orderDetail := entity.OrderDetail{}
	order, err := s.ckhdbRepo.GetOrder(ctx, poolID, orderID, isBid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			openOrder, err := s.repo.GetOpenOrder(ctx, poolID, orderID, isBid)
			if err != nil {
				return entity.OrderDetail{}, err
			}
			orderDetail.Order = entity.OrderFormOpenDB(*openOrder)
		}
		return entity.OrderDetail{}, err
	}
	orderDetail.Order = entity.OrderFromHistoryDB(*order)
	if orderDetail.Status == 3 {
		orderDetail.Trades = []entity.TradeDetail{}
		return orderDetail, nil
	}

	trades, err := s.ckhdbRepo.GetTrades(ctx, id)
	if err != nil {
		return entity.OrderDetail{}, err
	}
	orderDetail.Trades = entity.TradeDetailFromDB(trades)
	return orderDetail, nil
}
