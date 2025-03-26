package entity

type ReqSendTx struct {
	SingleSignedTransaction string `json:"single_signed_transaction" binding:"required"`
}

type RespSendTx struct {
	TransactionID string `json:"transaction_id"`
}

type RespSystemInfo struct {
	Version       string `json:"version"`
	PayEOSAccount string `json:"pay_eos_account"`
	TokenContract string `json:"token_contract"`
	AppContract   string `json:"one_dex_contract"`
}

type RespV1UserInfo struct {
	EOSAccount string `json:"eos_account"`
	Permission string `json:"permission"`
}
