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

	"exapp-go/pkg/eth"

	"github.com/shopspring/decimal"
)

// ----- Price Cache Methods -----

// fetchCoinUSDTPriceWithCache retrieves USDT prices for various coins with a 5-second cache
func (s *UserService) fetchCoinUSDTPriceWithCache(ctx context.Context) (map[string]string, error) {
	// Return cached prices if less than 5 seconds old
	if !s.priceCacheTime.IsZero() && time.Since(s.priceCacheTime) < 5*time.Second {
		return s.priceCache, nil
	}

	// Fetch fresh data from repository
	poolStatuses, err := s.ckhRepo.ListPoolStats(ctx)
	if err != nil {
		return nil, err
	}

	// Extract USDT prices
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

	// Update cache
	s.priceCache = coinUSDTPrice
	s.priceCacheTime = time.Now()

	return coinUSDTPrice, nil
}

// ----- Data Models -----

// UserPoolBalance represents a user's balance in a specific trading pool
type UserPoolBalance struct {
	PoolID     uint64          `json:"pool_id"`
	PoolSymbol string          `json:"pool_symbol"`
	Balance    decimal.Decimal `json:"balance"`
}

// UserBalance represents a user's balance for a specific coin
type UserBalance struct {
	Account string `gorm:"column:account;type:varchar(255);not null;uniqueIndex:idx_account_coin"`
	// contract-symbol format (e.g., "eosio.token-EOS")
	Coin    string          `gorm:"column:coin;type:varchar(255);not null;uniqueIndex:idx_account_coin"`
	Balance decimal.Decimal `gorm:"column:balance;type:decimal(36,18);not null;"`
}

// UserBalanceWithLock extends UserBalance with locked, depositing, and withdrawing amounts
type UserBalanceWithLock struct {
	UserBalance
	Locked      decimal.Decimal    // Total locked amount across all pools
	PoolBalance []*UserPoolBalance // Detailed breakdown of locked amounts by pool
	Depositing  decimal.Decimal    // Amount currently being deposited
	Withdrawing decimal.Decimal    // Amount currently being withdrawn
}

// ----- Helper Functions -----

// addOrUpdateLockedCoinPool adds or updates the locked amount for a specific coin and pool
func addOrUpdateLockedCoinPool(lockedCoins map[string][]*UserPoolBalance, coin string, poolID uint64, poolSymbol string, lockedAmount decimal.Decimal) {
	poolBalances, exists := lockedCoins[coin]
	if !exists {
		lockedCoins[coin] = []*UserPoolBalance{{
			PoolID:     poolID,
			PoolSymbol: poolSymbol,
			Balance:    lockedAmount,
		}}
		return
	}

	// Update existing pool balance if found
	for _, balance := range poolBalances {
		if balance.PoolID == poolID {
			balance.Balance = balance.Balance.Add(lockedAmount)
			return
		}
	}

	// Add new pool balance if not found
	lockedCoins[coin] = append(poolBalances, &UserPoolBalance{
		PoolID:     poolID,
		PoolSymbol: poolSymbol,
		Balance:    lockedAmount,
	})
}

// determineCoinAndLockedAmount calculates which coin is locked and how much based on order type
func determineCoinAndLockedAmount(order *db.OpenOrder) (string, decimal.Decimal) {
	if order.IsBid {
		// For buy orders, quote coin (e.g., USDT) is locked
		return order.PoolQuoteCoin, order.OriginalQuantity.Sub(order.ExecutedQuantity).Mul(order.Price).Round(int32(order.QuoteCoinPrecision))
	}
	// For sell orders, base coin (e.g., BTC) is locked
	return order.PoolBaseCoin, order.OriginalQuantity.Sub(order.ExecutedQuantity)
}

// ----- Shared Helper Functions -----

