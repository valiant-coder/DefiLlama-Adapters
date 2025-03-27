package db

import (
	"context"
	
	"gorm.io/gorm"
)

func init() {
	
	addMigrateFunc(func(r *Repo) error {
		
		return r.AutoMigrate(&UserPointsConf{})
	})
}

type UserPointsConf struct {
	gorm.Model
	
	// 交易配置
	BaseTradePoints   uint   `json:"base_trade_points" gorm:"column:base_trade_points;type:int(11);not null;default:0"`          // 基础交易积分
	MakerWeight       uint   `json:"maker_weight" gorm:"column:maker_weight;type:int(11);not null;default:0"`                    // maker权重
	TakerWeight       uint   `json:"taker_weight" gorm:"column:taker_weight;type:int(11);not null;default:0"`                    // taker权重
	FirstTradeRate    uint   `json:"first_trade_rate" gorm:"column:first_trade_rate;type:int(11);not null;default:2"`            // 首次交易倍数
	MaxPerTradePoints uint64 `json:"max_per_trade_points" gorm:"column:max_per_trade_points;type:bigint(20);not null;default:0"` // 每笔最大交易积分
	
	// 邀请配置
	InvitePercent        uint   `json:"invite_percent" gorm:"column:invite_percent;type:int(11);not null;default:2000"`                 // 邀请百分比
	InviteRebatePercent  uint   `json:"invite_rebate_percent" gorm:"column:invite_rebate_percent;type:int(11);not null;default:500"`    // 邀请返佣百分比
	MaxPerInvitePoints   uint64 `json:"max_per_invite_points" gorm:"column:max_per_invite_points;type:bigint(20);not null;default:0"`   // 每笔最大邀请积分
	UpgradeInviterCount  uint   `json:"upgrade_inviter_count" gorm:"column:upgrade_inviter_count;type:int(11);not null;default:100"`    // 升级邀请人数
	UpgradeInvitePercent uint   `json:"upgrade_invite_percent" gorm:"column:upgrade_invite_percent;type:int(11);not null;default:4000"` // 升级邀请奖励百分比
	
	// 其他配置
	OrderMinPendingTime uint   `json:"order_min_pending_time" gorm:"column:order_min_pending_time;type:int(11);not null;default:60"` // 订单最小挂单时间(seconds)
	OrderMinValue       uint64 `json:"order_min_value" gorm:"column:order_min_value;type:int(11);not null;default:0"`                // 订单最小金额
	MaxInviteLinkCount  uint   `json:"max_invite_link_count" gorm:"column:max_invite_link_count;type:int(11);not null;default:20"`   // 最大邀请链接数量
}

func (c *UserPointsConf) RedisKey() string {
	
	return UserPointsConfRedisKey()
}

func UserPointsConfRedisKey() string {
	
	return "points-conf:detail"
}

func (r *Repo) GetUserPointsConf(ctx context.Context) (*UserPointsConf, error) {
	
	return Get[UserPointsConf](&UserPointsConf{})
}

func (r *Repo) SaveUserPointsConf(ctx context.Context, conf *UserPointsConf) error {
	
	err := r.WithContext(ctx).DB.Save(conf).Error
	if err != nil {
		return err
	}
	
	r.SaveCache(conf)
	return nil
}
