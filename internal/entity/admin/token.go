package entity_admin

import (
	"encoding/json"
	"exapp-go/internal/db/db"
)

type RespToken struct {
	ID uint `json:"id"`
	//
	IconUrl            string `json:"icon_url"`
	Symbol             string `json:"symbol"`
	Name               string `json:"name"`
	EVMContractAddress string `json:"evm_contract_address"`
	ChainNames         string `json:"chain_names"`
	//
	TokenInfo string `json:"token_info"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func (r *RespToken) Fill(a *db.Token) *RespToken {
	chainName := ""
	for _, chain := range a.Chains {
		chainName += chain.ChainName + ", "
	}
	tokenInfo, _ := json.Marshal(a.TokenInfo)

	r.ID = a.ID
	r.IconUrl = a.IconUrl
	r.Symbol = a.Symbol
	r.Name = a.Name
	r.EVMContractAddress = a.EVMContractAddress
	r.ChainNames = chainName
	r.TokenInfo = string(tokenInfo)
	r.CreatedAt = a.CreatedAt.Format("2006-01-02 15:04:05")
	r.UpdatedAt = a.UpdatedAt.Format("2006-01-02 15:04:05")

	return r
}