// processBalanceWithLocks processes a balance with its locked amounts and returns a UserBalanceWithLock
func processBalanceWithLocks(
	account string,
	coin string,
	balance decimal.Decimal,
	lockedCoins map[string][]*UserPoolBalance,
	depositingBalance map[string]decimal.Decimal,
	withdrawingBalance map[string]decimal.Decimal,
) UserBalanceWithLock {
	var totalLocked decimal.Decimal
	poolBalances := lockedCoins[coin]
	if poolBalances == nil {
		poolBalances = []*UserPoolBalance{}
	}

	// Calculate total locked amount across all pools
	for _, pb := range poolBalances {
		totalLocked = totalLocked.Add(pb.Balance)
	}

	return UserBalanceWithLock{
		UserBalance: UserBalance{
			Account: account,
			Coin:    coin,
			Balance: balance,
		},
		Locked:      totalLocked,
		PoolBalance: poolBalances,
		Depositing:  depositingBalance[coin],
		Withdrawing: withdrawingBalance[coin],
	}
}

// isVisibleToken checks if a token is visible based on pool tokens and all tokens maps
func isVisibleToken(coin string, poolTokens, allTokens map[string]string) bool {
	if _, exists := poolTokens[coin]; exists {
		return true
	}
	if _, exists := allTokens[coin]; exists {
		return true
	}
	return false
}

// setUSDTPrice sets the USDT price for a coin in the given balance
func setUSDTPrice(balance interface{}, coin string, coinUSDTPrice map[string]string) {
	parts := strings.Split(coin, "-")
	if len(parts) != 2 {
		return
	}

	symbol := parts[1]
	if price, ok := coinUSDTPrice[symbol]; ok {
		switch b := balance.(type) {
		case *entity.UserBalance:
			b.USDTPrice = price
		case *entity.SubAccountBalance:
			b.USDTPrice = price
		}
	}
	if strings.Contains(coin, "USDT") {
		switch b := balance.(type) {
		case *entity.UserBalance:
			b.USDTPrice = "1"
		case *entity.SubAccountBalance:
			b.USDTPrice = "1"
		}
	}
}

// convertPoolBalancesToLocks converts UserPoolBalance to entity.LockBalance
func convertPoolBalancesToLocks(poolBalances []*UserPoolBalance) []entity.LockBalance {
	locks := make([]entity.LockBalance, 0, len(poolBalances))
	for _, poolBalance := range poolBalances {
		locks = append(locks, entity.LockBalance{
			PoolID:     poolBalance.PoolID,
			PoolSymbol: poolBalance.PoolSymbol,
			Balance:    poolBalance.Balance.String(),
		})
	}
	return locks
}

// ----- Main Balance Functions -----

