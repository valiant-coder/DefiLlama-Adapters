package entity

type SystemInfo struct {
	Version            string             `json:"version"`
	PayCPU             PayCPU             `json:"pay_cpu"`
	EVMAgentContract   string             `json:"evm_agent_contract"`
	VaultEVMAddress    string             `json:"vault_evm_address"`
	VaultEOSAddress    string             `json:"vault_eos_address"`
	TokenContract      string             `json:"token_contract"`
	AppContract        string             `json:"app_contract"`
	ExsatNetwork       ExsatNetwork       `json:"exsat_network"`
	TradingCompetition TradingCompetition `json:"trading_competition"`
}

type TradingCompetition struct {
	BeginTime         Time  `json:"begin_time"`
	EndTime           Time  `json:"end_time"`
	DailyPoints       []int `json:"daily_points"`
	AccumulatedPoints []int `json:"accumulated_points"`
}

type PayCPU struct {
	Account string `json:"account"`
}

type SysTradeInfo struct {
	TotalUserCount int64   `json:"total_user_count"`
	TotalTrades    uint64  `json:"total_trades"`
	TotalVolume    float64 `json:"total_volume"`
}

type ExsatNetwork struct {
	CurrencySymbol   string `json:"currency_symbol"`
	NetworkUrl       string `json:"network_url"`
	ChainId          int    `json:"chain_id"`
	NetworkName      string `json:"network_name"`
	BlockExplorerUrl string `json:"block_explorer_url"`
}
