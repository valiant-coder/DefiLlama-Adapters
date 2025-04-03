package entity_admin

import (
	"exapp-go/internal/db/db"
	"exapp-go/internal/entity"

	"gorm.io/datatypes"
)

type RespToken struct {
	ID uint `json:"id"`
	//
	IconUrl           string         `json:"icon_url"`
	Symbol            string         `json:"symbol"`
	Name              string         `json:"name"`
	Decimals          uint8          `json:"decimals"`
	ExsatTokenAddress string         `json:"exsat_token_address"`
	EOSContract       string         `json:"eos_contract"`
	WithdrawFee       string         `json:"withdraw_fee"`
	Chains            []entity.Chain `json:"chains"`
	//
	Info      entity.TokenInfo `json:"info"`
	CreatedAt entity.Time      `json:"created_at"`
	UpdatedAt entity.Time      `json:"updated_at"`
}

func TokenFromDB(token *db.Token) *RespToken {
	chains := make([]entity.Chain, 0)
	for _, chain := range token.Chains {
		if chain.MinDepositAmount.LessThan(chain.ExsatMinDepositAmount) {
			chain.MinDepositAmount = chain.ExsatMinDepositAmount
		}
		chains = append(chains, entity.Chain{
			ChainName: chain.ChainName,
			ChainID:   chain.ChainID,

			MinDepositAmount:  chain.MinDepositAmount.String(),
			MinWithdrawAmount: chain.MinWithdrawAmount.String(),

			WithdrawFee:       chain.WithdrawFee.String(),
			ExsatWithdrawFee:  chain.ExsatWithdrawFee.String(),
			ExsatTokenAddress: chain.ExsatTokenAddress,
		})
	}
	tokenData := token.TokenInfo.Data()

	links := make([]entity.TokenLink, 0)
	for _, link := range tokenData.Links {
		links = append(links, entity.TokenLink{
			Url:  link.Url,
			Name: link.Name,
		})
	}
	info := entity.TokenInfo{
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
	return &RespToken{
		ID:                token.ID,
		IconUrl:           token.IconUrl,
		Symbol:            token.Symbol,
		Name:              token.Name,
		Decimals:          token.Decimals,
		ExsatTokenAddress: token.ExsatTokenAddress,
		EOSContract:       token.EOSContractAddress,

		Chains:    chains,
		Info:      info,
		UpdatedAt: entity.Time(token.UpdatedAt),
		CreatedAt: entity.Time(token.CreatedAt),
	}
}

type ReqUpdateToken struct {
	IconUrl string           `json:"icon_url"`
	Info    entity.TokenInfo `json:"info"`
}

func DBFromUpdateToken(token *ReqUpdateToken) *db.Token {
	links := make([]db.TokenLink, 0)
	for _, link := range token.Info.Links {
		links = append(links, db.TokenLink{
			Name: link.Name,
			Url:  link.Url,
		})
	}
	tokenInfo := db.TokenInfo{
		Rank:                  token.Info.Rank,
		MarketCapitalization:  token.Info.MarketCapitalization,
		FullyDilutedMarketCap: token.Info.FullyDilutedMarketCap,
		MarketDominance:       token.Info.MarketDominance,
		Volume:                token.Info.Volume,
		VolumeDivMarketCap:    token.Info.VolumeDivMarketCap,
		CirculatingSupply:     token.Info.CirculatingSupply,
		TotalSupply:           token.Info.TotalSupply,
		MaximumSupply:         token.Info.MaximumSupply,
		IssueDate:             token.Info.IssueDate,
		HistoricalHigh:        token.Info.HistoricalHigh,
		HistoricalLow:         token.Info.HistoricalLow,
		HistoricalHighDate:    token.Info.HistoricalHighDate,
		HistoricalLowDate:     token.Info.HistoricalLowDate,
		Links:                 links,
		Intro:                 token.Info.Intro,
	}
	return &db.Token{
		IconUrl:   token.IconUrl,
		TokenInfo: datatypes.NewJSONType(tokenInfo),
	}
}
