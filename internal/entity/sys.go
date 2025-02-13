package entity

type SystemInfo struct {
	Version           string `json:"version"`
	PayCPU            PayCPU `json:"pay_cpu"`
	VaultEVMAddress   string `json:"vault_evm_address"`
	VaultEOSAddress   string `json:"vault_eos_address"`
}

type PayCPU struct {
	Account string `json:"account"`
}
