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
		Symbol:             "USDT",
		Name:               "TetherUS",
		Decimals:           6,
		EOSContractAddress: "asdftokencnt",
		Chains: []ChainInfo{
			{
				ChainID:           1,
				ChainName:         "sepolia",
				WithdrawalFee:     decimal.NewFromFloat(1.0),
				ExsatDepositLimit: decimal.NewFromFloat(0.01),
				ExsatWithdrawMax:  decimal.NewFromFloat(100000000.0),
				ExsatDepositFee:   decimal.NewFromFloat(0.0),
				ExsatWithdrawFee:  decimal.NewFromFloat(0.01),
				MinWithdrawAmount: decimal.NewFromFloat(5.0),
				ExsatTokenAddress:  "0x591578a39ee3c6f3751873f74172cc0f708a09b6",
				ExsatTokenDecimals: 6,
				ExsatHelperAddress: "0xdDCD3c161e452afB52e4EDC7620390c62F4676dC",
			},
		},
	}
	err := r.InsertToken(context.Background(), token)
	if err != nil {
		t.Fatal(err)
	}

}
