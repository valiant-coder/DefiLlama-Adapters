package entity

import "exapp-go/internal/db/db"

type Token struct {
	Symbol       string  `json:"symbol"`
	Name         string  `json:"name"`
	Decimals     uint8   `json:"decimals"`
	EOSContract  string  `json:"eos_contract"`
	SupportChain []Chain `json:"support_chain"`
	IconUrl      string  `json:"icon_url"`

	Info TokenInfo `json:"info"`
}

func TokenFromDB(token db.Token) Token {
	chains := make([]Chain, 0)
	for _, chain := range token.Chains {
		if chain.MinDepositAmount.LessThan(chain.ExsatMinDepositAmount) {
			chain.MinDepositAmount = chain.ExsatMinDepositAmount
		}
		chains = append(chains, Chain{
			ChainName: chain.ChainName,
			ChainID:   chain.ChainID,

			MinDepositAmount:  chain.MinDepositAmount.String(),
			MinWithdrawAmount: chain.MinWithdrawAmount.String(),

			WithdrawFee:       chain.WithdrawalFee.String(),
			ExsatWithdrawFee:  chain.ExsatWithdrawFee.String(),
			ExsatTokenAddress: chain.ExsatTokenAddress,
		})
	}
	tokenData := token.TokenInfo.Data()

	links := make([]TokenLink, 0)
	for _, link := range tokenData.Links {
		links = append(links, TokenLink{
			Url:  link.Url,
			Name: link.Name,
		})
	}
	info := TokenInfo{
		Rank:                  tokenData.Rank,
		MarketCapitalization:  tokenData.MarketCapitalization,
		FullyDilutedMarketCap: tokenData.FullyDilutedMarketCap,
		MarketDominance:       tokenData.MarketDominance,
		Volume:                tokenData.Volume,
		VolumeDivMarketCap:    tokenData.VolumeDivMarketCap,
		CirculatingSupply:     tokenData.CirculatingSupply,
		TotalSupply:           tokenData.TotalSupply,
		MaximumSupply:         tokenData.MaximumSupply,
		IssueDate:             tokenData.IssueDate,
		HistoricalHigh:        tokenData.HistoricalHigh,
		HistoricalLow:         tokenData.HistoricalLow,
		HistoricalHighDate:    tokenData.HistoricalHighDate,
		HistoricalLowDate:     tokenData.HistoricalLowDate,
		Links:                 links,
		Intro:                 tokenData.Intro,
	}
	return Token{
		Symbol:      token.Symbol,
		Name:        token.Name,
		Decimals:    token.Decimals,
		EOSContract: token.EOSContractAddress,
		IconUrl:     token.IconUrl,

		SupportChain: chains,
		Info:         info,
	}
}

type Chain struct {
	ChainID   uint8  `json:"chain_id"`
	ChainName string `json:"chain_name"`

	MinDepositAmount  string `json:"min_deposit_amount"`
	MinWithdrawAmount string `json:"min_withdraw_amount"`
	WithdrawFee       string `json:"withdraw_fee"`
	ExsatWithdrawFee  string `json:"exsat_withdraw_fee"`
	ExsatTokenAddress string `json:"exsat_token_address"`
}

type TokenInfo struct {
	Rank                  string      `json:"rank"`
	MarketCapitalization  string      `json:"market_capitalization"`
	FullyDilutedMarketCap string      `json:"fully_diluted_market_cap"`
	MarketDominance       string      `json:"market_dominance"`
	Volume                string      `json:"volume"`
	VolumeDivMarketCap    string      `json:"volume_div_market_cap"`
	CirculatingSupply     string      `json:"circulating_supply"`
	TotalSupply           string      `json:"total_supply"`
	MaximumSupply         string      `json:"maximum_supply"`
	IssueDate             string      `json:"issue_date"`
	HistoricalHigh        string      `json:"historical_high"`
	HistoricalLow         string      `json:"historical_low"`
	HistoricalHighDate    string      `json:"historical_high_date"`
	HistoricalLowDate     string      `json:"historical_low_date"`
	Links                 []TokenLink `json:"links"`
	Intro                 string      `json:"intro"`
}

type TokenLink struct {
	Url  string `json:"url"`
	Name string `json:"name"`
}
