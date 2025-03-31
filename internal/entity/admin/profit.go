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
