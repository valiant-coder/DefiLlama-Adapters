package entity

import (
	"exapp-go/internal/db/ckhdb"
	"exapp-go/internal/db/db"
	"fmt"
)

type OpenOrder struct {
	ID            string `json:"id"`
	OrderID       uint64 `json:"order_id"`
	PoolID        uint64 `json:"pool_id"`
	PoolSymbol    string `json:"pool_symbol"`
	PoolBaseCoin  string `json:"pool_base_coin"`
	PoolQuoteCoin string `json:"pool_quote_coin"`
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
		ID:             fmt.Sprintf("%d-%d-%d", openOrder.PoolID, openOrder.OrderID, side),
		OrderID:        openOrder.OrderID,
		PoolID:         openOrder.PoolID,
		PoolSymbol:     openOrder.PoolSymbol,
		PoolBaseCoin:   openOrder.PoolBaseCoin,
		PoolQuoteCoin:  openOrder.PoolQuoteCoin,
		ClientOrderID:  openOrder.ClientOrderID,
		Trader:         openOrder.Trader,
		OrderPrice:     openOrder.Price.String(),
		AvgPrice:       openOrder.Price.String(),
		OrderAmount:    openOrder.OriginalQuantity.String(),
		ExecutedAmount: openOrder.ExecutedQuantity.String(),
		OrderTotal:     openOrder.OriginalQuantity.Mul(openOrder.Price).Round(int32(openOrder.QuoteCoinPrecision)).String(),
		OrderTime:      Time(openOrder.CreatedAt),
		Side:           side,
		Type:           1,
	}
}

type Order struct {
	OrderTime     Time   `json:"order_time"`
	ID            string `json:"id"`
	OrderID       uint64 `json:"order_id"`
	PoolID        uint64 `json:"pool_id"`
	PoolSymbol    string `json:"pool_symbol"`
	PoolBaseCoin  string `json:"pool_base_coin"`
	PoolQuoteCoin string `json:"pool_quote_coin"`
	ClientOrderID string `json:"order_cid"`
	Trader        string `json:"trader"`
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

func OrderFromHistoryDB(order ckhdb.HistoryOrder) Order {
	side := uint8(1)
	if order.IsBid {
		side = 0
	}
	orderType := uint8(1)
	if order.IsMarket {
		orderType = 0
	}

	return Order{
		OrderTime:      Time(order.CreatedAt),
		ID:             fmt.Sprintf("%d-%d-%d", order.PoolID, order.OrderID, side),
		OrderID:        order.OrderID,
		PoolID:         order.PoolID,
		PoolSymbol:     order.PoolSymbol,
		PoolBaseCoin:   order.PoolBaseCoin,
		PoolQuoteCoin:  order.PoolQuoteCoin,
		ClientOrderID:  order.ClientOrderID,
		Trader:         order.Trader,
		Side:           side,
		Type:           orderType,
		OrderPrice:     order.Price.String(),
		AvgPrice:       order.AvgPrice.String(),
		OrderAmount:    order.OriginalQuantity.String(),
		ExecutedAmount: order.ExecutedQuantity.String(),
		FilledTotal:    order.ExecutedQuantity.Mul(order.AvgPrice).String(),
		Status:         uint8(order.Status),
	}
}

func OrderFormOpenDB(order db.OpenOrder) Order {
	side := uint8(1)
	if order.IsBid {
		side = 0
	}
	return Order{
		OrderTime:      Time(order.CreatedAt),
		ID:             fmt.Sprintf("%d-%d-%d", order.PoolID, order.OrderID, side),
		OrderID:        order.OrderID,
		PoolID:         order.PoolID,
		PoolSymbol:     order.PoolSymbol,
		PoolBaseCoin:   order.PoolBaseCoin,
		PoolQuoteCoin:  order.PoolQuoteCoin,
		ClientOrderID:  order.ClientOrderID,
		Trader:         order.Trader,
		Side:           side,
		Type:           1,
		OrderPrice:     order.Price.String(),
		AvgPrice:       order.Price.String(),
		OrderAmount:    order.OriginalQuantity.String(),
		ExecutedAmount: order.ExecutedQuantity.String(),
		FilledTotal:    order.ExecutedQuantity.Mul(order.Price).String(),
		Status:         uint8(order.Status),
	}
}

type OrderDetail struct {
	Order
	Trades []TradeDetail `json:"trades"`
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
			TakerFee:      trade.TakerFee.Add(trade.TakerAppFee).String(),
			MakerFee:      trade.MakerFee.Add(trade.MakerAppFee).String(),
			Timestamp:     Time(trade.Time),
		})
	}
	return result
}
