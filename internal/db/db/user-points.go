package db

import (
	"context"
	"exapp-go/pkg/types"
	
	"gorm.io/gorm"
)

func init() {
	
	addMigrateFunc(func(r *Repo) error {
		
		return r.AutoMigrate(&UserPoints{})
	})
}

type UserPoints struct {
	gorm.Model
	UID         string `gorm:"column:uid;type:varchar(255);not null;uniqueIndex:idx_uid"`
	Trade       uint64 `gorm:"column:trade;type:bigint(20);default:0"`        // 交易获得积分
	TradeRebate uint64 `gorm:"column:trade_rebate;type:bigint(20);default:0"` // 交易返佣获得积分
	Total       uint64 `gorm:"column:total;type:bigint(20);default:0"`        // 总积分
	Balance     uint64 `gorm:"column:balance;type:bigint(20);default:0"`      // 可用积分
	Invitation  uint64 `gorm:"column:invitation;type:bigint(20);default:0"`   // 邀请获得积分
}

func UserPointsRedisKey(uid string) string {
	return "user-points:detail:" + uid
}

func (p *UserPoints) RedisKey() string {
	return UserPointsRedisKey(p.UID)
}

func (r *Repo) GetUserPoints(ctx context.Context, uid string) (*UserPoints, error) {
	return Get[UserPoints](&UserPoints{UID: uid})
}

func (r *Repo) IncreaseUserPoints(ctx context.Context, uid string, points uint64, pointsType types.UserPointsType) error {
	
	params := map[string]interface{}{
		"balance": gorm.Expr("balance + ?", points),
		"total":   gorm.Expr("total + ?", points),
	}
	
	if pointsType == types.UserPointsTypeInvitation {
		params["invitation"] = gorm.Expr("invitation + ?", points)
	}
	
	if pointsType == types.UserPointsTypeTrade {
		params["trade"] = gorm.Expr("trade + ?", points)
	}
	
	if pointsType == types.UserPointsTypeTradeRebate {
		params["trade_rebate"] = gorm.Expr("trade_rebate + ?", points)
	}
	
	r.DelCache(UserPointsRedisKey(uid))
	return r.WithContext(ctx).DB.Model(&UserPoints{}).Where("uid = ?", uid).Updates(params).Error
}

func (r *Repo) DecreaseUserPoints(ctx context.Context, uid string, points uint64) error {
	return r.WithContext(ctx).DB.Model(&UserPoints{}).Where("uid = ?", uid).Update("balance", gorm.Expr("balance - ?", points)).Error
}
