package db

import (
	"context"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

func init() {
	addMigrateFunc(func(r *Repo) error {
		return r.AutoMigrate(&FaucetRecord{})
	})
}

type FaucetRecord struct {
	gorm.Model
	UID            string          `gorm:"type:varchar(255);not null;"`
	DepositAddress string          `gorm:"type:varchar(255);not null;"`
	Token          string          `gorm:"type:varchar(255);not null;"`
	Amount         decimal.Decimal `gorm:"type:decimal(24,6);not null;"`
	TxHash         string          `gorm:"type:varchar(255);not null;"`
}

func (FaucetRecord) TableName() string {
	return "faucet_records"
}

func (r *Repo) CreateFaucetRecord(ctx context.Context, record *FaucetRecord) error {
	return r.Create(record).Error
}

func (r *Repo) IsUserClaimFaucet(ctx context.Context, uid string) (bool, error) {
	var count int64
	err := r.DB.WithContext(ctx).Model(&FaucetRecord{}).Where("uid = ?", uid).Count(&count).Error
	return count > 0, err
}

func (r *Repo) ClaimFaucetCount(ctx context.Context) (int64, error) {
	var count int64
	err := r.DB.WithContext(ctx).Model(&FaucetRecord{}).Count(&count).Error
	return count, err
}
