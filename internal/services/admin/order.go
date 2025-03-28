package admin

import (
	"context"
	entity_admin "exapp-go/internal/entity/admin"
	"exapp-go/pkg/queryparams"
)

func (s *AdminServices) QueryHistoryOrders(ctx context.Context, params *queryparams.QueryParams) ([]*entity_admin.RespHistoryOrder, int64, error) {

	if poolBaseCoin := params.Query.Values["pool_base_coin"]; poolBaseCoin != nil {
		params.CustomQuery["pool_base_coin"] = []interface{}{poolBaseCoin}
	}
	if poolSymbol := params.Query.Values["pool_symbol"]; poolSymbol != nil {
		params.CustomQuery["pool_symbol"] = []interface{}{poolSymbol}
	}
	if app := params.Query.Values["app"]; app != nil {
		params.CustomQuery["app"] = []interface{}{app}
	}
	if trader := params.Query.Values["trader"]; trader != nil {
		params.CustomQuery["trader"] = []interface{}{trader}
	}
	if startTime := params.Query.Values["start_time"]; startTime != nil {
		params.CustomQuery["start_time"] = []interface{}{startTime}
	}
	if endTime := params.Query.Values["end_time"]; endTime != nil {
		params.CustomQuery["end_time"] = []interface{}{endTime}
	}

	orders, total, err := s.ckhdbRepo.QueryHistoryOrdersList(ctx, params)
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

func (s *AdminServices) GetOrdersSymbolTotal(ctx context.Context, startTime, endTime string) ([]*entity_admin.RespOrdersSymbolTotal, error) {

	orders, err := s.ckhdbRepo.GetOrdersSymbolTotal(ctx, startTime, endTime)
	if err != nil {
		return nil, err
	}

	var resp []*entity_admin.RespOrdersSymbolTotal
	for _, order := range orders {
		resp = append(resp, new(entity_admin.RespOrdersSymbolTotal).Fill(order))
	}

	return resp, nil
}
