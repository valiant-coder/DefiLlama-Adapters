package data

import "exapp-go/internal/types"

type UPRecordListParam struct {
	ListParam

	UID    string                 `json:"uid" url:"uid"`
	TxId   string                 `json:"tx_id" url:"tx_id"`
	Type   types.UserPointsType   `json:"type" url:"type"`
	Method types.UserPointsMethod `json:"method" url:"method"`
	Remark string                 `json:"remark" url:"remark" fuzzy:"true"`

	Timestamp int64 `json:"timestamp" url:"timestamp" ignore:"true"`
	Interval  int64 `json:"interval" url:"interval" ignore:"true"`
}
type UserPointsConfParam struct {
	BaseTradePoints   uint   `json:"base_trade_points" url:"base_trade_points"`
	MakerWeight       uint   `json:"maker_weight" url:"maker_weight"`
	TakerWeight       uint   `json:"taker_weight" url:"taker_weight"`
	FirstTradeRate    uint   `json:"first_trade_rate" url:"first_trade_rate"`
	MaxPerTradePoints uint64 `json:"max_per_trade_points" url:"max_per_trade_points"`

	InvitePercent        uint   `json:"invite_percent" url:"invite_percent"`
	InviteRebatePercent  uint   `json:"invite_rebate_percent" url:"invite_rebate_percent"`
	MaxPerInvitePoints   uint64 `json:"max_per_invite_points" url:"max_per_invite_points"`
	UpgradeInviterCount  uint   `json:"upgrade_inviter_count" url:"upgrade_inviter_count"`
	UpgradeInvitePercent uint   `json:"upgrade_invite_percent" url:"upgrade_invite_percent"`

	OrderMinPendingTime uint   `json:"order_min_pending_time" url:"order_min_pending_time"`
	OrderMinValue       uint64 `json:"order_min_value" url:"order_min_value"`
	MaxInviteLinkCount  uint   `json:"max_invite_link_count" url:"max_invite_link_count"`
}
