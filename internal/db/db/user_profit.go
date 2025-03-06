package db

import (
	"context"
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

type UserBalanceRecord struct {
	gorm.Model
	UID        string          `gorm:"column:uid;type:varchar(255);not null;uniqueIndex:idx_uid_time"`
	Time       time.Time       `gorm:"column:time;type:timestamp;not null;uniqueIndex:idx_uid_time;index:idx_time"`
	Account    string          `gorm:"column:account;type:varchar(255);not null;"`
	USDTAmount decimal.Decimal `gorm:"column:usdt_amount;type:decimal(20,6);not null;"`
}

func (t *UserBalanceRecord) TableName() string {
	return "user_balance_records"
}

func (r *Repo) UpsertUserBalanceRecord(ctx context.Context, record *UserBalanceRecord) error {
	return r.DB.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "uid"}, {Name: "time"}},
			DoUpdates: clause.AssignmentColumns([]string{"usdt_amount", "updated_at"}),
		}).
		Create(record).Error
}

func (r *Repo) GetUserBalanceRecordsInTimeRange(ctx context.Context, uid string, beginTime, endTime time.Time) ([]UserBalanceRecord, error) {
	var records []UserBalanceRecord
	err := r.DB.WithContext(ctx).
		Where("uid = ? AND time >= ? AND time <= ?", uid, beginTime, endTime).
		Order("time ASC").
		Find(&records).Error
	return records, err
}

func (r *Repo) GetUserBalanceRecordsInTimeRangeForUIDs(ctx context.Context, uids []string, beginTime, endTime time.Time) ([]UserBalanceRecord, error) {
	var records []UserBalanceRecord
	err := r.DB.WithContext(ctx).
		Where("uid IN ? AND time >= ? AND time <= ?", uids, beginTime, endTime).
		Order("time ASC").
		Find(&records).Error
	return records, err
}

func (r *Repo) GetUserBalanceRecordsByTimeRange(ctx context.Context, startTime, endTime time.Time) ([]UserBalanceRecord, error) {
	var records []UserBalanceRecord
	err := r.DB.WithContext(ctx).
		Where("time >= ? AND time <= ?", startTime, endTime).
		Order("time ASC").
		Find(&records).Error
	return records, err
}

func (r *Repo) BatchCreateUserBalanceRecords(ctx context.Context, records []*UserBalanceRecord) error {
	if len(records) == 0 {
		return nil
	}
	return r.DB.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "uid"}, {Name: "time"}},
			DoUpdates: clause.AssignmentColumns([]string{"usdt_amount", "updated_at"}),
		}).
		CreateInBatches(records, 100).Error
}

type UserDayProfitRecord struct {
	gorm.Model
	UID     string          `gorm:"column:uid;type:varchar(255);not null;uniqueIndex:idx_uid_time"`
	Time    time.Time       `gorm:"column:time;type:timestamp;not null;uniqueIndex:idx_uid_time;index:idx_time"`
	Account string          `gorm:"column:account;type:varchar(255);not null;"`
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

func (r *Repo) BatchUpsertUserDayProfitRecords(ctx context.Context, records []*UserDayProfitRecord) error {
	if len(records) == 0 {
		return nil
	}
	return r.DB.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "uid"}, {Name: "time"}},
			DoUpdates: clause.AssignmentColumns([]string{"profit", "updated_at"}),
		}).
		CreateInBatches(records, 100).Error
}

type UserAccumulatedProfitRecord struct {
	gorm.Model
	UID       string          `gorm:"column:uid;type:varchar(255);not null;uniqueIndex:idx_uid_begin_end_time"`
	BeginTime time.Time       `gorm:"column:begin_time;type:timestamp;not null;uniqueIndex:idx_uid_begin_end_time;index:idx_begin_end_time"`
	EndTime   time.Time       `gorm:"column:end_time;type:timestamp;not null;uniqueIndex:idx_uid_begin_end_time;index:idx_begin_end_time"`
	Account   string          `gorm:"column:account;type:varchar(255);not null;"`
	Profit    decimal.Decimal `gorm:"column:profit;type:decimal(20,6);not null;"`
}

func (t *UserAccumulatedProfitRecord) TableName() string {
	return "user_accumulated_profit_records"
}

func (r *Repo) CreateUserAccumulatedProfitRecord(ctx context.Context, record *UserAccumulatedProfitRecord) error {
	return r.DB.WithContext(ctx).Create(record).Error
}

func (r *Repo) BatchUpsertUserAccumulatedProfitRecords(ctx context.Context, records []*UserAccumulatedProfitRecord) error {
	if len(records) == 0 {
		return nil
	}
	return r.DB.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "uid"}, {Name: "begin_time"}, {Name: "end_time"}},
			DoUpdates: clause.AssignmentColumns([]string{"profit", "updated_at"}),
		}).
		CreateInBatches(records, 100).Error
}

func (r *Repo) GetUserAccumulatedProfitRecordByTimeRange(ctx context.Context, beginTime, endTime time.Time) ([]UserAccumulatedProfitRecord, error) {
	var records []UserAccumulatedProfitRecord
	err := r.DB.WithContext(ctx).
		Where("begin_time = ? AND end_time = ?", beginTime, endTime).
		Find(&records).Error
	return records, err
}
