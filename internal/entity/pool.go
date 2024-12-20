package entity

type Pool struct {
	ID             uint64   `json:"id" `
	BaseSymbol     string   `json:"base_symbol"`
	BaseContract   string   `json:"base_contract"`
	QuoteSymbol    string   `json:"quote_symbol"`
	QuoteContract  string   `json:"quote_contract"`
	AskingTime     Time     `json:"asking_time"`
	TradingTime    Time     `json:"trading_time"`
	MaxFluctuation uint64   `json:"max_flct"`
	PricePrecision uint8    `json:"price_precision"`
	TakerFeeRate   uint64   `json:"taker_fee_rate"`
	MakerFeeRate   uint64   `json:"maker_fee_rate"`
	Status         uint8    `json:"status"`
	PoolInfo       PoolInfo `json:"pool_info"`
}

type PoolInfo struct {
	PoolID     uint64 `json:"pool_id"`
	Symbol     string `json:"symbol"`
	Change     float64 `json:"change"`
	High       string `json:"high"`
	Low        string `json:"low"`
	Volume     string `json:"volume"`
	Turnover   string `json:"turnover"`
	TradeCount uint64 `json:"trade_count"`
}
