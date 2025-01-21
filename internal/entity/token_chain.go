package entity

type Token struct {
	Symbol       string  `json:"symbol"`
	Name         string  `json:"name"`
	Decimals     uint8   `json:"decimals"`
	EOSContract  string  `json:"eos_contract"`
	SupportChain []Chain `json:"support_chain"`
}

type Chain struct {
	ChainID   uint8  `json:"chain_id"`
	ChainName string `json:"chain_name"`

	MinDepositAmount  string `json:"min_deposit_amount"`
	MinWithdrawAmount string `json:"min_withdraw_amount"`
	WithdrawFee       string `json:"withdraw_fee"`
	ExsatWithdrawFee  string `json:"exsat_withdraw_fee"`
}
