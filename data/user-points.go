package data

import "exapp-go/internal/types"

type UPRecordListParam struct {
	ListParam

	UID    string                 `json:"uid" form:"uid"`
	TxId   string                 `json:"tx_id" form:"tx_id"`
	Type   types.UserPointsType   `json:"type" form:"type"`
	Method types.UserPointsMethod `json:"method" form:"method"`
	Remark string                 `json:"remark" form:"remark" fuzzy:"true"`

	Timestamp int64 `json:"timestamp" form:"timestamp" ignore:"true"`
	Interval  int64 `json:"interval" form:"interval" ignore:"true"`
}
type UserPointsConfParam struct {
	BaseTradePoints   uint   `json:"base_trade_points" form:"base_trade_points"`
	MakerWeight       uint   `json:"maker_weight" form:"maker_weight"`
	TakerWeight       uint   `json:"taker_weight" form:"taker_weight"`
	FirstTradeRate    uint   `json:"first_trade_rate" form:"first_trade_rate"`
	MaxPerTradePoints uint64 `json:"max_per_trade_points" form:"max_per_trade_points"`

	InvitePercent        uint `json:"invite_percent" form:"invite_percent"`
	InviteRebatePercent  uint `json:"invite_rebate_percent" form:"invite_rebate_percent"`
	MaxPerInvitePoints   uint `json:"max_per_invite_points" form:"max_per_invite_points"`
	UpgradeInviterCount  uint `json:"upgrade_inviter_count" form:"upgrade_inviter_count"`
	UpgradeInvitePercent uint `json:"upgrade_invite_percent" form:"upgrade_invite_percent"`

	OrderMinPendingTime uint   `json:"order_min_pending_time" form:"order_min_pending_time"`
	OrderMinValue       uint64 `json:"order_min_value" form:"order_min_value"`
	MaxInviteLinkCount  uint   `json:"max_invite_link_count" form:"max_invite_link_count"`
}

type UserPointsPairListParam struct {
	ListParam

	Pair string `json:"pair" form:"pair"`
}
