package admin

import (
	"context"
	"exapp-go/internal/db/db"
	entity_admin "exapp-go/internal/entity/admin"
	"exapp-go/pkg/queryparams"

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

func (s *AdminServices) QueryUserBalance(ctx context.Context, params *queryparams.QueryParams) ([]*entity_admin.RespUserBalance, error) {
	if uid := params.Query.Values["username"]; uid != nil {
		params.CustomQuery["username"] = []interface{}{uid}
	}
	if coin := params.Query.Values["uid"]; coin != nil {
		params.CustomQuery["uid"] = []interface{}{coin}
	}

	users, err := s.repo.QueryUserBalanceList(ctx, params)
	if err != nil {
		return nil, err
	}

	var resp []*entity_admin.RespUserBalance
	for _, user := range users {
		resp = append(resp, new(entity_admin.RespUserBalance).Fill(user))
	}

	return resp, nil
}
