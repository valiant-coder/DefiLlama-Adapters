package entity

type Depth struct {
	PoolID    uint64     `json:"pool_id"`
	Timestamp Time       `json:"timestamp"`
	Bids      [][]string `json:"bids"`
	Asks      [][]string `json:"asks"`
}

