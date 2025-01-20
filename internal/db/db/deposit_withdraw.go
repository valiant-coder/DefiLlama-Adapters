package db

import (
	"context"
	"exapp-go/pkg/queryparams"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func init() {
	addMigrateFunc(func(repo *Repo) error {
		return repo.AutoMigrate(
			&DepositRecord{},
			&UserDepositAddress{},
			&UserWithdrawRecord{},
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
	Symbol         string          `gorm:"column:symbol;type:varchar(255);not null"`
	UID            string          `gorm:"column:uid;type:varchar(255);not null;index:idx_uid"`
	ChainName      string          `gorm:"column:chain_name;type:varchar(255);not null"`
	SourceTxID     string          `gorm:"column:source_tx_id;type:varchar(255);default:null;uniqueIndex:idx_source_tx_id"`
	DepositAddress string          `gorm:"column:deposit_address;type:varchar(255);not null"`
	Amount         decimal.Decimal `gorm:"column:amount;type:decimal(36,18);not null"`
	Fee            decimal.Decimal `gorm:"column:fee;type:decimal(36,18);not null"`
	Status         DepositStatus   `gorm:"column:status;type:tinyint(3);not null"`
	TxHash         string          `gorm:"column:tx_hash;type:varchar(255);not null"`
	Time           time.Time       `gorm:"column:time;type:datetime;not null"`
}

func (d *DepositRecord) TableName() string {
	return "deposit_records"
}

func (r *Repo) UpsertDepositRecord(ctx context.Context, record *DepositRecord) error {
	return r.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{"status", "tx_hash"}),
	}).Create(record).Error
}

func (r *Repo) GetDepositRecords(ctx context.Context, uid string, queryParams *queryparams.QueryParams) ([]*DepositRecord, int64, error) {
	queryParams.Add("uid", uid)
	var records []*DepositRecord
	total, err := r.Query(ctx, &records, queryParams, "uid")
	return records, total, err
}

func (r *Repo) GetDepositRecordBySourceTxID(ctx context.Context, sourceTxID string) (*DepositRecord, error) {
	var record DepositRecord
	err := r.WithContext(ctx).Where("source_tx_id = ?", sourceTxID).First(&record).Error
	return &record, err
}

type UserDepositAddress struct {
	gorm.Model
	UID          string `gorm:"column:uid;type:varchar(255);not null;index:idx_uid"`
	Address      string `gorm:"column:address;type:varchar(255);not null;index:idx_address"`
	PermissionID uint64 `gorm:"column:permission_id;type:bigint(20);not null;index:idx_permission_id"`
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

func (r *Repo) GetUserDepositAddressByAddress(ctx context.Context, address string) (*UserDepositAddress, error) {
	var userDepositAddress UserDepositAddress
	err := r.DB.WithContext(ctx).Where("address = ?", address).First(&userDepositAddress).Error
	return &userDepositAddress, err
}

type WithdrawStatus uint8

const (
	WithdrawStatusPending WithdrawStatus = iota
	WithdrawStatusSuccess
	WithdrawStatusFailed
)

type UserWithdrawRecord struct {
	gorm.Model
	UID       string          `gorm:"column:uid;type:varchar(255);not null"`
	Symbol    string          `gorm:"column:symbol;type:varchar(255);not null"`
	ChainName string          `gorm:"column:chain_name;type:varchar(255);not null"`
	Amount    decimal.Decimal `gorm:"column:amount;type:decimal(36,18);not null"`
	Fee       decimal.Decimal `gorm:"column:fee;type:decimal(36,18);not null"`
	Status    WithdrawStatus  `gorm:"column:status;type:tinyint(3);not null"`
	TxHash    string          `gorm:"column:tx_hash;type:varchar(255);not null"`
	Time      time.Time       `gorm:"column:time;type:datetime;not null"`
}

func (u *UserWithdrawRecord) TableName() string {
	return "user_withdraw_records"
}

func (r *Repo) CreateWithdrawRecord(ctx context.Context, record *UserWithdrawRecord) error {
	return r.DB.WithContext(ctx).Create(record).Error
}

func (r *Repo) GetWithdrawRecords(ctx context.Context, uid string, queryParams *queryparams.QueryParams) ([]*UserWithdrawRecord, int64, error) {
	queryParams.Add("uid", uid)
	var records []*UserWithdrawRecord
	total, err := r.Query(ctx, &records, queryParams, "uid")
	return records, total, err
}
