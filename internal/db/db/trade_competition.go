package db

import (
	"context"
	"errors"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func init() {
	addMigrateFunc(func(repo *Repo) error {
		return repo.AutoMigrate(&TradeCompetitionRecord{})
	})
}

type TradeCompetitionRecord struct {
	gorm.Model
	UID       string    `gorm:"column:uid;type:varchar(255);not null;index:idx_uid;uniqueIndex:idx_uid_begin_end_time"`
	Points    int       `gorm:"column:points;type:int;not null;default:0"`
	BeginTime time.Time `gorm:"column:begin_time;type:timestamp;not null;index:idx_begin_end_time;uniqueIndex:idx_uid_begin_end_time"`
	EndTime   time.Time `gorm:"column:end_time;type:timestamp;not null;index:idx_begin_end_time;uniqueIndex:idx_uid_begin_end_time"`
}

func (TradeCompetitionRecord) TableName() string {
	return "trade_competition_records"
}

func (r *Repo) GetTradeCompetitionRecord(ctx context.Context, uid string) (*TradeCompetitionRecord, error) {
	record := &TradeCompetitionRecord{}
	err := r.DB.WithContext(ctx).Where("uid = ?", uid).First(record).Error
	return record, err
}

func (r *Repo) UpsertTradeCompetitionRecord(ctx context.Context, record *TradeCompetitionRecord) error {
	return r.DB.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "uid"}, {Name: "begin_time"}, {Name: "end_time"}},
			DoUpdates: clause.AssignmentColumns([]string{"points", "updated_at"}),
		}).
		Create(record).Error
}

func (r *Repo) GetUserTotalPoints(ctx context.Context, uid string, beginTime, endTime time.Time) (int, error) {
	var totalPoints int
	err := r.DB.WithContext(ctx).
		Model(&TradeCompetitionRecord{}).
		Where("uid = ? AND begin_time >= ? AND end_time <= ?", uid, beginTime, endTime).
		Select("COALESCE(SUM(points), 0)").
		Scan(&totalPoints).Error
	return totalPoints, err
}

func (r *Repo) GetIssuedPoints(ctx context.Context, beginTime, endTime time.Time) (int, error) {
	var totalPoints int
	err := r.DB.WithContext(ctx).
		Model(&TradeCompetitionRecord{}).
		Where("begin_time >= ? AND end_time <= ?", beginTime, endTime).
		Select("COALESCE(SUM(points), 0)").
		Scan(&totalPoints).Error
	return totalPoints, err
}

func (r *Repo) GetUserDayProfitRanking(ctx context.Context, dayTime time.Time, limit int, blacklist []string) ([]UserDayProfitRecord, error) {
	var records []UserDayProfitRecord
	query := r.DB.WithContext(ctx).
		Where("time = ? and profit > 0", dayTime)

	if len(blacklist) > 0 {
		query = query.Where("uid NOT IN (?)", blacklist)
	}

	err := query.Order("profit DESC").
		Limit(limit).
		Find(&records).Error
	return records, err
}

func (r *Repo) GetUserDayProfitRankAndProfit(ctx context.Context, dayTime time.Time, uid string, blacklist []string) (*UserDayProfitRecord, int, error) {
	var record UserDayProfitRecord
	err := r.DB.WithContext(ctx).
		Where("uid = ? and time = ?", uid, dayTime).
		First(&record).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, 0, nil
		}
		return nil, 0, err
	}
	if record.Profit.Equal(decimal.Zero) {
		return &record, 0, nil
	}

	var rank int64
	err = r.DB.WithContext(ctx).
		Model(&UserDayProfitRecord{}).
		Where("time = ? AND profit > ? AND uid NOT IN (?)", dayTime, record.Profit, blacklist).
		Count(&rank).Error
	if err != nil {
		return nil, 0, err
	}

	return &record, int(rank + 1), nil
}

func (r *Repo) GetUserAccumulatedProfitRanking(ctx context.Context, beginTime, endTime time.Time, limit int, blacklist []string) ([]UserAccumulatedProfitRecord, error) {
	var records []UserAccumulatedProfitRecord
	query := r.DB.WithContext(ctx).
		Where("begin_time = ? AND end_time = ? AND profit > 0", beginTime, endTime)

	if len(blacklist) > 0 {
		query = query.Where("uid NOT IN (?)", blacklist)
	}

	err := query.Order("profit DESC").
		Limit(limit).
		Find(&records).Error
	return records, err
}

func (r *Repo) GetUserAccumulatedProfitRankAndProfit(ctx context.Context, beginTime, endTime time.Time, uid string, blacklist []string) (*UserAccumulatedProfitRecord, int, error) {
	var record UserAccumulatedProfitRecord
	err := r.DB.WithContext(ctx).
		Where("begin_time = ? AND end_time = ? AND uid = ?", beginTime, endTime, uid).
		First(&record).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, 0, nil
		}
		return nil, 0, err
	}
	if record.Profit.Equal(decimal.Zero) {
		return &record, 0, nil
	}

	var rank int64
	err = r.DB.WithContext(ctx).
		Model(&UserAccumulatedProfitRecord{}).
		Where("begin_time = ? AND end_time = ? AND profit > ? AND uid NOT IN (?)", beginTime, endTime, record.Profit, blacklist).
		Count(&rank).Error
	if err != nil {
		return nil, 0, err
	}

	return &record, int(rank + 1), nil
}
