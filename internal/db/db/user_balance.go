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

func (r *Repo) GetUserBalances(ctx context.Context, accountName string) ([]UserBalanceWithLock, error) {
	var userBalances []UserBalance
	if err := r.WithContext(ctx).Where("account = ?", accountName).Find(&userBalances).Error; err != nil {
		return nil, err
	}
	userBalanceMap := make(map[string]decimal.Decimal)
	for _, balance := range userBalances {
		userBalanceMap[balance.Coin] = balance.Balance
	}

	openOrders, err := r.GetOpenOrderByTrader(ctx, accountName)
	if err != nil {
		return nil, err
	}
	lockedCoins := make(map[string][]*UserPoolBalance)
	for _, order := range openOrders {
		if order.IsBid {
			poolBalances, ok := lockedCoins[order.PoolQuoteCoin]
			if !ok {
				lockedCoins[order.PoolQuoteCoin] = append(lockedCoins[order.PoolQuoteCoin], &UserPoolBalance{
					PoolID:     order.PoolID,
					PoolSymbol: order.PoolSymbol,
					Balance:    order.OriginalQuantity.Sub(order.ExecutedQuantity).Mul(order.Price),
				})
			} else {
				found := false
				for _, poolBalance := range poolBalances {
					if poolBalance.PoolID == order.PoolID {
						poolBalance.Balance = poolBalance.Balance.Add(order.OriginalQuantity.Sub(order.ExecutedQuantity).Mul(order.Price))
						found = true
					}
				}
				if !found {
					lockedCoins[order.PoolQuoteCoin] = append(lockedCoins[order.PoolQuoteCoin], &UserPoolBalance{
						PoolID:     order.PoolID,
						PoolSymbol: order.PoolSymbol,
						Balance:    order.OriginalQuantity.Sub(order.ExecutedQuantity).Mul(order.Price),
					})
				}
			}
			
		} else {
			poolBalances, ok := lockedCoins[order.PoolBaseCoin]
			if !ok {
				lockedCoins[order.PoolBaseCoin] = append(lockedCoins[order.PoolBaseCoin], &UserPoolBalance{
					PoolID:     order.PoolID,
					PoolSymbol: order.PoolSymbol,
					Balance:    order.OriginalQuantity.Sub(order.ExecutedQuantity),
				})
			} else {
				found := false
				for _, poolBalance := range poolBalances {
					if poolBalance.PoolID == order.PoolID {
						poolBalance.Balance = poolBalance.Balance.Add(order.OriginalQuantity.Sub(order.ExecutedQuantity))
						found = true
					}
				}
				if !found {
					lockedCoins[order.PoolBaseCoin] = append(lockedCoins[order.PoolBaseCoin], &UserPoolBalance{
						PoolID:     order.PoolID,
						PoolSymbol: order.PoolSymbol,
						Balance:    order.OriginalQuantity.Sub(order.ExecutedQuantity),
					})
				}
			}
		}
	}

	result := []UserBalanceWithLock{}
	for coin, balances := range userBalanceMap {
		var totalLocked decimal.Decimal
		if lockedCoin, ok := lockedCoins[coin]; ok {
			for _, balance := range lockedCoin {
				totalLocked = totalLocked.Add(balance.Balance)
			}
		}
		result = append(result, UserBalanceWithLock{
			UserBalance: UserBalance{
				Account: accountName,
				Coin:    coin,
				Balance: balances,
			},
			Locked:      totalLocked,
			PoolBalance: lockedCoins[coin],
		})
	}

	for coin, lockedCoin := range lockedCoins {
		if _, ok := userBalanceMap[coin]; ok {
			continue
		}
		var totalLocked decimal.Decimal
		for _, balance := range lockedCoin {
			totalLocked = totalLocked.Add(balance.Balance)
		}
		result = append(result, UserBalanceWithLock{
			UserBalance: UserBalance{
				Account: accountName,
				Coin:    coin,
				Balance: decimal.Zero,
			},
			Locked:      totalLocked,
			PoolBalance: lockedCoins[coin],
		})
	}
	return result, nil
}

func (r *Repo) DeleteUserBalance(ctx context.Context, account string) error {
	return r.WithContext(ctx).Where("account = ?", account).Unscoped().Delete(&UserBalance{}).Error
}
