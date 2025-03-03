package entity

type UserProfit struct {
	UID    string `json:"uid"`
	Avatar string `json:"avatar"`
	Profit string `json:"profit"`
	Point  int    `json:"point"`
}

type UserProfitRank struct {
	Items []UserProfit `json:"items"`

	Avatar     string `json:"avatar"`
	UserProfit string `json:"user_profit"`
	Point      int    `json:"point"`
	Rank       int    `json:"rank"`
}
