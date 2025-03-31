package admin

import (
	"context"
	entity_admin "exapp-go/internal/entity/admin"

	"github.com/shopspring/decimal"
)

func (s *AdminServices) GetBalances(ctx context.Context, isEvmUser bool) (decimal.Decimal, error) {
	usdtBalance, err := s.repo.GetUserTotalBalanceByIsEvmUser(ctx, isEvmUser)
	if err != nil {
		return decimal.Zero, err
	}
	return usdtBalance, nil
}

func (s *AdminServices) GetCoinBalances(ctx context.Context, isEvmUser bool) ([]*entity_admin.RespCoinBalance, error) {

	records, err := s.repo.GetUserCoinTotalBalanceByIsEvmUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var resp []*entity_admin.RespCoinBalance
	for _, record := range records {
		resp = append(resp, new(entity_admin.RespCoinBalance).Fill(record))
	}
	return resp, nil
}
