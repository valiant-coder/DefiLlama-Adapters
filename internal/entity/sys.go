package entity

type SystemInfo struct {
	Version string `json:"version"`
	PayCPU  PayCPU `json:"pay_cpu"`
}


type PayCPU struct {
	Account string `json:"account"`
}
