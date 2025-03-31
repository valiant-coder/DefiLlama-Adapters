package marketplace

import (
	"context"
	"exapp-go/internal/entity"

	"github.com/shopspring/decimal"
)

func (s *UserService) GetUserSubaccountBalances(ctx context.Context, eosAccount, permission string) ([]entity.SubAccountBalance, error) {

	userAvailableBalances, err := s.getPasskeyUserAvailableBalances(ctx, eosAccount)
	if err != nil {
		return nil, err
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

		if !isVisibleToken(coin, poolTokens, allTokens) {
			continue
		}

		// Create subaccount balance
		var subAccountBalance entity.SubAccountBalance
		subAccountBalance.Coin = coin
		subAccountBalance.Balance = availBalance.Balance.String()

		// Process locked amounts
		userBalanceWithLock := processBalanceWithLocks(
			eosAccount,
			coin,
			availBalance.Balance,
			lockedCoins,
			make(map[string]decimal.Decimal), // No depositing for subaccounts
			make(map[string]decimal.Decimal), // No withdrawing for subaccounts
		)
		subAccountBalance.Locked = userBalanceWithLock.Locked.String()
		subAccountBalance.Locks = convertPoolBalancesToLocks(userBalanceWithLock.PoolBalance)

		// Set USDT price
		setUSDTPrice(&subAccountBalance, coin, coinUSDTPrice)

		result = append(result, subAccountBalance)
	}

	// Process coins that only have locked amounts
	for coin := range lockedCoins {
		if processedCoins[coin] {
			continue // Already processed
		}

		if !isVisibleToken(coin, poolTokens, allTokens) {
			continue
		}

		// Create subaccount balance for coins with only locked amounts
		var subAccountBalance entity.SubAccountBalance
		subAccountBalance.Coin = coin
		subAccountBalance.Balance = "0"

		// Process locked amounts
		userBalanceWithLock := processBalanceWithLocks(
			eosAccount,
			coin,
			decimal.Zero,
			lockedCoins,
			make(map[string]decimal.Decimal), // No depositing for subaccounts
			make(map[string]decimal.Decimal), // No withdrawing for subaccounts
		)
		subAccountBalance.Locked = userBalanceWithLock.Locked.String()
		subAccountBalance.Locks = convertPoolBalancesToLocks(userBalanceWithLock.PoolBalance)

		// Set USDT price
		setUSDTPrice(&subAccountBalance, coin, coinUSDTPrice)

		result = append(result, subAccountBalance)
	}

	return result, nil
}
