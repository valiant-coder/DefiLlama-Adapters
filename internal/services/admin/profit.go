package admin

import (
	"context"
	"exapp-go/internal/db/db"
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

func (s *AdminServices) GetUserBalanceStat(ctx context.Context, isEvmUser bool, minValue, maxValue int64, rangeCount int) ([]db.BalanceRange, error) {
	return s.repo.GetUserBalanceDistribution(ctx, db.BalanceRangeConfig{
		MinValue:   decimal.New(minValue, 0),
		MaxValue:   decimal.New(maxValue, 0),
		RangeCount: rangeCount,
	}, isEvmUser)
}
