package entity

type ReqUserLogin struct {
	// google,apple
	Method  string `json:"method"`
	IdToken string `json:"id_token"`
}

type UserCredential struct {
	CredentialID string `json:"credential_id"`
	PublicKey    string `json:"public_key"`
}

type UserBalance struct {
	Contract string        `json:"contract"`
	Symbol   string        `json:"symbol"`
	Balance  string        `json:"balance"`
	Locked   string        `json:"locked"`
	Locks    []LockBalance `json:"locks"`
}

type LockBalance struct {
	PoolID     uint64 `json:"pool_id"`
	PoolSymbol string `json:"pool_symbol"`
	Symbol     string `json:"symbol"`
	Balance    string `json:"balance"`
}
