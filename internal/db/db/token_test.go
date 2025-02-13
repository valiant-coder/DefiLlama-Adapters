package db

import (
	"context"
	"exapp-go/config"
	"exapp-go/pkg/utils"
	"testing"

	"github.com/shopspring/decimal"
	"gorm.io/datatypes"
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
				ChainID:               64,
				ChainName:             "btc",
				WithdrawalFee:         decimal.NewFromFloat(0.0001),
				MinWithdrawAmount:     decimal.NewFromFloat(0.0001),
				MinDepositAmount:      decimal.NewFromFloat(0.0001),
				ExsatMinDepositAmount: decimal.NewFromFloat(0.0001),

				ExsatWithdrawFee: decimal.NewFromFloat(0.0001),
			},
			{
				ChainName:             "eos",
				ChainID:               110,
				WithdrawalFee:         decimal.NewFromFloat(0),
				MinWithdrawAmount:     decimal.NewFromFloat(0.0001),
				MinDepositAmount:      decimal.NewFromFloat(0.0001),
				ExsatMinDepositAmount: decimal.NewFromFloat(0.0001),
				ExsatWithdrawFee:      decimal.NewFromFloat(0.0001),
			},
		},
		TokenInfo: datatypes.NewJSONType(TokenInfo{
			Rank:                  "1",
			MarketCapitalization:  "$1.91T",
			FullyDilutedMarketCap: "$2.02T",
			MarketDominance:       "59.7%",
			Volume:                "$48.41B",
			VolumeDivMarketCap:    "2.54%",
			CirculatingSupply:     "19,823,540 BTC",
			MaximumSupply:         "21,000,000 BTC",
			TotalSupply:           "19,823,540 BTC",
			IssueDate:             "2008-11-01",
			HistoricalHigh:        "$109,114.88483408831",
			HistoricalLow:         "$0.04864654",
			HistoricalHighDate:    "2025-01-20",
			HistoricalLowDate:     "2010-07-15",
			Links: []TokenLink{
				{
					Url:  "https://bitcoin.org/en/",
					Name: "Official Website",
				},
			},
			Intro: `Bitcoin (BTC) is a peer-to-peer cryptocurrency that aims to function as a means of exchange that is independent of any central authority. BTC can be transferred electronically in a secure, verifiable, and immutable way.
Launched in 2009, BTC is the first virtual currency to solve the double-spending issue by timestamping transactions before broadcasting them to all of the nodes in the Bitcoin network. The Bitcoin Protocol offered a solution to the Byzantine Generals' Problem with a blockchain network structure, a notion first created by Stuart Haber and W. Scott Stornetta in 1991.`,
		}),
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
