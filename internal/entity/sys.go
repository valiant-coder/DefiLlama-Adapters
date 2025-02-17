package entity

type SystemInfo struct {
	Version         string `json:"version"`
	PayCPU          PayCPU `json:"pay_cpu"`
	VaultEVMAddress string `json:"vault_evm_address"`
	VaultEOSAddress string `json:"vault_eos_address"`
	TokenContract   string `json:"token_contract"`
}

type PayCPU struct {
	Account string `json:"account"`
}

type SysTradeInfo struct {
	TotalUserCount int64   `json:"total_user_count"`
	TotalTrades    uint64  `json:"total_trades"`
	TotalVolume    float64 `json:"total_volume"`
}
