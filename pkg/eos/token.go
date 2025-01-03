package eos

import (
	"context"

	"github.com/eoscanada/eos-go"
	"github.com/shopspring/decimal"
)

type TokenStats struct {
	Supply    decimal.Decimal `json:"supply"`
	MaxSupply decimal.Decimal `json:"max_supply"`
	Issuer    string          `json:"issuer"`
	Precision uint8           `json:"precision"`
}


func GetTokenStats(ctx context.Context, nodeUrl string, token, symbol string) (TokenStats, error) {
	api := eos.New(nodeUrl)

	response, err := api.GetCurrencyStats(ctx, eos.AN(token), symbol)
	if err != nil {
		return TokenStats{}, err
	}
	precision := response.MaxSupply.Precision

	tokenStats := TokenStats{
		Supply:    decimal.New(int64(response.Supply.Amount), int32(precision)),
		MaxSupply: decimal.New(int64(response.MaxSupply.Amount), int32(precision)),
		Issuer:    response.Issuer.String(),
		Precision: precision,
	}
	return tokenStats, nil
}

