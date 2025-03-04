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

type TotalTradeStats struct {
	UserClaimedFaucet int     `json:"user_claimed_faucet"`
	UserPoints        int     `json:"user_points"`
	TotalPointsIssued int     `json:"total_points_issued"`
	TotalTradeVolume  float64 `json:"total_trade_volume"`
	TotalTradeUser    int     `json:"total_trade_user"`
}





