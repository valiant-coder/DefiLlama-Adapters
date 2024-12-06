package entity

type Token struct {
	Contract  string `json:"contract"`
	Symbol    string `json:"symbol"`
	Decimals  int    `json:"decimals"`
	LastPrice string `json:"last_price"`
	Change24h string `json:"change_24h"`
}
