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

type UserPoolBalance struct {
	PoolID        uint64 `json:"pool_id"`
	PoolSymbol    string `json:"pool_symbol"`
	TokenContract string `json:"token_contract"`
	TokenSymbol   string `json:"token_symbol"`
	Balance       string `json:"balance"`
}

type UserTokenBalance struct {
	Contract string `json:"contract"`
	Symbol   string `json:"symbol"`
	Balance  string `json:"balance"`
}

type UserBalance struct {
	PoolBalances []UserPoolBalance `json:"pool_balances"`
	TokenBalances []UserTokenBalance `json:"token_balances"`
}
