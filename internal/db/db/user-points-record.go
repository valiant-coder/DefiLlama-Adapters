package db

import (
	"context"
	"exapp-go/data"
	"exapp-go/internal/types"
	"exapp-go/pkg/tools"
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
	
	UID       string `json:"uid" gorm:"column:uid;type:varchar(255);not null;uniqueIndex:idx_uid_tx_id;"`
	TxId      string `json:"tx_id" gorm:"column:tx_id;type:varchar(255);not null;uniqueIndex:idx_uid_tx_id;"`
	GlobalSeq uint64 `json:"global_seq" gorm:"column:global_seq;type:bigint(20);default:0;uniqueIndex:idx_uid_tx_id;"`
	
	Type        types.UserPointsType   `json:"type" gorm:"column:type;type:varchar(255);not null;index:idx_type_method"`
	Method      types.UserPointsMethod `json:"method" gorm:"column:method;type:varchar(255);not null;index:idx_type_method"`
	Points      uint64                 `json:"points" gorm:"column:points;type:int;default:0"`
	Balance     uint64                 `json:"balance" gorm:"column:balance;type:bigint(20);default:0"`
	SnapBalance uint64                 `json:"snap_balance" gorm:"column:snap_balance;type:bigint(20);default:0"`
	Remark      string                 `json:"remark" gorm:"column:remark;type:varchar(255);"`
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
