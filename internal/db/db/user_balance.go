package db

import (
	"context"
	"errors"
	"exapp-go/config"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func init() {
	addMigrateFunc(func(repo *Repo) error {
		return repo.AutoMigrate(
			&UserBalanceRecord{},
			&UserDayProfitRecord{},
			&UserAccumulatedProfitRecord{},
		)
	})
}

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

type UserBalanceRecord struct {
	gorm.Model
	Time       time.Time       `gorm:"column:time;type:timestamp;not null;index:idx_time"`
	Account    string          `gorm:"column:account;type:varchar(255);not null;"`
	UID        string          `gorm:"column:uid;type:varchar(255);not null;index:idx_uid"`
	USDTAmount decimal.Decimal `gorm:"column:usdt_amount;type:decimal(20,6);not null;"`
}

func (t *UserBalanceRecord) TableName() string {
	return "user_balance_records"
}

func (r *Repo) CreateUserBalanceRecord(ctx context.Context, record *UserBalanceRecord) error {
	return r.DB.WithContext(ctx).Create(record).Error
}

func (r *Repo) GetUserBalanceRecordsInTimeRange(ctx context.Context, uid string, beginTime, endTime time.Time) ([]UserBalanceRecord, error) {
	var records []UserBalanceRecord
	err := r.DB.WithContext(ctx).
		Where("uid = ? AND time >= ? AND time <= ?", uid, beginTime, endTime).
		Order("time ASC").
		Find(&records).Error
	return records, err
}

func (r *Repo) GetUserBalanceRecordsByTimeRange(ctx context.Context, startTime, endTime time.Time) ([]UserBalanceRecord, error) {
	var records []UserBalanceRecord
	err := r.DB.WithContext(ctx).
		Where("time >= ? AND time < ?", startTime, endTime).
		Order("time ASC").
		Find(&records).Error
	return records, err
}

type UserDayProfitRecord struct {
	gorm.Model
	Time    time.Time       `gorm:"column:time;type:timestamp;not null;uniqueIndex:idx_uid_time"`
	Account string          `gorm:"column:account;type:varchar(255);not null;"`
	UID     string          `gorm:"column:uid;type:varchar(255);not null;uniqueIndex:idx_uid_time"`
	Profit  decimal.Decimal `gorm:"column:profit;type:decimal(20,6);not null;"`
}

func (t *UserDayProfitRecord) TableName() string {
	return "user_day_profit_records"
}

func (r *Repo) CreateUserDayProfitRecord(ctx context.Context, record *UserDayProfitRecord) error {
	return r.DB.WithContext(ctx).Create(record).Error
}



func (r *Repo) UpsertUserDayProfitRecord(ctx context.Context, record *UserDayProfitRecord) error {
	return r.DB.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "uid"}, {Name: "time"}},
			DoUpdates: clause.AssignmentColumns([]string{"profit", "updated_at"}),
		}).
		Create(record).Error
}

type UserAccumulatedProfitRecord struct {
	gorm.Model
	BeginTime time.Time       `gorm:"column:begin_time;type:timestamp;not null;uniqueIndex:idx_uid_begin_end_time"`
	EndTime   time.Time       `gorm:"column:end_time;type:timestamp;not null;uniqueIndex:idx_uid_begin_end_time"`
	Account   string          `gorm:"column:account;type:varchar(255);not null;"`
	UID       string          `gorm:"column:uid;type:varchar(255);not null;uniqueIndex:idx_uid_begin_end_time"`
	Profit    decimal.Decimal `gorm:"column:profit;type:decimal(20,6);not null;"`
}

func (t *UserAccumulatedProfitRecord) TableName() string {
	return "user_accumulated_profit_records"
}

func (r *Repo) CreateUserAccumulatedProfitRecord(ctx context.Context, record *UserAccumulatedProfitRecord) error {
	return r.DB.WithContext(ctx).Create(record).Error
}


func (r *Repo) UpsertUserAccumulatedProfitRecord(ctx context.Context, record *UserAccumulatedProfitRecord) error {
	return r.DB.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "uid"}, {Name: "begin_time"}, {Name: "end_time"}},
			DoUpdates: clause.AssignmentColumns([]string{"profit", "updated_at"}),
		}).
		Create(record).Error
}


func (r *Repo) GetUserAccumulatedProfitRecordByTimeRange(ctx context.Context, beginTime, endTime time.Time) ([]UserAccumulatedProfitRecord, error) {
	var records []UserAccumulatedProfitRecord
	err := r.DB.WithContext(ctx).
		Where("begin_time = ? AND end_time = ?", beginTime, endTime).
		Find(&records).Error
	return records, err
}
