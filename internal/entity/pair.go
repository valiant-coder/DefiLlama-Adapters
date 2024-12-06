package entity

type Pair struct {
	ID            string `json:"id"`
	Symbol        string `json:"symbol"`
	BaseSymbol    string `json:"base_symbol"`
	QuoteSymbol   string `json:"quote_symbol"`
	BaseContract  string `json:"base_contract"`
	QuoteContract string `json:"quote_contract"`
}

type PairInfo struct {
	// Trading pair ID
	PairID string `json:"pair_id"`
	// Trading pair symbol
	Symbol string `json:"symbol"`
	// 24h price change percentage
	Change24h float64 `json:"change_24h"`
	// 24h price change
	PriceChange24h float64 `json:"price_change_24h"`
	// 24h trading volume
	Volume24h float64 `json:"volume_24h"`
	// 24h trading amount
	Amount24h float64 `json:"amount_24h"`
	// 24h trade count
	Count24h int64 `json:"count_24h"`
	// 24h highest price
	High24h float64 `json:"high_24h"`
	// 24h lowest price
	Low24h float64 `json:"low_24h"`
	// Latest price
	LastPrice float64 `json:"last_price"`
}
