package db

import (
	"context"
	"exapp-go/data"
	"exapp-go/pkg/tools"
	"exapp-go/pkg/types"
	"time"
	
	"gorm.io/gorm"
)

func init() {
	
	addMigrateFunc(func(r *Repo) error {
		
		return r.AutoMigrate(&UserPointsRecord{})
	})
}

type UserPointsRecord struct {
	gorm.Model
	
	UID         string                 `gorm:"column:uid;type:varchar(255);not null;uniqueIndex:idx_uid_tx_id;"`
	TxId        string                 `gorm:"column:tx_id;type:varchar(255);not null;uniqueIndex:idx_uid_tx_id;"`
	Type        types.UserPointsType   `gorm:"column:type;type:varchar(255);not null;index:idx_type_method"`
	Method      types.UserPointsMethod `gorm:"column:method;type:varchar(255);not null;index:idx_type_method"`
	Points      uint                   `gorm:"column:points;type:int;default:0"`
	Balance     uint64                 `gorm:"column:balance;type:bigint(20);default:0"`
	SnapBalance uint64                 `gorm:"column:snap_balance;type:bigint(20);default:0"`
	Remark      string                 `gorm:"column:remark;type:varchar(255);"`
}

type UPRecordListRes ListResult[UserPointsRecord]

func (r *Repo) ListUPRecords(ctx context.Context, param data.UPRecordListParam) (result UPRecordListRes, err error) {
	if len(param.Order) == 0 {
		
		param.Order = "created_at desc"
	}
	
	res, err := List[data.UPRecordListParam, UserPointsRecord](param, func(db *gorm.DB, param data.ListParamInterface) any {
		rawParam := param.(data.UPRecordListParam)
		
		if rawParam.Timestamp > 0 {
			
			startTime := tools.GetBeijingTime(rawParam.Timestamp)
			endTime := startTime.Add(time.Duration(rawParam.Interval) * time.Second)
			
			db = db.Where("created_at BETWEEN ? AND ?", startTime, endTime)
		}
		
		return db
	}, func(db *gorm.DB, param data.ListParamInterface) any {
		
		var summary OrderSummary
		
		sql := "count(*) as total_count, sum(points) as total_value"
		if err = db.Select(sql).Scan(&summary).Error; err != nil {
			
			return nil
		}
		
		return &summary
	})
	
	if err != nil {
		
		return
	}
	
	result = UPRecordListRes(res)
	return
	
}
