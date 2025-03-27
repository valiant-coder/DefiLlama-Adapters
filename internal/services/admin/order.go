package admin

import (
	"context"
	"exapp-go/internal/db/db"
	entity_admin "exapp-go/internal/entity/admin"
	"exapp-go/pkg/queryparams"
)

func (s *AdminServices) QueryOpenOrders(ctx context.Context, queryParams *queryparams.QueryParams) ([]*entity_admin.RespOpenOrder, int64, error) {
	var orders []*db.OpenOrder

	total, err := s.repo.Query(ctx, &orders, queryParams)
	if err != nil {
		return nil, 0, err
	}

	var resp []*entity_admin.RespOpenOrder
	for _, order := range orders {
		resp = append(resp, new(entity_admin.RespOpenOrder).Fill(order))
	}
	return resp, total, nil
}

func (s *AdminServices) GetOrdersCoinTotal(ctx context.Context, startTime, endTime string) ([]*entity_admin.RespOrdersCoinTotal, int64, error) {

	_, err := s.repo.GetOrdersCoinTotal(ctx, startTime, endTime)
	if err != nil {
		return nil, 0, err
	}

	return nil, 0, nil
}
