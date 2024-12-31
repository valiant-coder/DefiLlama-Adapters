package entity

import (
	"exapp-go/internal/db/ckhdb"
	"exapp-go/internal/db/db"
)

type OpenOrder struct {
	ID            uint64 `json:"id"`
	PoolID        uint64 `json:"pool_id"`
	ClientOrderID string `json:"order_cid"`
	Trader        string `json:"trader"`
	OrderTime     Time   `json:"order_time"`
	// 0 buy 1 sell
	Side uint8 `json:"side"`
	// 0 market 1 limit
	Type           uint8  `json:"type"`
	OrderPrice     string `json:"order_price"`
	AvgPrice       string `json:"avg_price"`
	OrderAmount    string `json:"order_amount"`
	ExecutedAmount string `json:"executed_amount"`
	OrderTotal     string `json:"order_total"`
}

func OpenOrderFromDB(openOrder db.OpenOrder) OpenOrder {
	side := uint8(1)
	if openOrder.IsBid {
		side = 0
	}
	return OpenOrder{
		ID:             openOrder.OrderID,
		PoolID:         openOrder.PoolID,
		ClientOrderID:  openOrder.ClientOrderID,
		Trader:         openOrder.Trader,
		OrderPrice:     openOrder.Price.String(),
		AvgPrice:       openOrder.Price.String(),
		OrderAmount:    openOrder.OriginalQuantity.String(),
		ExecutedAmount: openOrder.ExecutedQuantity.String(),
		OrderTotal:     openOrder.OriginalQuantity.String(),
		OrderTime:      Time(openOrder.CreatedAt),
		Side:           side,
		Type:           1,
	}
}

type OrderStatus uint8

const (
	OrderStatusOpen OrderStatus = iota
	OrderStatusPartiallyFilled
	OrderStatusFilled
	OrderStatusCancelled
)

type HistoryOrder struct {
	OrderTime     Time   `json:"order_time"`
	ID            uint64 `json:"id"`
	PoolID        uint64 `json:"pool_id"`
	ClientOrderID string `json:"order_cid"`
	// 0 buy 1 sell
	Side uint8 `json:"side"`
	// 0 market 1 limit
	Type           uint8  `json:"type"`
	OrderPrice     string `json:"order_price"`
	AvgPrice       string `json:"avg_price"`
	OrderAmount    string `json:"order_amount"`
	ExecutedAmount string `json:"executed_amount"`
	FilledTotal    string `json:"filled_total"`
	// 1partially_filled 2full_filled 3.canceled
	Status uint8 `json:"status"`
}

func HistoryOrderFromDB(order ckhdb.HistoryOrder) HistoryOrder {
	side := uint8(1)
	if order.IsBid {
		side = 0
	}
	orderType := uint8(1)
	if order.IsMarket {
		orderType = 0
	}
	return HistoryOrder{
		OrderTime:      Time(order.CreatedAt),
		ID:             order.OrderID,
		PoolID:         order.PoolID,
		ClientOrderID:  order.ClientOrderID,
		Side:           side,
		Type:           orderType,
		OrderPrice:     order.Price.String(),
		AvgPrice:       order.Price.String(),
		OrderAmount:    order.OriginalQuantity.String(),
		ExecutedAmount: order.ExecutedQuantity.String(),
		FilledTotal:    order.ExecutedQuantity.Mul(order.Price).String(),
		Status:         uint8(order.Status),
	}
}

type HistoryOrderDetail struct {
	HistoryOrder
	Trades []TradeDetail `json:"trades"`
}


func HistoryOrderDetailFromDB(order *ckhdb.OrderWithTrades) HistoryOrderDetail {
	return HistoryOrderDetail{
		HistoryOrder: HistoryOrderFromDB(order.HistoryOrder),
		Trades:       TradeDetailFromDB(order.Trades),
	}
}

func TradeDetailFromDB(trades []ckhdb.Trade) []TradeDetail {
	var result []TradeDetail
	for _, trade := range trades {
		result = append(result, TradeDetail{
			PoolID:        trade.PoolID,
			TxID:          trade.TxID,
			Taker:         trade.Taker,
			Maker:         trade.Maker,
			MakerOrderID:  trade.MakerOrderID,
			MakerOrderCID: trade.MakerOrderCID,
			TakerOrderID:  trade.TakerOrderID,
			TakerOrderCID: trade.TakerOrderCID,
			Price:         trade.Price.String(),
			TakerIsBid:    trade.TakerIsBid,
			BaseQuantity:  trade.BaseQuantity.String(),
			QuoteQuantity: trade.QuoteQuantity.String(),
			TakerFee:      trade.TakerFee.String(),
			MakerFee:      trade.MakerFee.String(),
			Timestamp:     Time(trade.Time),
		})
	}
	return result
}
