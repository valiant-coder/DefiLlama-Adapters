package marketplace

import (
	"context"
	"exapp-go/config"
	"exapp-go/internal/entity"
	"exapp-go/pkg/onedex"
	"fmt"
	"strings"

	"github.com/eoscanada/eos-go"
	"github.com/shopspring/decimal"
)

func (s *UserService) GetUserSubaccountBalances(ctx context.Context, eosAccount, permission string) ([]entity.SubAccountBalance, error) {
	// Initialize EOS API client
	api := eos.New(config.Conf().Eos.NodeURL)

	// Fetch subaccount balances from OneDex
	balancesResp, err := onedex.GetSubaccountBalances(ctx, api, eosAccount, permission)
	if err != nil {
		return nil, fmt.Errorf("failed to get subaccount balances: %w", err)
	}

	// Convert to UserBalance format for processing
	var userAvailableBalances []UserBalance
	for _, balance := range balancesResp.Balances {
		userBalance := UserBalance{
			Account: eosAccount,
			Coin:    fmt.Sprintf("%s-%s", balance.Contract, balance.Symbol),
			Balance: balance.Amount,
		}
		userAvailableBalances = append(userAvailableBalances, userBalance)
	}

	// Get token metadata
	allTokens, err := s.repo.GetAllTokens(ctx)
	if err != nil {
		return nil, err
	}

	poolTokens, err := s.repo.GetVisiblePoolTokens(ctx)
	if err != nil {
		return nil, err
	}

	// Get open orders to calculate locked amounts
	openOrders, err := s.repo.GetOpenOrdersByTraderPermission(ctx, eosAccount, permission)
	if err != nil {
		return nil, err
	}

	// Calculate locked amounts by pool
	lockedCoins := make(map[string][]*UserPoolBalance)
	for _, order := range openOrders {
		coin, lockedAmount := determineCoinAndLockedAmount(order)
		addOrUpdateLockedCoinPool(lockedCoins, coin, order.PoolID, order.PoolSymbol, lockedAmount)
	}

	// Build user balances map for quick lookup
	userBalanceMap := make(map[string]decimal.Decimal, len(userAvailableBalances))
	for _, balance := range userAvailableBalances {
		userBalanceMap[balance.Coin] = balance.Balance
	}

	// Get current USDT prices for all coins
	coinUSDTPrice, err := s.fetchCoinUSDTPriceWithCache(ctx)
	if err != nil {
		return nil, err
	}

	// Build result
	result := make([]entity.SubAccountBalance, 0)

	// Process coins with available balances
	processedCoins := make(map[string]bool)
	for _, availBalance := range userAvailableBalances {
		coin := availBalance.Coin
		processedCoins[coin] = true

		var totalLocked decimal.Decimal
		poolBalances := lockedCoins[coin]
		if poolBalances == nil {
			poolBalances = []*UserPoolBalance{}
		}

		// Calculate total locked amount
		for _, pb := range poolBalances {
			totalLocked = totalLocked.Add(pb.Balance)
		}

		// Only include visible tokens
		if _, exists := poolTokens[coin]; !exists {
			if _, exists := allTokens[coin]; !exists {
				continue
			}
		}

		// Create subaccount balance
		var subAccountBalance entity.SubAccountBalance
		subAccountBalance.Coin = coin
		subAccountBalance.Balance = availBalance.Balance.String()
		subAccountBalance.Locked = totalLocked.String()

		// Set USDT price if available
		parts := strings.Split(coin, "-")
		if len(parts) == 2 {
			symbol := parts[1]
			if price, ok := coinUSDTPrice[symbol]; ok {
				subAccountBalance.USDTPrice = price
			}
			if strings.Contains(coin, "USDT") {
				subAccountBalance.USDTPrice = "1" // USDT price is always 1 USDT
			}
		}

		// Convert pool balances to locks
		subAccountBalance.Locks = make([]entity.LockBalance, 0, len(poolBalances))
		for _, poolBalance := range poolBalances {
			subAccountBalance.Locks = append(subAccountBalance.Locks, entity.LockBalance{
				PoolID:     poolBalance.PoolID,
				PoolSymbol: poolBalance.PoolSymbol,
				Balance:    poolBalance.Balance.String(),
			})
		}

		result = append(result, subAccountBalance)
	}

	// Process coins that only have locked amounts
	for coin, poolBalances := range lockedCoins {
		if processedCoins[coin] {
			continue // Already processed
		}

		// Only include visible tokens
		if _, exists := poolTokens[coin]; !exists {
			if _, exists := allTokens[coin]; !exists {
				continue
			}
		}

		var totalLocked decimal.Decimal
		for _, pb := range poolBalances {
			totalLocked = totalLocked.Add(pb.Balance)
		}

		// Create subaccount balance for coins with only locked amounts
		var subAccountBalance entity.SubAccountBalance
		subAccountBalance.Coin = coin
		subAccountBalance.Balance = "0"
		subAccountBalance.Locked = totalLocked.String()

		// Set USDT price if available
		parts := strings.Split(coin, "-")
		if len(parts) == 2 {
			symbol := parts[1]
			if price, ok := coinUSDTPrice[symbol]; ok {
				subAccountBalance.USDTPrice = price
			}
			if strings.Contains(coin, "USDT") {
				subAccountBalance.USDTPrice = "1" // USDT price is always 1 USDT
			}
		}

		// Convert pool balances to locks
		subAccountBalance.Locks = make([]entity.LockBalance, 0, len(poolBalances))
		for _, poolBalance := range poolBalances {
			subAccountBalance.Locks = append(subAccountBalance.Locks, entity.LockBalance{
				PoolID:     poolBalance.PoolID,
				PoolSymbol: poolBalance.PoolSymbol,
				Balance:    poolBalance.Balance.String(),
			})
		}

		result = append(result, subAccountBalance)
	}

	return result, nil
}
