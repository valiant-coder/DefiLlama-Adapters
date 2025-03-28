package onedex

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/eoscanada/eos-go"
)

/*
	{
	        "id": 0,
	        "token_contract": "exsat.xsat",
	        "address": "8266f2fbc720012e5ac038ad3dbb29d2d613c459",
	        "ingress_fee": "0.00000000 XSAT",
	        "balance": "984303.40517803 XSAT",
	        "fee_balance": "0.04263700 XSAT",
	        "erc20_precision": 18,
	        "from_evm_to_native": 0,
	        "original_erc20_token_address": ""
	    },
*/
type EvmEosTokenMapping struct {
	Id                        uint64 `json:"id"`
	TokenContract             string `json:"token_contract"`
	Address                   string `json:"address"`
	Erc20Precision            uint8  `json:"erc20_precision"`
	FromEvmToNative           uint8  `json:"from_evm_to_native"`
	OriginalErc20TokenAddress string `json:"original_erc20_token_address"`
	Balance                   string `json:"balance"`
	Precision                 uint8  `json:"precision,omitempty"`
	Symbol                    string `json:"symbol,omitempty"`
	ExsatTokenAddress         string `json:"exsat_token_address,omitempty"`
}

func GetEvmEosTokenMapping(ctx context.Context, client *eos.API, erc2oContract string) ([]*EvmEosTokenMapping, error) {
	request := eos.GetTableRowsRequest{
		Code:       erc2oContract,
		Scope:      erc2oContract,
		Table:      "tokens",
		JSON:       true,
		LowerBound: "0",
		UpperBound: "-1",
		Limit:      500,
	}

	response, err := client.GetTableRows(ctx, request)
	if err != nil {
		return nil, err
	}

	var mappings []*EvmEosTokenMapping
	err = json.Unmarshal(response.Rows, &mappings)
	if err != nil {
		return nil, err
	}

	for _, mapping := range mappings {
		asset, err := eos.NewAssetFromString(mapping.Balance)
		if err != nil {
			return nil, err
		}
		mapping.Precision = asset.Symbol.Precision
		mapping.Symbol = asset.Symbol.Symbol
		if mapping.FromEvmToNative == 0 {
			mapping.ExsatTokenAddress = strings.ToLower("0x" + mapping.Address)
		} else {
			mapping.ExsatTokenAddress = strings.ToLower("0x" + mapping.OriginalErc20TokenAddress)
		}
	}
	return mappings, nil

}
