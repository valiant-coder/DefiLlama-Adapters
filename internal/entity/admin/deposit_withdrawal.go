package entity_admin

import (
	"exapp-go/internal/db/db"

	"github.com/shopspring/decimal"
)

type RespGetDepositWithdrawalTotal struct {
	Symbol string          `json:"symbol"`
	Amount decimal.Decimal `json:"amount"`
}

func (r *RespGetDepositWithdrawalTotal) FillDepositRecord(a *db.DepositRecord) *RespGetDepositWithdrawalTotal {
	r.Symbol = a.Symbol
	r.Amount = a.Amount
	return r
}

func (r *RespGetDepositWithdrawalTotal) FillWithdrawRecord(a *db.WithdrawRecord) *RespGetDepositWithdrawalTotal {
	r.Symbol = a.Symbol
	r.Amount = a.Amount
	return r
}
