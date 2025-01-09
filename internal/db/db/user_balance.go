package db

import (
	"context"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func init() {
	addMigrateFunc(func(r *Repo) error {
		return r.AutoMigrate(&UserBalance{})
	})
}

type UserPoolBalance struct {
	PoolID     uint64          `json:"pool_id"`
	PoolSymbol string          `json:"pool_symbol"`
	Balance    decimal.Decimal `json:"balance"`
}

type UserBalance struct {
	gorm.Model
	Account string `gorm:"column:account;type:varchar(255);not null;uniqueIndex:idx_account_coin"`
	// contract-symbol
	Coin    string          `gorm:"column:coin;type:varchar(255);not null;uniqueIndex:idx_account_coin"`
	Balance decimal.Decimal `gorm:"column:balance;type:decimal(36,18);not null;"`
}

func (UserBalance) TableName() string {
	return "user_balances"
}

func (r *Repo) UpsertUserBalance(ctx context.Context, userBalance *UserBalance) error {
	return r.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "account"}, {Name: "coin"}},
		DoUpdates: clause.AssignmentColumns([]string{"balance"}),
	}).Create(userBalance).Error
}

type UserBalanceWithLock struct {
	UserBalance
	Locked      decimal.Decimal
	PoolBalance []*UserPoolBalance
}

func (r *Repo) updateLockedCoins(lockedCoins map[string][]*UserPoolBalance, coin string, poolID uint64, poolSymbol string, lockedAmount decimal.Decimal) {
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

func (r *Repo) calculateOrderLock(order *OpenOrder) (string, decimal.Decimal) {
	if order.IsBid {
		return order.PoolQuoteCoin, order.OriginalQuantity.Sub(order.ExecutedQuantity).Mul(order.Price)
	}
	return order.PoolBaseCoin, order.OriginalQuantity.Sub(order.ExecutedQuantity)
}

func (r *Repo) GetUserBalances(ctx context.Context, accountName string) ([]UserBalanceWithLock, error) {
	// 1. Get user balances
	var userBalances []UserBalance
	if err := r.WithContext(ctx).Where("account = ?", accountName).Find(&userBalances).Error; err != nil {
		return nil, err
	}

	// Build balance mapping
	userBalanceMap := make(map[string]decimal.Decimal, len(userBalances))
	for _, balance := range userBalances {
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
		coin, lockedAmount := r.calculateOrderLock(order)
		r.updateLockedCoins(lockedCoins, coin, order.PoolID, order.PoolSymbol, lockedAmount)
	}

	// 4. Build result
	result := make([]UserBalanceWithLock, 0, len(userBalanceMap)+len(lockedCoins))

	// Handle coins with existing balances
	for coin, balance := range userBalanceMap {
		var totalLocked decimal.Decimal
		poolBalances := lockedCoins[coin]
		for _, pb := range poolBalances {
			totalLocked = totalLocked.Add(pb.Balance)
		}

		result = append(result, UserBalanceWithLock{
			UserBalance: UserBalance{
				Account: accountName,
				Coin:    coin,
				Balance: balance,
			},
			Locked:      totalLocked,
			PoolBalance: poolBalances,
		})
	}

	// Handle coins with only locked amounts
	for coin, poolBalances := range lockedCoins {
		if _, exists := userBalanceMap[coin]; exists {
			continue
		}

		var totalLocked decimal.Decimal
		for _, pb := range poolBalances {
			totalLocked = totalLocked.Add(pb.Balance)
		}

		result = append(result, UserBalanceWithLock{
			UserBalance: UserBalance{
				Account: accountName,
				Coin:    coin,
				Balance: decimal.Zero,
			},
			Locked:      totalLocked,
			PoolBalance: poolBalances,
		})
	}

	return result, nil
}

func (r *Repo) DeleteUserBalance(ctx context.Context, account string) error {
	return r.WithContext(ctx).Where("account = ?", account).Unscoped().Delete(&UserBalance{}).Error
}
