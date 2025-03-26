package entity_admin

import (
	"exapp-go/internal/db/db"

	"github.com/shopspring/decimal"
)

type RespGetDepositWithdrawal struct {
	Symbol string          `json:"symbol"`
	Amount decimal.Decimal `json:"amount"`
}

func (r *RespGetDepositWithdrawal) Fill(a *db.DepositRecord) *RespGetDepositWithdrawal {
	r.Symbol = a.Symbol
	r.Amount = a.Amount
	return r
}
