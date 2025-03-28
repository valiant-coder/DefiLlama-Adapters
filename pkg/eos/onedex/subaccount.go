package onedex

import (
	"context"
	"encoding/json"

	"github.com/eoscanada/eos-go"
	"github.com/shopspring/decimal"
)

type Balance struct {
	Id             uint64          `json:"id"`
	Contract       string          `json:"contract"`
	Balance        string          `json:"balance"`
	BalanceDecimal decimal.Decimal `json:"balance_decimal"`
	Symbol         string          `json:"symbol"`
}

func GetSubaccountBalances(ctx context.Context, client *eos.API, account, permission string) ([]*Balance, error) {

	request := eos.GetTableRowsRequest{
		Code:       account,
		Scope:      permission,
		Table:      "funds",
		JSON:       true,
		LowerBound: "0",
		UpperBound: "-1",
		Limit:      100,
	}

	response, err := client.GetTableRows(ctx, request)
	if err != nil {
		return nil, err
	}

	var balances []*Balance
	err = json.Unmarshal(response.Rows, &balances)
	if err != nil {
		return nil, err
	}

	for _, balance := range balances {
		asset, err := eos.NewAssetFromString(balance.Balance)
		if err != nil {
			return nil, err
		}
		balance.Symbol = asset.Symbol.Symbol
		balance.BalanceDecimal = decimal.NewFromInt(int64(asset.Amount)).Shift(-int32(asset.Symbol.Precision))
	}
	return balances, nil

}
