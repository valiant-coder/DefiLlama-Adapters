package entity

type ReqClaimFaucet struct {
	DepositAddress string `json:"deposit_address"`
}

type RespClaimFaucet struct {
	TxHash string `json:"tx_hash"`
}
