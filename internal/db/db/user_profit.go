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

func (r *Repo) GetUserDayProfitRanking(ctx context.Context, dayTime time.Time, limit int) ([]UserDayProfitRecord, error) {
	var records []UserDayProfitRecord
	err := r.DB.WithContext(ctx).
		Where("time = ?", dayTime).
		Order("profit DESC").
		Limit(limit).
		Find(&records).Error
	return records, err
}

func (r *Repo) GetUserDayProfitRankAndProfit(ctx context.Context, dayTime time.Time, uid string) (*UserDayProfitRecord, int, error) {
	var record UserDayProfitRecord
	err := r.DB.WithContext(ctx).
		Where("uid = ? AND time = ?", uid, dayTime).
		First(&record).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, 0, nil
		}
		return nil, 0, err
	}

	var rank int64
	err = r.DB.WithContext(ctx).
		Model(&UserDayProfitRecord{}).
		Where("time = ? AND profit > ?", dayTime, record.Profit).
		Count(&rank).Error
	if err != nil {
		return nil, 0, err
	}

	return &record, int(rank + 1), nil
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

func (r *Repo) GetUserAccumulatedProfitRanking(ctx context.Context, beginTime, endTime time.Time, limit int) ([]UserAccumulatedProfitRecord, error) {
	var records []UserAccumulatedProfitRecord
	err := r.DB.WithContext(ctx).
		Where("begin_time = ? AND end_time = ?", beginTime, endTime).
		Order("profit DESC").
		Limit(limit).
		Find(&records).Error
	return records, err
}

func (r *Repo) GetUserAccumulatedProfitRankAndProfit(ctx context.Context, beginTime, endTime time.Time, uid string) (*UserAccumulatedProfitRecord, int, error) {
	var record UserAccumulatedProfitRecord
	err := r.DB.WithContext(ctx).
		Where("uid = ? AND begin_time = ? AND end_time = ?", uid, beginTime, endTime).
		First(&record).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, 0, nil
		}
		return nil, 0, err
	}

	var rank int64
	err = r.DB.WithContext(ctx).
		Model(&UserAccumulatedProfitRecord{}).
		Where("begin_time = ? AND end_time = ? AND profit > ?", beginTime, endTime, record.Profit).
		Count(&rank).Error
	if err != nil {
		return nil, 0, err
	}

	return &record, int(rank + 1), nil
}
