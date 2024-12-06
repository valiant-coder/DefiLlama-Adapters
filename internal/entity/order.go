package entity

type OrderSide string

const (
	OrderSideBuy  OrderSide = "buy"
	OrderSideSell OrderSide = "sell"
)

type OrderStatus string

const (
	OrderStatusOpen            OrderStatus = "open"
	OrderStatusFilled          OrderStatus = "filled"
	OrderStatusPartiallyFilled OrderStatus = "partially_filled"
	OrderStatusCanceled        OrderStatus = "canceled"
)

type OrderType string

const (
	OrderTypeLimit  OrderType = "limit"
	OrderTypeMarket OrderType = "market"
)

type Order struct {
	ID string `json:"id"`
	// User
	User string `json:"user"`
	// Trading pair ID
	PairID string `json:"pair_id"`
	// Order side
	Side OrderSide `json:"side"`
	// Order price
	Price string `json:"price"`
	// Order quantity
	Quantity string `json:"quantity"`
	// Order amount
	Amount string `json:"amount"`
	// Order timestamp
	Timestamp Time `json:"timestamp"`
	// Order status
	Status OrderStatus `json:"status"`
	// Order type
	Type OrderType `json:"type"`
	// Maker
	Maker string `json:"maker"`
	// Fee
	Fee string `json:"fee"`
	// Fee asset
	FeeAsset string `json:"fee_asset"`
	// Executed quantity
	ExecutedQuantity string `json:"executed_quantity"`
	// Executed amount
	ExecutedAmount string `json:"executed_amount"`
	// Last trade time
	LastTradeTime Time `json:"last_trade_time"`
}
