package db

import (
	"context"

	"github.com/shopspring/decimal"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

func init() {
	addMigrateFunc(func(repo *Repo) error {
		return repo.AutoMigrate(
			&Token{},
		)
	})
}

type TokenType string

const (
	TokenTypeEOS   TokenType = "eos_native"
	TokenTypeBTC   TokenType = "btc_native"
	TokenTypeExsat TokenType = "exsat_bridge"
)

type ChainInfo struct {
	ChainName    string `json:"chain_name"`
	ChainID      uint8  `json:"chain_id"`
	PermissionID uint64 `json:"permission_id"`

	WithdrawalFee     decimal.Decimal `json:"withdrawal_fee"`
	MinWithdrawAmount decimal.Decimal `json:"min_withdraw_amount"`
	ExsatWithdrawFee  decimal.Decimal `json:"exsat_withdraw_fee"`
	ExsatDepositLimit decimal.Decimal `json:"exsat_deposit_limit"`
	ExsatWithdrawMax  decimal.Decimal `json:"exsat_withdraw_max"`
	ExsatDepositFee   decimal.Decimal `json:"exsat_deposit_fee"`
	ExsatTokenAddress  string                         `gorm:"column:exsat_token_address;type:varchar(255);not null"`
	ExsatTokenDecimals uint8                          `gorm:"column:exsat_token_decimals;type:tinyint(3);not null"`
	ExsatHelperAddress string                         `gorm:"column:exsat_helper_address;type:varchar(255);not null"`
}
type Token struct {
	gorm.Model
	TokenType          TokenType                      `gorm:"column:token_type;type:varchar(255);default:exsat_bridge"`
	Symbol             string                         `gorm:"column:symbol;type:varchar(255);not null;uniqueIndex:idx_symbol"`
	Name               string                         `gorm:"column:name;type:varchar(255);default:null"`
	EOSContractAddress string                         `gorm:"column:eos_contract_address;type:varchar(255);not null"`
	Decimals           uint8                          `gorm:"column:decimals;type:tinyint(3);not null"`
	
	Chains             datatypes.JSONSlice[ChainInfo] `gorm:"column:chains;type:json;not null"`
}

func (t *Token) TableName() string {
	return "tokens"
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


func (r *Repo) InsertToken(ctx context.Context, token *Token) error {
	return r.WithContext(ctx).Create(token).Error
}
