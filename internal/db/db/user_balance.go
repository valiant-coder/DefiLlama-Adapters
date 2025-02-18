package db

import (
	"context"
	"errors"
	"exapp-go/config"
	"fmt"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type UserPoolBalance struct {
	PoolID     uint64          `json:"pool_id"`
	PoolSymbol string          `json:"pool_symbol"`
	Balance    decimal.Decimal `json:"balance"`
}

type UserBalance struct {
	Account string `gorm:"column:account;type:varchar(255);not null;uniqueIndex:idx_account_coin"`
	// contract-symbol
	Coin    string          `gorm:"column:coin;type:varchar(255);not null;uniqueIndex:idx_account_coin"`
	Balance decimal.Decimal `gorm:"column:balance;type:decimal(36,18);not null;"`
}

type UserBalanceWithLock struct {
	UserBalance
	Locked      decimal.Decimal
	PoolBalance []*UserPoolBalance
	Depositing  decimal.Decimal
	Withdrawing decimal.Decimal
}

func updateLockedCoins(lockedCoins map[string][]*UserPoolBalance, coin string, poolID uint64, poolSymbol string, lockedAmount decimal.Decimal) {
	poolBalances, exists := lockedCoins[coin]
	if !exists {
		lockedCoins[coin] = []*UserPoolBalance{{
			PoolID:     poolID,
			PoolSymbol: poolSymbol,
			Balance:    lockedAmount,
		}}
		return
	}

	for _, balance := range poolBalances {
		if balance.PoolID == poolID {
			balance.Balance = balance.Balance.Add(lockedAmount)
			return
		}
	}

	lockedCoins[coin] = append(poolBalances, &UserPoolBalance{
		PoolID:     poolID,
		PoolSymbol: poolSymbol,
		Balance:    lockedAmount,
	})
}

func calculateOrderLock(order *OpenOrder) (string, decimal.Decimal) {
	if order.IsBid {
		return order.PoolQuoteCoin, order.OriginalQuantity.Sub(order.ExecutedQuantity).Mul(order.Price).Round(int32(order.QuoteCoinPrecision))
	}
	return order.PoolBaseCoin, order.OriginalQuantity.Sub(order.ExecutedQuantity)
}

func (r *Repo) GetUserBalances(ctx context.Context, accountName string, userAvailableBalances []UserBalance, allTokens, poolTokens map[string]string) ([]UserBalanceWithLock, error) {
	uid, err := r.GetUIDByEOSAccount(ctx, accountName)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return []UserBalanceWithLock{}, nil
		}
		return nil, err
	}

	// 1. Build balance mapping
	userBalanceMap := make(map[string]decimal.Decimal, len(userAvailableBalances))
	for _, balance := range userAvailableBalances {
		userBalanceMap[balance.Coin] = balance.Balance
	}

	// 2. Get open orders
	openOrders, err := r.GetOpenOrderByTrader(ctx, accountName)
	if err != nil {
		return nil, err
	}

	// 3. Calculate locked amounts
	lockedCoins := make(map[string][]*UserPoolBalance)
	for _, order := range openOrders {
		coin, lockedAmount := calculateOrderLock(order)
		updateLockedCoins(lockedCoins, coin, order.PoolID, order.PoolSymbol, lockedAmount)
	}

	// 4. Get pending deposit records
	depositingRecords, err := r.GetPendingDepositRecords(ctx, uid)
	if err != nil {
		return nil, err
	}

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
	withdrawingRecords, err := r.GetPendingWithdrawRecords(ctx, uid)
	if err != nil {
		return nil, err
	}
	withdrawingBalance := make(map[string]decimal.Decimal)
	for _, record := range withdrawingRecords {
		coin := fmt.Sprintf("%s-%s", config.Conf().Eos.OneDex.TokenContract, record.Symbol)
		if _, exists := userBalanceMap[coin]; exists {
			withdrawingBalance[coin] = withdrawingBalance[coin].Add(record.Amount)
		} else {
			withdrawingBalance[coin] = record.Amount
		}
	}
	// 6. Build userBalances
	userBalances := make([]UserBalanceWithLock, 0, len(userBalanceMap)+len(lockedCoins))

	// 7. Handle coins with existing balances
	for coin, balance := range userBalanceMap {
		var totalLocked decimal.Decimal
		poolBalances := lockedCoins[coin]
		if poolBalances == nil {
			poolBalances = []*UserPoolBalance{}
		}
		for _, pb := range poolBalances {
			totalLocked = totalLocked.Add(pb.Balance)
		}


		userBalances = append(userBalances, UserBalanceWithLock{
			UserBalance: UserBalance{
				Account: accountName,
				Coin:    coin,
				Balance: balance,
			},
			Locked:      totalLocked,
			PoolBalance: poolBalances,
			Depositing:  depositingBalance[coin],
			Withdrawing: withdrawingBalance[coin],
		})
	}

	// 8. Handle coins with only locked amounts
	for coin, poolBalances := range lockedCoins {
		if _, exists := userBalanceMap[coin]; exists {
			continue
		}

		var totalLocked decimal.Decimal
		for _, pb := range poolBalances {
			totalLocked = totalLocked.Add(pb.Balance)
		}

		

		userBalances = append(userBalances, UserBalanceWithLock{
			UserBalance: UserBalance{
				Account: accountName,
				Coin:    coin,
				Balance: decimal.Zero,
			},
			Locked:      totalLocked,
			PoolBalance: poolBalances,
			Depositing:  depositingBalance[coin],
			Withdrawing: withdrawingBalance[coin],
		})
	}

	// 9. Filter visible tokens
	var result []UserBalanceWithLock
	for _, balance := range userBalances {

		if _, exists := poolTokens[balance.Coin]; exists {
			result = append(result, balance)
			continue
		}
		if _, exists := allTokens[balance.Coin]; exists {
			result = append(result, balance)
		}
	}

	return result, nil
}
