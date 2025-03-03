package entity

type UserProfit struct {
	UID    string `json:"uid"`
	Avatar string `json:"avatar"`
	Profit string `json:"profit"`
	Score  int    `json:"score"`
}

type UserProfitRank struct {
	Items []UserProfit `json:"items"`

	Avatar     string `json:"avatar"`
	UserProfit string `json:"user_profit"`
	Score      int    `json:"score"`
	Rank       int    `json:"rank"`
}
