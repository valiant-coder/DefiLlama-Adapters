package db

import (
	"exapp-go/internal/types"

	"gorm.io/gorm"
)

// 邀请返佣余额
type UserInviteRebate struct {
	gorm.Model
	UID     string `json:"uid" gorm:"column:uid;type:varchar(255);not null;uniqueIndex:idx_uid_symbol"`
	Symbol  string `json:"symbol" gorm:"column:symbol;type:varchar(255);not null;uniqueIndex:idx_uid_symbol"`
	Balance uint64 `json:"balance" gorm:"column:balance;type:bigint(20);default:0"`
	Freeze  uint64 `json:"freeze" gorm:"column:freeze;type:bigint(20);default:0"`
	Total   uint64 `json:"total" gorm:"column:total;type:bigint(20);default:0"`
}

// 邀请返佣记录
type UIRebateRecord struct {
	gorm.Model

	UID       string `json:"uid" gorm:"column:uid;type:varchar(255);not null;uniqueIndex:idx_uid_symbol"`
	Symbol    string `json:"symbol" gorm:"column:symbol;type:varchar(255);not null;uniqueIndex:idx_uid_symbol"`
	TxId      string `json:"tx_id" gorm:"column:tx_id;type:varchar(255);not null;uniqueIndex:idx_uid_symbol"`
	GlobalSeq uint64 `json:"global_seq" gorm:"column:global_seq;type:bigint(20);default:0"`

	Type    types.UserInviteRebateType `json:"type" gorm:"column:type;type:varchar(255);not null;index:idx_type"`
	Method  types.UserPointsMethod     `json:"method" gorm:"column:method;type:varchar(255);not null;index:idx_type"`
	Fees    uint64                     `json:"fees" gorm:"column:fees;type:bigint(20);default:0"`
	Amount  uint64                     `json:"amount" gorm:"column:amount;type:bigint(20);default:0"`
	Rebate  uint64                     `json:"rebate" gorm:"column:rebate;type:bigint(20);default:0"`
	Remark  string                     `json:"remark" gorm:"column:remark;type:varchar(255);"`
	Percent uint64                     `json:"percent" gorm:"column:percent;type:bigint(20);default:0"`
}
