package marketplace

import (
	"context"
	"errors"
	"exapp-go/internal/entity"
)

func (s *UserService) GetUserBalance(ctx context.Context, accountName string) ([]entity.UserBalance, error) {
	if accountName == "" {
		return nil, errors.New("account is required")
	}
	userBalances, err := s.db.GetUserBalances(ctx, accountName)
	if err != nil {
		return nil, err
	}

	var result []entity.UserBalance
	for _, ub := range userBalances {
		var userBalance entity.UserBalance
		userBalance.Coin = ub.Coin
		userBalance.Balance = ub.Balance.String()
		userBalance.Locked = ub.Locked.String()
		userBalance.Locks = []entity.LockBalance{}
		for _, poolBalance := range ub.PoolBalance {
			userBalance.Locks = append(userBalance.Locks, entity.LockBalance{
				PoolID:     poolBalance.PoolID,
				PoolSymbol: poolBalance.PoolSymbol,
				Balance:    poolBalance.Balance.String(),
			})
		}
		result = append(result, userBalance)
	}
	return result, nil
}
