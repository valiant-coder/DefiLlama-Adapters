package entity

type UserProfit struct {
	UID    string `json:"uid"`
	Profit string `json:"profit"`
	Score  int    `json:"score"`
}

type UserProfitRank struct {
	Items []UserProfit `json:"items"`

	UserProfit string `json:"user_profit"`
	Score      int    `json:"score"`
	Rank       int    `json:"rank"`
}

