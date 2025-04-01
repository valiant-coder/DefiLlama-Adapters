package types

type UserPointsType string

const (
	UserPointsTypeTrade      UserPointsType = "trade"      // 交易
	UserPointsTypeActive     UserPointsType = "active"     // 活动
	UserPointsTypeManual     UserPointsType = "manual"     // 手动
	UserPointsTypeInvitation UserPointsType = "invitation" // 邀请
	UserPointsTypeCommunity  UserPointsType = "community"  // 社区
)

type UserPointsMethod string

const (
	UserPointsMethodIn  UserPointsMethod = "in"
	UserPointsMethodOut UserPointsMethod = "out"
)

type UserInviteRebateType string

const (
	UserInviteRebateTypeTrade       UserInviteRebateType = "trade"        // 交易
	UserInviteRebateTypeTradeInvite UserInviteRebateType = "trade-invite" // 邀请返佣
)

type NSQTopic string

const (
	TopicActionSync  NSQTopic = "cdex_action_sync"
	TopicCdexUpdates NSQTopic = "cdex_updates"
)

type NSQMessageType string

// NSQ message types
const (
	MsgTypeOrderUpdate     NSQMessageType = "order_update"
	MsgTypeBalanceUpdate   NSQMessageType = "balance_update"
	MsgTypeTradeUpdate     NSQMessageType = "trade_update"
	MsgTypeTradeDetail     NSQMessageType = "trade_detail"
	MsgTypeDepthUpdate     NSQMessageType = "depth_update"
	MsgTypeKlineUpdate     NSQMessageType = "kline_update"
	MsgTypePoolStatsUpdate NSQMessageType = "pool_stats_update"
	MsgTypeUserCredential  NSQMessageType = "new_user_credential"
)
