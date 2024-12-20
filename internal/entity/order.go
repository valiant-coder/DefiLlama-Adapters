package entity

type Order struct {
	ID                      uint64 `json:"id"`
	PoolID                  uint64 `json:"pool_id"`
	ClientOrderID           string `json:"order_cid"`
	Trader                  string `json:"trader"`
	// 0: no restriction, 1: immediate or cancel, 2: fill or kill, 3: post only
	Type                    uint8 `json:"type"`
	Price                   uint64 `json:"price"`
	IsBid                   bool   `json:"is_bid"`
	OriginalQuantity        string `json:"original_quantity"`
	ExecutedQuantity        string `json:"executed_quantity"`
	CumulativeQuoteQuantity string `json:"cumulative_quote_quantity"`
	PaidFees                string `json:"paid_fees"`
	Status                  string `json:"status"`
	IsMarket                bool   `json:"is_market"`
	CreatedAt               Time   `json:"created_at"`

}


type OrderDetail struct {
	Order
	Trades []TradeDetail `json:"trades"`
}

type OrderBook struct {
	PoolID    uint64     `json:"pool_id"`
	Timestamp Time       `json:"timestamp"`
	Bids      [][]string `json:"bids"`
	Asks      [][]string `json:"asks"`
}
