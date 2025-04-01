package db

import (
	"context"
	"exapp-go/data"
	"exapp-go/internal/types"

	"gorm.io/gorm"
)

func init() {

	addMigrateFunc(func(r *Repo) error {

		return r.AutoMigrate(&UserPoints{})
	})
}

type UserPoints struct {
	gorm.Model
	UID         string `json:"uid" gorm:"column:uid;type:varchar(255);not null;uniqueIndex:idx_uid"`
	Trade       uint64 `json:"trade" gorm:"column:trade;type:bigint(20);default:0"`               // 交易获得积分
	TradeRebate uint64 `json:"trade_rebate" gorm:"column:trade_rebate;type:bigint(20);default:0"` // 交易返佣获得积分
	Total       uint64 `json:"total" gorm:"column:total;type:bigint(20);default:0"`               // 总积分
	Balance     uint64 `json:"balance" gorm:"column:balance;type:bigint(20);default:0"`           // 可用积分
	Invitation  uint64 `json:"invitation" gorm:"column:invitation;type:bigint(20);default:0"`     // 邀请获得积分
	Community   uint64 `json:"community" gorm:"column:community;type:bigint(20);default:0"`       // 社区获得积分
}

func UserPointsRedisKey(uid string) string {
	return "points:detail:" + uid
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

	if pointsType == types.UserPointsTypeCommunity {
		params["community"] = gorm.Expr("community + ?", points)
	}

	r.DelCache(UserPointsRedisKey(uid))
	return r.WithContext(ctx).DB.Model(&UserPoints{}).Where("uid = ?", uid).Updates(params).Error
}

func (r *Repo) DecreaseUserPoints(ctx context.Context, uid string, points uint64) error {
	return r.WithContext(ctx).DB.Model(&UserPoints{}).Where("uid = ?", uid).Update("balance", gorm.Expr("balance - ?", points)).Error
}

func (r *Repo) AddTradeUserPoints(ctx context.Context, uid, txId string, points, globalSeq uint64, pointsType types.UserPointsType) error {

	// 先查询当前积分信息
	userPoints, e := r.GetUserPoints(ctx, uid)
	if e != nil {

		return e
	}

	balance := userPoints.Balance
	// 先创建积分记录
	record := &UserPointsRecord{
		UID:         uid,
		TxId:        txId,
		GlobalSeq:   globalSeq,
		Type:        types.UserPointsTypeTrade,
		Method:      types.UserPointsMethodIn,
		Points:      points,
		Balance:     balance + points,
		SnapBalance: balance,
		Remark:      "",
	}

	// 插入积分记录
	if e = r.Insert(ctx, record); e != nil {

		return e
	}

	// 更新积分
	if e = r.IncreaseUserPoints(ctx, uid, points, types.UserPointsTypeTrade); e != nil {

		return e
	}

	return nil
}

type UserPointsListRes ListResult[UserPoints]

func (r *Repo) ListUserPoints(ctx context.Context, param data.UserPointsListParam) (result UserPointsListRes, err error) {

	if len(param.Order) == 0 {
		param.Order = "created_at desc"
	}

	res, err := List[data.UserPointsListParam, UserPoints](param)
	if err != nil {
		return
	}

	result = UserPointsListRes(res)
	return
}

// 获取用户积分排名
func (r *Repo) GetUserPointsRank(ctx context.Context, uid string) (uint64, error) {

	sql := "SELECT COUNT(*) FROM user_points WHERE total > (SELECT total FROM user_points WHERE uid = ?)"

	var rank uint64
	if err := r.WithContext(ctx).DB.Raw(sql, uid).Scan(&rank).Error; err != nil {
		return 0, err
	}

	return rank, nil
}
