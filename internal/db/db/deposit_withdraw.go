package db

import (
	"context"
	"exapp-go/pkg/queryparams"
	"time"

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

type DepositStatus uint8

const (
	DepositStatusPending DepositStatus = iota
	DepositStatusSuccess
	DepositStatusFailed
)

type DepositRecord struct {
	gorm.Model
	Symbol string          `gorm:"column:symbol;type:varchar(255);not null"`
	UID    string          `gorm:"column:uid;type:varchar(255);not null;index:idx_uid"`
	Amount decimal.Decimal `gorm:"column:amount;type:decimal(36,18);not null"`
	Fee    decimal.Decimal `gorm:"column:fee;type:decimal(36,18);not null"`
	Status DepositStatus   `gorm:"column:status;type:tinyint(3);not null"`
	TxHash string          `gorm:"column:tx_hash;type:varchar(255);not null"`
	Time   time.Time       `gorm:"column:time;type:datetime;not null"`
}

func (d *DepositRecord) TableName() string {
	return "deposit_records"
}

func (r *Repo) CreateDepositRecord(ctx context.Context, record *DepositRecord) error {
	return r.DB.WithContext(ctx).Create(record).Error
}

func (r *Repo) GetDepositRecords(ctx context.Context, uid string, queryParams *queryparams.QueryParams) ([]*DepositRecord, int64, error) {
	queryParams.Add("uid", uid)
	var records []*DepositRecord
	total, err := r.Query(ctx, &records, queryParams, "uid")
	return records, total, err
}

type UserDepositAddress struct {
	gorm.Model
	UID          string `gorm:"column:uid;type:varchar(255);not null"`
	Address      string `gorm:"column:address;type:varchar(255);not null"`
	PermissionID uint64 `gorm:"column:permission_id;type:bigint(20);not null"`
	Remark       string `gorm:"column:remark;type:varchar(255);not null"`
}

func (u *UserDepositAddress) TableName() string {
	return "user_deposit_addresses"
}

func (r *Repo) CreateUserDepositAddress(ctx context.Context, address *UserDepositAddress) error {
	return r.DB.WithContext(ctx).Create(address).Error
}

func (r *Repo) GetUserDepositAddress(ctx context.Context, uid string, permissionID uint64) ([]UserDepositAddress, error) {
	var addresses []UserDepositAddress
	err := r.DB.WithContext(ctx).Where("uid = ? and permission_id = ?", uid, permissionID).Find(&addresses).Error
	return addresses, err
}
