package entity_admin

import (
	"exapp-go/internal/db/db"

	"github.com/shopspring/decimal"
)

type RespGetDepositWithdrawTotal struct {
	Symbol string          `json:"symbol"`
	Amount decimal.Decimal `json:"amount"`
}

func (r *RespGetDepositWithdrawTotal) FillDepositRecord(a *db.DepositRecord) *RespGetDepositWithdrawTotal {
	r.Symbol = a.Symbol
	r.Amount = a.Amount
	return r
}

func (r *RespGetDepositWithdrawTotal) FillWithdrawRecord(a *db.WithdrawRecord) *RespGetDepositWithdrawTotal {
	r.Symbol = a.Symbol
	r.Amount = a.Amount
	return r
}
