package entity

type KlineInterval string

type Kline struct {
	PairID    string  `json:"pair_id"`
	Symbol    string  `json:"symbol"`
	Timestamp Time    `json:"timestamp"`
	Open      float64 `json:"open"`
	High      float64 `json:"high"`
	Low       float64 `json:"low"`
	Close     float64 `json:"close"`
	Quantity  float64 `json:"quantity"`
	Amount    float64 `json:"amount"`
	Count     int64   `json:"count"`
}
