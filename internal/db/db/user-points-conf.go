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
	BaseTradePoints   uint   `gorm:"column:base_trade_points;type:int(11);not null;default:0"`       // 基础交易积分
	MakerWeight       uint   `gorm:"column:maker_weight;type:int(11);not null;default:0"`            // maker权重
	TakerWeight       uint   `gorm:"column:taker_weight;type:int(11);not null;default:0"`            // taker权重
	FirstTradeRate    uint   `gorm:"column:first_trade_rate;type:int(11);not null;default:2"`        // 首次交易倍数
	MaxPerTradePoints uint64 `gorm:"column:max_per_trade_points;type:bigint(20);not null;default:0"` // 每笔最大交易积分
	
	// 邀请配置
	InvitePercent        uint   `gorm:"column:invite_percent;type:int(11);not null;default:2000"`         // 邀请百分比
	InviteRebatePercent  uint   `gorm:"column:invite_rebate_percent;type:int(11);not null;default:500"`   // 邀请返佣百分比
	MaxPerInvitePoints   uint64 `gorm:"column:max_per_invite_points;type:bigint(20);not null;default:0"`  // 每笔最大邀请积分
	UpgradeInviterCount  uint   `gorm:"column:upgrade_inviter_count;type:int(11);not null;default:100"`   // 升级邀请人数
	UpgradeInvitePercent uint   `gorm:"column:upgrade_invite_percent;type:int(11);not null;default:4000"` // 升级邀请奖励百分比
	
	// 其他配置
	OrderMinPendingTime uint   `gorm:"column:order_min_pending_time;type:int(11);not null;default:60"` // 订单最小挂单时间(seconds)
	OrderMinValue       uint64 `gorm:"column:order_min_value;type:int(11);not null;default:0"`         // 订单最小金额
	MaxInviteLinkCount  uint   `gorm:"column:max_invite_link_count;type:int(11);not null;default:20"`  // 最大邀请链接数量
}

func (c *UserPointsConf) RedisKey() string {
	
	return UserPointsConfRedisKey()
}

func UserPointsConfRedisKey() string {
	
	return "user-points-conf:detail"
}

func (r *Repo) GetUserPointsConf(ctx context.Context) (*UserPointsConf, error) {
	
	if res := GetCache[UserPointsConf](UserPointsConfRedisKey()); res != nil {
		return res, nil
	}
	
	var conf UserPointsConf
	err := r.WithContext(ctx).DB.First(&conf).Error
	if err != nil {
		return nil, err
	}
	
	r.SaveCache(&conf)
	return &conf, nil
}

func (r *Repo) SaveUserPointsConf(ctx context.Context, conf *UserPointsConf) error {
	
	err := r.WithContext(ctx).DB.Save(conf).Error
	if err != nil {
		return err
	}
	
	r.SaveCache(conf)
	return nil
}
