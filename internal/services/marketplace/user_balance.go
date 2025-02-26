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
	"time"

	"github.com/shopspring/decimal"
)

func (s *UserService) getCoinUSDTPrice(ctx context.Context) (map[string]string, error) {
	if !s.priceCacheTime.IsZero() && time.Since(s.priceCacheTime) < 5*time.Second {
		return s.priceCache, nil
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

	s.priceCache = coinUSDTPrice
	s.priceCacheTime = time.Now()

	return coinUSDTPrice, nil
}

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

	allTokens, err := s.repo.GetAllTokens(ctx)
	if err != nil {
		return nil, err
	}

	poolTokens, err := s.repo.GetVisiblePoolTokens(ctx)
	if err != nil {
		return nil, err
	}
	userBalances, err := s.repo.GetUserBalances(ctx, accountName, userAvailableBalances, allTokens, poolTokens)
	if err != nil {
		return nil, err
	}

	coinUSDTPrice, err := s.getCoinUSDTPrice(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]entity.UserBalance, 0)
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
		userBalance.Depositing = ub.Depositing.String()
		userBalance.Withdrawing = ub.Withdrawing.String()
		userBalance.Locks = make([]entity.LockBalance, 0)
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

func (s *UserService) CalculateUserUSDTBalance(ctx context.Context, accountName string) (decimal.Decimal, error) {
	balances, err := s.GetUserBalance(ctx, accountName)
	if err != nil {
		return decimal.Zero, err
	}

	total := decimal.Zero
	for _, balance := range balances {
		if balance.USDTPrice == "" || balance.Balance == "" {
			continue
		}
		price, err := decimal.NewFromString(balance.USDTPrice)
		if err != nil {
			continue
		}
		amount, err := decimal.NewFromString(balance.Balance)
		if err != nil {
			continue
		}
		locked, err := decimal.NewFromString(balance.Locked)
		if err != nil {
			continue
		}
		total = total.Add(price.Mul(amount.Add(locked)))
	}
	return total, nil
}
