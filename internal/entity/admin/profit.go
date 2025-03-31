package entity_admin

import (
	"exapp-go/internal/db/db"

	"github.com/shopspring/decimal"
)

type RespCoinBalance struct {
	Coin   string          `json:"coin"`
	Amount decimal.Decimal `json:"amount"`
}

func (r *RespCoinBalance) Fill(a *db.UserCoinBalanceRecord) *RespCoinBalance {
	r.Coin = a.Coin
	r.Amount = a.Amount
	return r
}

type RespUserBalance struct {
	Username  string          `json:"username"`
	IsEvmUser bool            `json:"is_evm_user"`
	UID       string          `json:"uid"`
	Balance   decimal.Decimal `json:"balance"`
}

func (r *RespUserBalance) Fill(a *db.UserBalanceWithUsername) *RespUserBalance {
	r.Username = a.Username
	r.IsEvmUser = a.IsEvmUser
	r.UID = a.UID
	r.Balance = a.USDTAmount
	return r
}
