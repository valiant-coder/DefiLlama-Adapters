package onedex

import (
	"context"

	"github.com/eoscanada/eos-go"
	"github.com/shopspring/decimal"
)

type GetBalancesResponse struct {
	Balances []Balance `json:"balances"`
}

type Balance struct {
	Symbol    string          `json:"symbol"`
	Precision uint8           `json:"precision"`
	Amount    decimal.Decimal `json:"amount"`
	Contract  string          `json:"contract"`
}

func GetSubaccountBalances(ctx context.Context, client *eos.API, account, permission string) (*GetBalancesResponse, error) {
	return nil, nil
}
