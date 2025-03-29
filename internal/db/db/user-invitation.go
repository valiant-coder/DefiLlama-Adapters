package db

import (
	"context"
	"exapp-go/data"
	"time"
	
	"gorm.io/gorm"
)

func init() {
	
	addMigrateFunc(func(r *Repo) error {
		
		return r.AutoMigrate(&UserInvitation{})
	})
}

type UserInvitation struct {
	gorm.Model
	UID           string `json:"uid" gorm:"column:uid;type:varchar(255);not null;uniqueIndex:idx_uid"`
	Inviter       string `json:"inviter" gorm:"column:inviter;type:varchar(255);index:idx_inviter"`             // 邀请人
	InviteCode    string `json:"invite_code" gorm:"column:invite_code;type:varchar(255);index:idx_invite_code"` // 被邀请码
	InvitePercent uint   `json:"invite_percent" gorm:"column:invite_percent;type:int(11);not null;default:0"`   // 被邀请返佣
	
	InviteCount      uint `json:"invite_count" gorm:"column:invite_count;type:int(11);not null;default:0"`             // 邀请数量
	ValidInviteCount uint `json:"valid_invite_count" gorm:"column:valid_invite_count;type:int(11);not null;default:0"` // 有效邀请数量（有交易用户
	
	LinkCount  uint `json:"link_count" gorm:"column:link_count;type:int(11);not null;default:0"`      // 邀请链接数量
	MaxPercent uint `json:"max_percent" gorm:"column:max_percent;type:int(11);not null;default:2000"` // 20% 最大比例，百分位
	
	FirstTradeTxId string `json:"first_trade_tx_id" gorm:"column:first_trade_tx_id;type:varchar(255);default:null;index:idx_first_trade_tx_id"`
	FirstTradeAt   int64  `json:"first_trade_at"`
}

func UIRedisKey(uid string) string {
	
	return "user-invitation:uid:" + uid
}

func (ui *UserInvitation) RedisKey() string {
	
	return UIRedisKey(ui.UID)
}

type UserInvitationListRes ListResult[UserInvitation]

func (r *Repo) ListUserInvitation(param data.UserInvitationListParam) (result UserInvitationListRes, err error) {
	if len(param.Order) == 0 {
		param.Order = "created_at desc"
	}
	
	res, err := List[data.UserInvitationListParam, UserInvitation](param)
	if err != nil {
		return
	}
	
	result = UserInvitationListRes(res)
	return
}

func (r *Repo) UpdateUILinkCount(ctx context.Context, uid string, isDelete bool) (err error) {
	expr := "link_count + 1"
	if isDelete {
		expr = "link_count - 1"
	}
	
	r.DelCache(UIRedisKey(uid))
	
	return r.WithContext(ctx).DB.Model(&UserInvitation{}).Where("uid = ?", uid).Update("link_count", gorm.Expr(expr)).Error
}

func (r *Repo) UpdateUIInviteCount(ctx context.Context, uid string) (err error) {
	
	r.DelCache(UIRedisKey(uid))
	return r.WithContext(ctx).DB.Model(&UserInvitation{}).Where("uid = ?", uid).Update("invite_count", gorm.Expr("invite_count + 1")).Error
}

func (r *Repo) UpdateUIValidInviteCount(ctx context.Context, uid string, maxCount, maxPercent uint) (err error) {
	
	params := map[string]interface{}{
		"valid_invite_count": gorm.Expr("valid_invite_count + 1"),
		"max_percent":        gorm.Expr("IF(valid_invite_count >= ?, ?, max_percent)", maxCount, maxPercent),
	}
	
	r.DelCache(UIRedisKey(uid))
	return r.WithContext(ctx).DB.Model(&UserInvitation{}).Where("uid = ?", uid).Updates(params).Error
}

func (r *Repo) GetUserInvitation(ctx context.Context, uid string) (*UserInvitation, error) {
	
	return Get[UserInvitation](&UserInvitation{UID: uid})
}

func (r *Repo) SetUserFirstTrade(ctx context.Context, uid, txId string) (err error) {
	
	sql := "UPDATE user_invitation SET first_trade_tx_id = ?, first_trade_at = ? WHERE uid = ? and first_trade_tx_id is null"
	
	if err = r.WithContext(ctx).DB.Exec(sql, txId, time.Now().Unix(), uid).Error; err != nil {
		
		return
	}
	
	r.DelCache(UIRedisKey(uid))
	return
}
