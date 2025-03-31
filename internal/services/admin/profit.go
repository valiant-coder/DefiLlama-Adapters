package admin

import (
	"context"
	entity_admin "exapp-go/internal/entity/admin"
)

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
