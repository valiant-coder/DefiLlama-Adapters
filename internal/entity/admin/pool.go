package entity_admin

import (
	"exapp-go/internal/db/db"
)

type RespPool struct {
	ID            uint
	PoolID        uint64        `json:"id"`
	BaseSymbol    string        `json:"base_symbol"`
	BaseContract  string        `json:"base_contract"`
	QuoteSymbol   string        `json:"quote_symbol"`
	QuoteContract string        `json:"quote_contract"`
	Symbol        string        `json:"symbol"`
	Status        db.PoolStatus `json:"status"`
	Visible       bool          `json:"visible"`
	CreatedAt     string        `json:"created_at"`
	UpdatedAt     string        `json:"updated_at"`
}

func (r *RespPool) Fill(a *db.Pool) *RespPool {
	r.BaseContract = a.BaseContract
	r.BaseSymbol = a.BaseSymbol
	r.PoolID = a.PoolID
	r.QuoteContract = a.QuoteContract
	r.QuoteSymbol = a.QuoteSymbol
	r.Symbol = a.Symbol
	r.Status = a.Status
	r.Visible = a.Visible
	return r
}

type ReqUpsertPool struct {
	PoolID        uint64        `json:"pool_id" binding:"required"`
	BaseSymbol    string        `json:"base_symbol" binding:"required"`
	BaseContract  string        `json:"base_contract" binding:"required"`
	QuoteSymbol   string        `json:"quote_symbol" binding:"required"`
	QuoteContract string        `json:"quote_contract" binding:"required"`
	Symbol        string        `json:"symbol" binding:"required"`
	Status        db.PoolStatus `json:"status" binding:"required"`
	Visible       bool          `json:"visible" binding:"required"`
}
