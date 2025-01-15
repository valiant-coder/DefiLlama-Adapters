package db

import (
	"context"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

func init() {
	addMigrateFunc(func(repo *Repo) error {
		return repo.AutoMigrate(
			&Token{},
		)
	})
}

type Token struct {
	gorm.Model
	Symbol       string `gorm:"column:symbol;type:varchar(255);not null;uniqueIndex:idx_symbol_chain_name"`
	ChainName    string `gorm:"column:chain_name;type:varchar(255);not null;uniqueIndex:idx_symbol_chain_name"`
	PermissionID uint64 `gorm:"column:permission_id;type:bigint(20);not null"`
	Decimals     uint8  `gorm:"column:decimals;type:tinyint(3);not null"`

	EOSContractAddress string          `gorm:"column:eos_contract_address;type:varchar(255);not null"`
	ExsatTokenAddress  string          `gorm:"column:exsat_token_address;type:varchar(255);not null"`
	WithdrawalFee      decimal.Decimal `gorm:"column:withdrawal_fee;type:decimal(36,18);not null"`

	ExsatHelperAddress string          `gorm:"column:exsat_helper_address;type:varchar(255);not null"`
	ExsatDepositLimit  decimal.Decimal `gorm:"column:exsat_deposit_limit;type:decimal(36,18);not null;default:0"`
	ExsatWithdrawMax   decimal.Decimal `gorm:"column:exsat_withdraw_max;type:decimal(36,18);not null;default:0"`
	ExsatDepositFee    decimal.Decimal `gorm:"column:exsat_deposit_fee;type:decimal(36,18);not null;default:0"`
	ExsatWithdrawFee   decimal.Decimal `gorm:"column:exsat_withdraw_fee;type:decimal(36,18);not null;default:0"`
}

func (t *Token) TableName() string {
	return "tokens"
}

func (r *Repo) GetToken(ctx context.Context, symbol string, chainName string) (*Token, error) {
	var token Token
	err := r.WithContext(ctx).Where("symbol = ? and chain_name = ?", symbol, chainName).First(&token).Error
	return &token, err
}

func (r *Repo) ListTokens(ctx context.Context) ([]Token, error) {
	var tokens []Token
	err := r.WithContext(ctx).Find(&tokens).Error
	return tokens, err
}