// fetchUserDetailedBalances retrieves detailed balance information for a user
func (s *UserService) fetchUserDetailedBalances(
	ctx context.Context,
	isEvmUser bool,
	uid, eosAccount, permission string,
	userAvailableBalances []UserBalance,
	allTokens, poolTokens map[string]string) ([]UserBalanceWithLock, error) {

	// 1. Build a map of available balances for quick lookup
	userBalanceMap := make(map[string]decimal.Decimal, len(userAvailableBalances))
	for _, balance := range userAvailableBalances {
		userBalanceMap[balance.Coin] = balance.Balance
	}

	var err error
	var openOrders []*db.OpenOrder
	if !isEvmUser {
		openOrders, err = s.repo.GetOpenOrdersByTrader(ctx, eosAccount)
	} else {
		openOrders, err = s.repo.GetOpenOrdersByTraderPermission(ctx, eosAccount, permission)
	}
	if err != nil {
		return nil, err
	}

	// 3. Calculate locked amounts by pool
	lockedCoins := make(map[string][]*UserPoolBalance)
	for _, order := range openOrders {
		coin, lockedAmount := determineCoinAndLockedAmount(order)
		addOrUpdateLockedCoinPool(lockedCoins, coin, order.PoolID, order.PoolSymbol, lockedAmount)
	}

	// 4. Get pending deposit records
	depositingRecords, err := s.repo.GetPendingDepositRecords(ctx, uid)
	if err != nil {
		return nil, err
	}

	// Calculate depositing balances
	depositingBalance := make(map[string]decimal.Decimal)
	for _, record := range depositingRecords {
		coin := fmt.Sprintf("%s-%s", config.Conf().Eos.OneDex.TokenContract, record.Symbol)
		if _, exists := userBalanceMap[coin]; exists {
			depositingBalance[coin] = depositingBalance[coin].Add(record.Amount)
		} else {
			depositingBalance[coin] = record.Amount
		}
	}

	// 5. Get pending withdraw records
	withdrawingRecords, err := s.repo.GetPendingWithdrawRecords(ctx, uid)
	if err != nil {
		return nil, err
	}

	// Calculate withdrawing balances
	withdrawingBalance := make(map[string]decimal.Decimal)
	for _, record := range withdrawingRecords {
		coin := fmt.Sprintf("%s-%s", config.Conf().Eos.OneDex.TokenContract, record.Symbol)
		if _, exists := userBalanceMap[coin]; exists {
			withdrawingBalance[coin] = withdrawingBalance[coin].Add(record.Amount)
		} else {
			withdrawingBalance[coin] = record.Amount
		}
	}

	// 6. Build user balances result
	userBalances := make([]UserBalanceWithLock, 0, len(userBalanceMap)+len(lockedCoins))

	// 7. Process coins with existing balances
	for coin, balance := range userBalanceMap {
		if !isVisibleToken(coin, poolTokens, allTokens) {
			continue
		}

		userBalances = append(userBalances, processBalanceWithLocks(
			eosAccount,
			coin,
			balance,
			lockedCoins,
			depositingBalance,
			withdrawingBalance,
		))
	}

	// 8. Process coins with only locked amounts (no available balance)
	for coin := range lockedCoins {
		if _, exists := userBalanceMap[coin]; exists {
			continue // Already processed in previous step
		}

		if !isVisibleToken(coin, poolTokens, allTokens) {
			continue
		}

		userBalances = append(userBalances, processBalanceWithLocks(
			eosAccount,
			coin,
			decimal.Zero,
			lockedCoins,
			depositingBalance,
			withdrawingBalance,
		))
	}

	return userBalances, nil
}

