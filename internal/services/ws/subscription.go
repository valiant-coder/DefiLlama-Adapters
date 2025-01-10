package ws


type SubscriptionType string

const (
	SubTypeKline  SubscriptionType = "kline"
	SubTypeDepth  SubscriptionType = "depth"
	SubTypeTrades SubscriptionType = "trades"
)


type Subscription struct {
	PoolID   uint64
	Type     SubscriptionType
	Interval string 
}

