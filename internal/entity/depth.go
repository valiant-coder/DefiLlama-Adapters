package entity

type Depth struct {
	PoolID    uint64     `json:"pool_id"`
	Timestamp uint64     `json:"timestamp"`
	Bids      [][]string `json:"bids"`
	Asks      [][]string `json:"asks"`
	Precision string     `json:"precision"`
}
