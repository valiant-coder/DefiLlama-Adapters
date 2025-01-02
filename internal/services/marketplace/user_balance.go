package marketplace

import (
	"context"
	"exapp-go/internal/entity"
	"exapp-go/pkg/utils"
)

func (s *UserService) GetUserBalance(ctx context.Context, accountName string) ([]entity.UserBalance, error) {
	userBalances, err := s.db.GetUserBalances(ctx, accountName)
	if err != nil {
		return nil, err
	}
	var poolIDs []uint64
	for _, userBalance := range userBalances {
		for _, poolBalance := range userBalance.PoolBalances {
			poolIDs = append(poolIDs, poolBalance.PoolID)
		}
	}
	poolIDs = utils.RemoveDuplicateUint64(poolIDs)

	poolSymbols, err := s.db.GetPoolSymbolsByIDs(ctx, poolIDs)
	if err != nil {
		return nil, err
	}
	var result []entity.UserBalance
	for _, ub := range userBalances {
		var userBalance entity.UserBalance
		userBalance.Contract = ub.Contract
		userBalance.Symbol = ub.Symbol
		userBalance.Balance = ub.Balance.String()
		userBalance.Locked = ub.Locked.String()
		userBalance.Locks = []entity.LockBalance{}
		for _, poolBalance := range ub.PoolBalances {
			poolSymbol, ok := poolSymbols[poolBalance.PoolID]
			if !ok {
				continue
			}
			userBalance.Locks = append(userBalance.Locks, entity.LockBalance{
				PoolID:     poolBalance.PoolID,
				PoolSymbol: poolSymbol,
				Balance:    poolBalance.Balance.String(),
			})
		}
		result = append(result, userBalance)
	}
	return result, nil
}
