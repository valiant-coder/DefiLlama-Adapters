package entity_admin

import (
	"exapp-go/internal/db/ckhdb"

	"github.com/shopspring/decimal"
)

type RespHistoryOrder struct {
	OrderID      uint64 `json:"order_id"`
	CreatedAt    int64  `json:"created_at"`
	CompleteAt   int64  `json:"complete_at"`
	PoolBaseCoin string `json:"pool_base_coin"`
	// Name
	Trader string `json:"trader"`
	//
	PoolSymbol string `json:"pool_symbol"`
	//
	OriginalQuantity decimal.Decimal `json:"original_quantity"`
	ExecutedQuantity decimal.Decimal `json:"executed_quantity"`
	Price            decimal.Decimal `json:"price"`
	AvgPrice         decimal.Decimal `json:"avg_price"`
	TxID             TxID            `json:"tx_id"`
}

type TxID struct {
	CreateTxID string `json:"create_tx_id"`
	CancelTxID string `json:"cancel_tx_id"`
}

func (r *RespHistoryOrder) Fill(a *ckhdb.HistoryOrder) *RespHistoryOrder {
	r.CreatedAt = a.CreatedAt.Unix()
	r.CompleteAt = a.CanceledAt.Unix()
	r.Trader = a.Trader
	r.PoolSymbol = a.PoolSymbol
	r.Price = a.Price
	r.OriginalQuantity = a.OriginalQuantity
	r.ExecutedQuantity = a.ExecutedQuantity
	r.TxID = TxID{
		CreateTxID: a.CancelTxID,
		CancelTxID: a.CancelTxID,
	}

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
