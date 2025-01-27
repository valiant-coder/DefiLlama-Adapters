package marketplace

import (
	"context"
	"errors"
	"exapp-go/config"
	"exapp-go/internal/db/db"
	"exapp-go/internal/entity"
	"exapp-go/pkg/hyperion"
	"fmt"
	"log"
	"strings"
)

func (s *UserService) GetUserBalance(ctx context.Context, accountName string) ([]entity.UserBalance, error) {
	if accountName == "" {
		return nil, errors.New("account is required")
	}

	hyperionCfg := config.Conf().Eos.Hyperion
	hyperionClient := hyperion.NewClient(hyperionCfg.Endpoint)
	tokens, err := hyperionClient.GetTokens(ctx, accountName)
	if err != nil {
		log.Printf("Get tokens failed: %v-%v", accountName, err)
		return nil, err
	}

	var userAvailableBalances []db.UserBalance
	for _, token := range tokens {
		userBalance := db.UserBalance{
			Account: accountName,
			Coin:    fmt.Sprintf("%s-%s", token.Contract, token.Symbol),
			Balance: token.Amount,
		}
		userAvailableBalances = append(userAvailableBalances, userBalance)
	}

	userBalances, err := s.db.GetUserBalances(ctx, accountName, userAvailableBalances)
	if err != nil {
		return nil, err
	}

	poolStatuses, err := s.ckhRepo.ListPoolStats(ctx)
	if err != nil {
		return nil, err
	}

	coinUSDTPrice := make(map[string]string)

	for _, poolStatus := range poolStatuses {
		if strings.Contains(poolStatus.QuoteCoin, "USDT") {
			parts := strings.Split(poolStatus.BaseCoin, "-")
			if len(parts) != 2 {
				continue
			}
			coin := parts[1]
			coinUSDTPrice[coin] = poolStatus.LastPrice.String()
		}
	}

	var result []entity.UserBalance
	for _, ub := range userBalances {
		var userBalance entity.UserBalance
		parts := strings.Split(ub.Coin, "-")
		if len(parts) != 2 {
			continue
		}
		coin := parts[1]
		if price, ok := coinUSDTPrice[coin]; ok {
			userBalance.USDTPrice = price
		}
		if strings.Contains(ub.Coin, "USDT") {
			userBalance.USDTPrice = "1"
		}
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
