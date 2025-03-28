package entity_admin

import (
	"exapp-go/internal/db/ckhdb"
	"exapp-go/internal/entity"
	"fmt"

	"github.com/shopspring/decimal"
)

type RespHistoryOrder struct {
	ID               string          `json:"id"`
	CreatedAt        entity.Time     `json:"created_at"`
	CompleteAt       entity.Time     `json:"complete_at"`
	PoolBaseCoin     string          `json:"pool_base_coin"`
	Trader           string          `json:"trader"`
	PoolSymbol       string          `json:"pool_symbol"`
	OriginalQuantity decimal.Decimal `json:"original_quantity"`
	ExecutedQuantity decimal.Decimal `json:"executed_quantity"`
	Fee              decimal.Decimal `json:"fee"`
	App              string          `json:"app"`
	Price            decimal.Decimal `json:"price"`
	TxIDs            string          `json:"tx_id"`
}

type TxID struct {
	CreateTxID string `json:"create_tx_id"`
	CancelTxID string `json:"cancel_tx_id"`
}

func (r *RespHistoryOrder) Fill(a *ckhdb.HistoryOrderForm) *RespHistoryOrder {
	side := uint8(1)
	if a.IsBid {
		side = 0
	}
	r.ID = fmt.Sprintf("%d-%d-%d", a.PoolID, a.OrderID, side)
	r.CreatedAt = entity.Time(a.CreatedAt)
	r.CompleteAt = entity.Time(a.CompleteAt)
	r.Trader = a.Trader
	r.PoolBaseCoin = a.PoolBaseCoin
	r.PoolSymbol = a.PoolSymbol
	r.Price = a.Price
	r.Fee = a.Fee
	r.App = a.App
	r.Price = a.Price
	r.ExecutedQuantity = a.ExecutedQuantity
	r.TxIDs = a.TxIDs
	return r
}

type RespOrdersCoinTotal struct {
	Symbol string          `json:"symbol"`
	Total  decimal.Decimal `json:"total"`
}

func (r *RespOrdersCoinTotal) Fill(a *ckhdb.HistoryOrder) *RespOrdersCoinTotal {
	r.Symbol = a.PoolBaseCoin
	r.Total = a.ExecutedQuantity
	return r
}
