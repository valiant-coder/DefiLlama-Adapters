package db

import (
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type UserPoolBalance struct {
	gorm.Model
	AccountName string          `gorm:"column:account_name;type:varchar(255);not null;"`
	PoolID      uint64          `gorm:"column:pool_id;type:bigint;not null;"`
	Contract    string          `gorm:"column:contract;type:varchar(255);not null;"`
	Symbol      string          `gorm:"column:symbol;type:varchar(255);not null;"`
	Balance     decimal.Decimal `gorm:"column:balance;type:decimal(36,18);not null;"`
}

func (UserPoolBalance) TableName() string {
	return "user_pool_balances"
}

type UserBalance struct {
	gorm.Model
	AccountName string          `gorm:"column:account_name;type:varchar(255);not null;"`
	Contract    string          `gorm:"column:contract;type:varchar(255);not null;"`
	Symbol      string          `gorm:"column:symbol;type:varchar(255);not null;"`
	Balance     decimal.Decimal `gorm:"column:balance;type:decimal(36,18);not null;"`
}


func (UserBalance) TableName() string {
	return "user_balances"
}
