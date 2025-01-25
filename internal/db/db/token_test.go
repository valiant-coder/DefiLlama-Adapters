package db

import (
	"context"
	"exapp-go/config"
	"exapp-go/pkg/utils"
	"testing"

	"github.com/shopspring/decimal"
)



func TestRepo_InsertToken(t *testing.T) {
	utils.WorkInProjectPath("exapp-go")
	config.Load("config/config_dev.yaml")
	r := New()
	token := &Token{
		Symbol:             "BTC",
		Name:               "Bitcoin",
		Decimals:           8,
		EOSContractAddress: "btc.xsat",
		Chains: []ChainInfo{
			{
				ChainID:           64,
				ChainName:         "btc",
				WithdrawalFee:     decimal.NewFromFloat(0.0001),
				MinWithdrawAmount: decimal.NewFromFloat(0.0001),
				ExsatDepositLimit: decimal.NewFromFloat(0.0001),

				ExsatWithdrawFee: decimal.NewFromFloat(0.0001),
			},
			{
				ChainName:         "eos",
				ChainID:           110,
				WithdrawalFee:     decimal.NewFromFloat(0),
				MinWithdrawAmount: decimal.NewFromFloat(0.0001),
				ExsatDepositLimit: decimal.NewFromFloat(0.0001),

				ExsatWithdrawFee: decimal.NewFromFloat(0.0001),
			},
		},
	}
	err := r.InsertToken(context.Background(), token)
	if err != nil {
		t.Fatal(err)
	}

}

func TestRepo_GetToken(t *testing.T) {
	utils.WorkInProjectPath("exapp-go")
	config.Load("config/config_dev.yaml")
	r := New()
	token, err := r.GetToken(context.Background(), "USDT")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(token)
}
