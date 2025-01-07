package entity

type Pool struct {
	PoolID             uint64    `json:"pool_id" `
	BaseSymbol         string    `json:"base_symbol"`
	BaseContract       string    `json:"base_contract"`
	BaseCoin           string    `json:"base_coin"`
	BaseCoinPrecision  uint8     `json:"base_coin_precision"`
	QuoteSymbol        string    `json:"quote_symbol"`
	QuoteContract      string    `json:"quote_contract"`
	QuoteCoin          string    `json:"quote_coin"`
	Symbol             string    `json:"symbol"`
	QuoteCoinPrecision uint8     `json:"quote_coin_precision"`
	AskingTime         Time      `json:"asking_time"`
	TradingTime        Time      `json:"trading_time"`
	MaxFluctuation     uint64    `json:"max_flct"`
	PricePrecision     uint8     `json:"price_precision"`
	TakerFeeRate       float64   `json:"taker_fee_rate"`
	MakerFeeRate       float64   `json:"maker_fee_rate"`
	Status             uint8     `json:"status"`
	PoolStats          PoolStats `json:"pool_stats"`
}

type PoolStats struct {
	PoolID    uint64  `json:"pool_id"`
	BaseCoin  string  `json:"base_coin,omitempty"`
	QuoteCoin string  `json:"quote_coin,omitempty"`
	Symbol    string  `json:"symbol,omitempty"`
	LastPrice string  `json:"last_price,omitempty"`
	Change    float64 `json:"change"`
	High      string  `json:"high"`
	Low       string  `json:"low"`
	Volume    string  `json:"volume"`
	Turnover  string  `json:"turnover"`
	Trades    uint64  `json:"trades"`
	UpdatedAt Time    `json:"updated_at"`
}
