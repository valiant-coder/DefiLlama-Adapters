package types

type UserPointsType string

const (
	UserPointsTypeTrade       UserPointsType = "trade"        // 交易
	UserPointsTypeTradeRebate UserPointsType = "trade_rebate" // 交易返佣
	UserPointsTypeActive      UserPointsType = "active"       // 活动
	UserPointsTypeManual      UserPointsType = "manual"       // 手动
	UserPointsTypeInvitation  UserPointsType = "invitation"   // 邀请
)

type UserPointsMethod string

const (
	UserPointsMethodIn  UserPointsMethod = "in"
	UserPointsMethodOut UserPointsMethod = "out"
)
