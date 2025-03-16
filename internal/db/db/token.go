package db

import (
	"context"
	"fmt"

	"github.com/shopspring/decimal"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func init() {
	addMigrateFunc(func(repo *Repo) error {
		return repo.AutoMigrate(
			&Token{},
			&Chain{},
		)
	})
}

type TokenInfo struct {
	Rank                  string      `json:"rank"`
	MarketCapitalization  string      `json:"market_capitalization"`
	FullyDilutedMarketCap string      `json:"fully_diluted_market_cap"`
	MarketDominance       string      `json:"market_dominance"`
	Volume                string      `json:"volume"`
	VolumeDivMarketCap    string      `json:"volume_div_market_cap"`
	CirculatingSupply     string      `json:"circulating_supply"`
	MaximumSupply         string      `json:"maximum_supply"`
	TotalSupply           string      `json:"total_supply"`
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

type ChainInfo struct {
	ChainName    string `json:"chain_name"`
	ChainID      uint8  `json:"chain_id"`
	PermissionID uint64 `json:"permission_id"`

	DepositByBTCBridge bool `json:"deposit_by_btc_bridge"`

	WithdrawalFee     decimal.Decimal `json:"withdrawal_fee"`
	MinWithdrawAmount decimal.Decimal `json:"min_withdraw_amount"`
	MinDepositAmount  decimal.Decimal `json:"min_deposit_amount"`

	ExsatWithdrawFee      decimal.Decimal `json:"exsat_withdraw_fee"`
	ExsatMinDepositAmount decimal.Decimal `json:"exsat_min_deposit_amount"`
	ExsatTokenAddress     string          `json:"exsat_token_address"`
	ExsatTokenDecimals    uint8           `json:"exsat_token_decimals"`
}
type Token struct {
	gorm.Model
	Symbol             string `gorm:"column:symbol;type:varchar(255);not null;uniqueIndex:idx_symbol"`
	Name               string `gorm:"column:name;type:varchar(255);default:null"`
	EOSContractAddress string `gorm:"column:eos_contract_address;type:varchar(255);not null"`
	Decimals           uint8  `gorm:"column:decimals;type:tinyint(3);not null"`
	IconUrl            string `gorm:"column:icon_url;type:varchar(255);default:null"`

	BlockNum uint64 `gorm:"column:block_num;type:bigint(20);default:0"`

	Chains datatypes.JSONSlice[ChainInfo] `gorm:"column:chains;type:json;not null"`

	TokenInfo datatypes.JSONType[TokenInfo] `gorm:"column:token_info;type:json;default:null"`
}

func (t *Token) TableName() string {
	return "tokens"
}

func (r *Repo) UpsertToken(ctx context.Context, token *Token) error {
	return r.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "symbol"}},
		DoUpdates: clause.AssignmentColumns([]string{"block_num"}),
	}).Create(token).Error
}

func (r *Repo) GetToken(ctx context.Context, symbol string) (*Token, error) {
	var token Token
	err := r.WithContext(ctx).Where("symbol = ?", symbol).First(&token).Error
	return &token, err
}

func (r *Repo) ListTokens(ctx context.Context) ([]Token, error) {
	var tokens []Token
	err := r.WithContext(ctx).Find(&tokens).Error
	return tokens, err
}

func (r *Repo) GetAllTokens(ctx context.Context) (map[string]string, error) {
	tokens, err := r.ListTokens(ctx)
	if err != nil {
		return nil, err
	}
	tokenMap := make(map[string]string)
	for _, token := range tokens {
		tokenMap[fmt.Sprintf("%s-%s", token.EOSContractAddress, token.Symbol)] = token.Symbol
	}
	return tokenMap, nil
}

func (r *Repo) InsertToken(ctx context.Context, token *Token) error {
	return r.WithContext(ctx).Create(token).Error
}

func (r *Repo) GetTokenMaxBlockNum(ctx context.Context) (uint64, error) {
	var blockNum *uint64
	err := r.WithContext(ctx).Model(&Token{}).Select("COALESCE(MAX(block_num), 0)").Scan(&blockNum).Error
	if err != nil {
		return 0, err
	}
	return *blockNum, nil
}

type Chain struct {
	gorm.Model
	ChainName    string `gorm:"column:chain_name;type:varchar(255);not null"`
	ChainID      uint8  `gorm:"column:chain_id;type:tinyint(3);not null;uniqueIndex:idx_chain_id"`
	PermissionID uint64 `gorm:"column:permission_id;type:bigint(20);not null"`
	BlockNum     uint64 `gorm:"column:block_num;type:bigint(20);default:0"`
}

func (c *Chain) TableName() string {
	return "chains"
}

func (r *Repo) UpsertChain(ctx context.Context, chain *Chain) error {
	return r.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "chain_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"block_num"}),
	}).Create(chain).Error
}

func (r *Repo) GetChain(ctx context.Context, chainID uint8) (*Chain, error) {
	var chain Chain
	err := r.WithContext(ctx).Where("chain_id = ?", chainID).First(&chain).Error
	return &chain, err
}

func (r *Repo) GetChainMaxBlockNum(ctx context.Context) (uint64, error) {
	var blockNum *uint64
	err := r.WithContext(ctx).Model(&Chain{}).Select("COALESCE(MAX(block_num), 0)").Scan(&blockNum).Error
	if err != nil {
		return 0, err
	}
	return *blockNum, nil
}
