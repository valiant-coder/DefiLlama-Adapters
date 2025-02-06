package entity

type ReqUserLogin struct {
	// google,apple
	Method  string `json:"method"`
	IdToken string `json:"id_token"`
}

type RespUserInfo struct {
	UID      string               `json:"uid"`
	UserName string               `json:"user_name"`
	Passkeys []RespUserCredential `json:"passkeys"`
}

type UserCredential struct {
	DeviceID     string `json:"device_id"`
	CredentialID string `json:"credential_id"`
	PublicKey    string `json:"public_key"`
	Name         string `json:"name"`
	Synced       bool   `json:"synced"`
}


type RespUserCredential struct {
	UserCredential
	CreatedAt     Time     `json:"created_at"`
	LastUsedIP    string   `json:"last_used_ip"`
	LastUsedAt    Time     `json:"last_used_at"`
	EOSAccount    string   `json:"eos_account"`
	EOSPermission []string `json:"eos_permission"`
	DeviceID      string   `json:"device_id"`
}

type UserBalance struct {
	Coin      string        `json:"coin"`
	Balance   string        `json:"balance"`
	USDTPrice string        `json:"usdt_price"`
	Locked    string        `json:"locked"`
	Locks     []LockBalance `json:"locks"`
}

type LockBalance struct {
	PoolID     uint64 `json:"pool_id"`
	PoolSymbol string `json:"pool_symbol"`
	Balance    string `json:"balance"`
}
