package admin

import (
	"context"
	entity_admin "exapp-go/internal/entity/admin"
	"exapp-go/pkg/queryparams"

	"github.com/shopspring/decimal"
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

func (s *AdminServices) GetOrdersCoinTotal(ctx context.Context, startTime, endTime string) (decimal.Decimal, error) {
	return s.ckhdbRepo.GetOrdersCoinTotal(ctx, startTime, endTime)
}

func (s *AdminServices) GetOrdersCoinQuantity(ctx context.Context, startTime, endTime string) ([]*entity_admin.RespOrdersCoinQuantity, error) {

	orders, err := s.ckhdbRepo.GetOrdersCoinQuantity(ctx, startTime, endTime)
	if err != nil {
		return nil, err
	}

	var resp []*entity_admin.RespOrdersCoinQuantity
	for _, order := range orders {
		resp = append(resp, new(entity_admin.RespOrdersCoinQuantity).Fill(order))
	}

	return resp, nil
}

func (s *AdminServices) GetOrdersSymbolQuantity(ctx context.Context, startTime, endTime string) ([]*entity_admin.RespOrdersSymbolQuantity, error) {

	orders, err := s.ckhdbRepo.GetOrdersSymbolQuantity(ctx, startTime, endTime)
	if err != nil {
		return nil, err
	}

	var resp []*entity_admin.RespOrdersSymbolQuantity
	for _, order := range orders {
		resp = append(resp, new(entity_admin.RespOrdersSymbolQuantity).Fill(order))
	}

	return resp, nil
}

func (s *AdminServices) GetOrdersFeeTotal(ctx context.Context, startTime, endTime string) (decimal.Decimal, error) {
	return s.ckhdbRepo.GetOrdersFeeTotal(ctx, startTime, endTime)
}

func (s *AdminServices) GetOrdersCoinFee(ctx context.Context, startTime, endTime string) ([]*entity_admin.RespOrdersCoinFee, error) {

	orders, err := s.ckhdbRepo.GetOrdersCoinFee(ctx, startTime, endTime)
	if err != nil {
		return nil, err
	}

	var resp []*entity_admin.RespOrdersCoinFee
	for _, order := range orders {
		resp = append(resp, new(entity_admin.RespOrdersCoinFee).Fill(order))
	}

	return resp, nil
}

func (s *AdminServices) GetOrdersSymbolFee(ctx context.Context, startTime, endTime string) ([]*entity_admin.RespOrdersSymbolFee, error) {

	orders, err := s.ckhdbRepo.GetOrdersSymbolFee(ctx, startTime, endTime)
	if err != nil {
		return nil, err
	}

	var resp []*entity_admin.RespOrdersSymbolFee
	for _, order := range orders {
		resp = append(resp, new(entity_admin.RespOrdersSymbolFee).Fill(order))
	}

	return resp, nil
}
