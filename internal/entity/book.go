package entity

type OrderEntry struct {
	Price    float64 `json:"price"`
	Quantity float64 `json:"quantity"`
}

type OrderBook struct {
	PairID string `json:"pair_id"`
	Timestamp Time `json:"timestamp"`
	Bids []OrderEntry `json:"bids"`
	Asks []OrderEntry `json:"asks"`
}
