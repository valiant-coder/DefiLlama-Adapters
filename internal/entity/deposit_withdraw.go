package entity

type ReqFirstDeposit struct {
	PublicKey string `json:"public_key" binding:"required"`
	Symbol    string `json:"symbol" binding:"required"`
	ChainName string `json:"chain_name" binding:"required"`
}

type RespFirstDeposit struct {
	Address string `json:"address"`
}

type ReqDeposit struct {
	Symbol    string `json:"symbol" binding:"required"`
	ChainName string `json:"chain_name" binding:"required"`
}

type RespDeposit struct {
	Address string `json:"address"`
}


type RespDepositRecord struct {
	ID        uint64 `json:"id"`
	Address   string `json:"address"`
	CreatedAt string `json:"created_at"`
}
