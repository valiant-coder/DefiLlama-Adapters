package entity

type TradeSide string

const (
	TradeSideBuy  TradeSide = "buy"
	TradeSideSell TradeSide = "sell"
)

type Trade struct {
	ID       uint64    `json:"id"`
	PairID   uint64    `json:"pair_id"`
	Buyer    string    `json:"buyer"`
	Seller   string    `json:"seller"`
	Quantity string    `json:"quantity"`
	Price    string    `json:"price"`
	TradedAt Time      `json:"traded_at"`
	Side     TradeSide `json:"side"`
}
