package entity

type Token struct {
	Symbol       string  `json:"symbol"`
	Name         string  `json:"name"`
	SupportChain []Chain `json:"support_chain"`
}

type Chain struct {
	ChainID   uint8  `json:"chain_id"`
	ChainName string `json:"chain_name"`
	Decimals  uint8  `json:"decimals"`

	MinDepositAmount  string `json:"min_deposit_amount"`
	MinWithdrawAmount string `json:"min_withdraw_amount"`
	WithdrawFee       string `json:"withdraw_fee"`
	ExsatWithdrawFee  string `json:"exsat_withdraw_fee"`
}
