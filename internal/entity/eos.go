package entity

type ReqPayCPU struct {
	SingleSignedTransaction string `json:"single_signed_transaction" binding:"required"`
}

type RespPayCPU struct {
	TransactionID string `json:"transaction_id"`
}
