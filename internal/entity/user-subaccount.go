package entity

type ReqAddSubAccount struct {
	// sub account name
	Name string `json:"name"`
	// eos permission
	Permission string `json:"permission"`
}

type RespAddSubAccount struct {
	APIKey string `json:"api_key"`
}

// Added new structures for delete and get sub-accounts
type ReqDeleteSubAccount struct {
	Name string `json:"name"`
}

type RespDeleteSubAccount struct {
	Success bool `json:"success"`
}

type SubAccountInfo struct {
	Name       string              `json:"name"`
	EOSAccount string              `json:"eos_account"`
	Permission string              `json:"permission"`
	APIKey     string              `json:"api_key"`
	PublicKeys []string            `json:"public_keys"`
	Balances   []SubAccountBalance `json:"balances"`
}

type SubAccountBalance struct {
	Coin      string        `json:"coin"`
	Balance   string        `json:"balance"`
	USDTPrice string        `json:"usdt_price"`
	Locked    string        `json:"locked"`
	Locks     []LockBalance `json:"locks"`
}
