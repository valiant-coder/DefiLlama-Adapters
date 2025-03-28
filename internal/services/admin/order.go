package admin

import (
	"context"
	ckhdb "exapp-go/internal/db/ckhdb"
	entity_admin "exapp-go/internal/entity/admin"
	"exapp-go/pkg/queryparams"
)

func (s *AdminServices) QueryHistoryOrders(ctx context.Context, queryParams *queryparams.QueryParams) ([]*entity_admin.RespHistoryOrder, int64, error) {
	var orders []*ckhdb.HistoryOrder

	total, err := s.ckhdbRepo.Query(ctx, &orders, queryParams)
	if err != nil {
		return nil, 0, err
	}

	var resp []*entity_admin.RespHistoryOrder
	for _, order := range orders {
		resp = append(resp, new(entity_admin.RespHistoryOrder).Fill(order))
	}
	return resp, total, nil
}

func (s *AdminServices) GetOrdersCoinTotal(ctx context.Context, startTime, endTime string) ([]*entity_admin.RespOrdersCoinTotal, error) {

	orders, err := s.ckhdbRepo.GetOrdersCoinTotal(ctx, startTime, endTime)
	if err != nil {
		return nil, err
	}

	var resp []*entity_admin.RespOrdersCoinTotal
	for _, order := range orders {
		resp = append(resp, new(entity_admin.RespOrdersCoinTotal).Fill(order))
	}

	return resp, nil
}
