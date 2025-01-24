package entity

type SystemInfo struct {
	Version           string `json:"version"`
	PayCPU            PayCPU `json:"pay_cpu"`
	VaultEVMAddress   string `json:"vault_evm_address"`
}

type PayCPU struct {
	Account string `json:"account"`
}
