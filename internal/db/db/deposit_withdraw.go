package db

import (
	"context"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

func init() {
	addMigrateFunc(func(repo *Repo) error {
		return repo.AutoMigrate(
			&DepositRecord{},
			&UserDepositAddress{},
		)
	})
}

type DepositRecord struct {
	gorm.Model
	Symbol    string          `gorm:"column:symbol;type:varchar(255);not null"`
	ChainName string          `gorm:"column:chain_name;type:varchar(255);not null"`
	UID       string          `gorm:"column:uid;type:varchar(255);not null"`
	Address   string          `gorm:"column:address;type:varchar(255);not null"`
	Amount    decimal.Decimal `gorm:"column:amount;type:decimal(36,18);not null"`
	Fee       decimal.Decimal `gorm:"column:fee;type:decimal(36,18);not null"`
	Status    uint8           `gorm:"column:status;type:varchar(255);not null"`
	TxHash    string          `gorm:"column:tx_hash;type:varchar(255);not null"`
	Remark    string          `gorm:"column:remark;type:varchar(255);not null"`
}

func (d *DepositRecord) TableName() string {
	return "deposit_records"
}

func (r *Repo) CreateDepositRecord(ctx context.Context, record *DepositRecord) error {
	return r.DB.WithContext(ctx).Create(record).Error
}

type UserDepositAddress struct {
	gorm.Model
	UID       string `gorm:"column:uid;type:varchar(255);not null"`
	Address   string `gorm:"column:address;type:varchar(255);not null"`
	Symbol    string `gorm:"column:symbol;type:varchar(255);not null"`
	ChainName string `gorm:"column:chain_name;type:varchar(255);not null"`
	Remark    string `gorm:"column:remark;type:varchar(255);not null"`
}

func (u *UserDepositAddress) TableName() string {
	return "user_deposit_addresses"
}

func (r *Repo) CreateUserDepositAddress(ctx context.Context, address *UserDepositAddress) error {
	return r.DB.WithContext(ctx).Create(address).Error
}

func (r *Repo) GetUserDepositAddress(ctx context.Context, uid string, symbol string, chainName string) ([]UserDepositAddress, error) {
	var addresses []UserDepositAddress
	err := r.DB.WithContext(ctx).Where("uid = ? and symbol = ? and chain_name = ?", uid, symbol, chainName).Find(&addresses).Error
	return addresses, err
}