// FetchUserBalanceByUID retrieves all user balances by user ID
func (s *UserService) FetchUserBalanceByUID(ctx context.Context, uid string) ([]entity.UserBalance, error) {
	// Validate input
	if uid == "" {
		return nil, errors.New("uid is required")
	}

	user, err := s.repo.GetUser(ctx, uid)
	if err != nil {
		return nil, err
	}

	var eosAccount string
	var userAvailableBalances []UserBalance
	var isEvmUser bool
	if user.LoginMethod == db.LoginMethodEVM {
		isEvmUser = true
		eosAccount = user.EOSAccount
		userAvailableBalances, err = s.getEvmUserAvailableBalances(ctx, user.EVMAddress)
		if err != nil {
			return nil, err
		}
	} else {
		eosAccount, err = s.repo.GetEosAccountByUID(ctx, uid)
		if err != nil {
			return nil, err
		}
		userAvailableBalances, err = s.getPasskeyUserAvailableBalances(ctx, eosAccount)
		if err != nil {
			return nil, err
		}
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

	// Fetch detailed balances including locked amounts
	userBalances, err := s.fetchUserDetailedBalances(
		ctx,
		isEvmUser,
		uid,
		eosAccount,
		user.Permission,
		userAvailableBalances,
		allTokens,
		poolTokens)
	if err != nil {
		return nil, err
	}

	// Get current USDT prices for all coins
	coinUSDTPrice, err := s.fetchCoinUSDTPriceWithCache(ctx)
	if err != nil {
		return nil, err
	}

	// Convert to entity.UserBalance format for API response
	result := make([]entity.UserBalance, 0, len(userBalances))
	for _, ub := range userBalances {
		// Create user balance entity
		var userBalance entity.UserBalance
		userBalance.Coin = ub.Coin
		setUSDTPrice(&userBalance, ub.Coin, coinUSDTPrice)

		// Convert decimal values to strings for JSON response
		userBalance.Balance = ub.Balance.String()
		userBalance.Locked = ub.Locked.String()
		userBalance.Depositing = ub.Depositing.String()
		userBalance.Withdrawing = ub.Withdrawing.String()

		// Convert pool balances to entity format
		userBalance.Locks = convertPoolBalancesToLocks(ub.PoolBalance)

		result = append(result, userBalance)
	}

	return result, nil
}

func (s *UserService) getPasskeyUserAvailableBalances(ctx context.Context, accountName string) ([]UserBalance, error) {
	// Fetch on-chain token balances using Hyperion API
	hyperionCfg := config.Conf().Eos.Hyperion
	hyperionClient := hyperion.NewClient(hyperionCfg.Endpoint)
	tokens, err := hyperionClient.GetTokens(ctx, accountName)
	if err != nil {
		log.Printf("Get tokens failed: %v-%v", accountName, err)
		return nil, err
	}

	// Convert token balances to UserBalance format
	var userAvailableBalances []UserBalance
	for _, token := range tokens {
		userBalance := UserBalance{
			Account: accountName,
			Coin:    fmt.Sprintf("%s-%s", token.Contract, token.Symbol),
			Balance: token.Amount,
		}
		userAvailableBalances = append(userAvailableBalances, userBalance)
	}

	return userAvailableBalances, nil
}

func (s *UserService) getEvmUserAvailableBalances(ctx context.Context, evmAddress string) ([]UserBalance, error) {

	ethscanCfg := config.Conf().Evm.Ethscan
	ethscanClient := eth.NewEthScanClient(ethscanCfg.Endpoint, ethscanCfg.ApiKey)

	tokenBalances, err := ethscanClient.GetTokenBalancesByAddress(ctx, evmAddress)
	if err != nil {
		return nil, err
	}

	// Get all tokens from the database to map EVM addresses to EOS contract-symbol format
	tokens, err := s.repo.ListTokens(ctx)
	if err != nil {
		return nil, err
	}

	// Create a map of EVM addresses to tokens for quick lookup
	// Using lowercase addresses to make the lookup case-insensitive
	evmToTokenMap := make(map[string]*db.Token)
	for i := range tokens {
		if tokens[i].ExsatTokenAddress != "" {
			// Store with lowercase key for case-insensitive lookup
			evmToTokenMap[strings.ToLower(tokens[i].ExsatTokenAddress)] = &tokens[i]
		}
	}

	var userAvailableBalances []UserBalance
	for _, tokenBalance := range tokenBalances {
		// Look up the token in our map using lowercase address for case-insensitive matching
		token, exists := evmToTokenMap[strings.ToLower(tokenBalance.TokenAddress)]
		if !exists {
			// Skip tokens that don't have a mapping in our database
			continue
		}

		userBalance := UserBalance{
			Account: evmAddress,
			Coin:    fmt.Sprintf("%s-%s", token.EOSContractAddress, token.Symbol),
			Balance: tokenBalance.Balance,
		}
		userAvailableBalances = append(userAvailableBalances, userBalance)
	}

	return userAvailableBalances, nil
}

// CalculateTotalUSDTValueForUser calculates the total value of all user assets in USDT
func (s *UserService) CalculateTotalUSDTValueForUser(ctx context.Context, uid string) (decimal.Decimal, error) {
	// Get all user balances
	balances, err := s.FetchUserBalanceByUID(ctx, uid)
	if err != nil {
		return decimal.Zero, err
	}

	// Calculate total value
	total := decimal.Zero
	for _, balance := range balances {
		// Skip coins without price or balance
		if balance.USDTPrice == "" || balance.Balance == "" {
			continue
		}

		// Parse price and amounts
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

		// Add to total (both available and locked balances)
		total = total.Add(price.Mul(amount.Add(locked)))
	}

	return total, nil
}
