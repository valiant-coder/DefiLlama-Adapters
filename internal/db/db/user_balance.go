package db

import (
	"context"

	"github.com/shopspring/decimal"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type UserPoolBalance struct {
	PoolID     uint64          `json:"pool_id"`
	PoolSymbol string          `json:"pool_symbol"`
	Balance    decimal.Decimal `json:"balance"`
}

type UserBalance struct {
	gorm.Model
	Account      string                               `gorm:"column:account;type:varchar(255);not null;"`
	Contract     string                               `gorm:"column:contract;type:varchar(255);not null;"`
	Symbol       string                               `gorm:"column:symbol;type:varchar(255);not null;"`
	Balance      decimal.Decimal                      `gorm:"column:balance;type:decimal(36,18);not null;"`
	Locked       decimal.Decimal                      `gorm:"column:locked;type:decimal(36,18);not null;"`
	PoolBalances datatypes.JSONSlice[UserPoolBalance] `gorm:"column:pool_balances;type:jsonb;not null;"`
}

func (UserBalance) TableName() string {
	return "user_balances"
}



func (r *Repo) GetUserBalances(ctx context.Context, accountName string) ([]UserBalance, error) {
	var userBalances []UserBalance
	if err := r.WithContext(ctx).Where("account = ?", accountName).Find(&userBalances).Error; err != nil {
		return nil, err
	}
	return userBalances, nil
}
