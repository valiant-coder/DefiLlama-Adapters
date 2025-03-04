package db

import (
	"context"
	"time"

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
	BeginTime time.Time `gorm:"column:begin_time;type:timestamp;not null;default:null;index:idx_begin_end_time;uniqueIndex:idx_uid_begin_end_time"`
	EndTime   time.Time `gorm:"column:end_time;type:timestamp;not null;default:null;index:idx_begin_end_time;uniqueIndex:idx_uid_begin_end_time"`
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
