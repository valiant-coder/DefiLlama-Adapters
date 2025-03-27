package entity_admin

import (
	"exapp-go/internal/db/db"
	"exapp-go/internal/entity"

	"github.com/shopspring/decimal"
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
	CreatedAt string           `json:"created_at"`
	UpdatedAt string           `json:"updated_at"`
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
		ID:          token.ID,
		IconUrl:     token.IconUrl,
		Symbol:            token.Symbol,
		Name:              token.Name,
		Decimals:          token.Decimals,
		ExsatTokenAddress: token.ExsatTokenAddress,
		EOSContract:       token.EOSContractAddress,

		Chains:    chains,
		Info:      info,
		UpdatedAt: token.UpdatedAt.Format("2006-01-02 15:04:05"),
		CreatedAt: token.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}

type ReqCreateToken struct {
	IconUrl     string           `json:"icon_url"`
	Symbol      string           `json:"symbol"`
	Name        string           `json:"name"`
	Decimals    uint8            `json:"decimals"`
	EVMContract string           `json:"evm_contract"`
	EOSContract string           `json:"eos_contract"`
	MaxSupply   string           `json:"max_supply"`
	WithdrawFee string           `json:"withdraw_fee"`
	BlockNum    uint64           `json:"block_num"`
	Chains      []entity.Chain   `json:"chains"`
	Info        entity.TokenInfo `json:"info"`
}

func DBFromCreateToken(token *ReqCreateToken) *db.Token {
	chains := make([]db.ChainInfo, 0)
	for _, chain := range token.Chains {
		exsatWithdrawFee, _ := decimal.NewFromString(chain.ExsatWithdrawFee)
		minDepositAmount, _ := decimal.NewFromString(chain.MinDepositAmount)
		minWithdrawAmount, _ := decimal.NewFromString(chain.MinWithdrawAmount)
		withdrawFee, _ := decimal.NewFromString(chain.WithdrawFee)

		chains = append(chains, db.ChainInfo{
			ChainID:           chain.ChainID,
			ChainName:         chain.ChainName,
			ExsatTokenAddress: chain.ExsatTokenAddress,
			ExsatWithdrawFee:  exsatWithdrawFee,
			MinDepositAmount:  minDepositAmount,
			MinWithdrawAmount: minWithdrawAmount,
			WithdrawFee:       withdrawFee,
		})
	}

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

	maxSupply, _ := decimal.NewFromString(token.MaxSupply)
	withdrawFee, _ := decimal.NewFromString(token.WithdrawFee)

	return &db.Token{
		Symbol:             token.Symbol,
		Name:               token.Name,
		Decimals:           token.Decimals,
		ExsatTokenAddress:  token.EVMContract,
		EOSContractAddress: token.EOSContract,
		IconUrl:            token.IconUrl,
		MaxSupply:          maxSupply,
		WithdrawFee:        withdrawFee,
		BlockNum:           token.BlockNum,
		Chains:             chains,
		TokenInfo:          datatypes.NewJSONType(tokenInfo),
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
