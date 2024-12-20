package entity


type Kline struct {
	PoolID    uint64  `json:"pool_id"`
	Timestamp Time    `json:"timestamp"`
	Open      float64 `json:"open"`
	High      float64 `json:"high"`
	Low       float64 `json:"low"`
	Close     float64 `json:"close"`
	Volume    float64 `json:"volume"`
	Turnover  float64 `json:"turnover"`
	Count     int64   `json:"count"`
}
