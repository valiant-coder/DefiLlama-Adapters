package onedex

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/eoscanada/eos-go"
)

type OneDexSupportToken struct {
	ID              uint64 `json:"id"`
	Contract        string `json:"contract"`
	Sym             string `json:"sym"`
	Symbol          string `json:"symbol"`
	DepositEnabled  uint8  `json:"deposit_enabled"`
	WithdrawEnabled uint8  `json:"withdraw_enabled"`
	WithdrawFee     string `json:"withdraw_fee"`
}

/*
[

	{
	    "id": 1,
	    "contract": "token.1dex",
	    "sym": "6,USDT",
	    "deposit_enabled": 1,
	    "withdraw_enabled": 1,
	    "withdraw_fee": "1.000000 USDT"
	},

]
*/
func GetOneDexSupportTokens(ctx context.Context, nodeUrl string, onedexContract string) ([]*OneDexSupportToken, error) {
	api := eos.New(nodeUrl)

	request := eos.GetTableRowsRequest{
		Code:       onedexContract,
		Scope:      onedexContract,
		Table:      "tokens",
		JSON:       true,
		LowerBound: "0",
		UpperBound: "-1",
		Limit:      100,
	}

	response, err := api.GetTableRows(ctx, request)
	if err != nil {
		return nil, err
	}

	var tokens []*OneDexSupportToken
	err = json.Unmarshal(response.Rows, &tokens)
	if err != nil {
		return nil, err
	}
	for _, token := range tokens {
		token.Symbol = strings.Split(token.Sym, ",")[1]
	}

	return tokens, nil
}
