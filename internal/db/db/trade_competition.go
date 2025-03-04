package db

import (
	"context"
	"time"

	"gorm.io/gorm"
)

type TradeCompetitionRecord struct {
	gorm.Model
	UID       string    `gorm:"column:uid;type:varchar(255);not null;index:idx_uid"`
	Points    int       `gorm:"column:points;type:int;not null;default:0"`
	BeginTime time.Time `gorm:"column:begin_time;type:timestamp;not null;default:null"`
	EndTime   time.Time `gorm:"column:end_time;type:timestamp;not null;default:null"`
}

func (TradeCompetitionRecord) TableName() string {
	return "trade_competition_records"
}

func (r *Repo) GetTradeCompetitionRecord(ctx context.Context, uid string) (*TradeCompetitionRecord, error) {
	record := &TradeCompetitionRecord{}
	err := r.DB.WithContext(ctx).Where("uid = ?", uid).First(record).Error
	return record, err
}

