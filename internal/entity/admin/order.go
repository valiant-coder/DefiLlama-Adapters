package entity_admin

import (
	"exapp-go/internal/db/db"

	"github.com/shopspring/decimal"
)

type RespOpenOrder struct {
	CreatedAt        string          `json:"created_at"`
	CompleteAt       string          `json:"complete_at"`
	Symbol           string          `json:"symbol"`
	Name             string          `json:"name"`
	EVMContract      string          `json:"evm_contract"`
	Trader           string          `json:"trader"`
	Fee              string          `json:"fee"`
	PoolSymbol       string          `json:"pool_symbol"`
	PoolContract     string          `json:"pool_contract"`
	Price            decimal.Decimal `json:"price"`
	OriginalQuantity decimal.Decimal `json:"original_quantity"`
	ExecutedQuantity decimal.Decimal `json:"executed_quantity"`
	TxID             string          `json:"tx_id"`
	Status           db.OrderStatus  `json:"status"`
}

func (r *RespOpenOrder) Fill(a *db.OpenOrder) *RespOpenOrder {
	r.CreatedAt = a.CreatedAt.Format("2006-01-02 15:04:05")
	// r.CompleteAt = a.CompleteAt.Format("2006-01-02 15:04:05")
	// r.Symbol = a.Symbol
	// r.Name = a.Name
	// r.EVMContractAddress = a.EVMContractAddress
	r.Trader = a.Trader
	// r.Fee = a.Fee
	r.PoolSymbol = a.PoolSymbol
	// r.PoolContract = a.PoolContract
	r.Price = a.Price
	r.OriginalQuantity = a.OriginalQuantity
	r.ExecutedQuantity = a.ExecutedQuantity
	r.TxID = a.TxID
	r.Status = a.Status
	return r
}

type RespOrdersCoinTotal struct {
	Symbol string          `json:"symbol"`
	Total  decimal.Decimal `json:"total"`
}
