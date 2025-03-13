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
	Type               uint8  `json:"type"`
	OrderPrice         string `json:"order_price"`
	AvgPrice           string `json:"avg_price"`
	OrderAmount        string `json:"order_amount"`
	ExecutedAmount     string `json:"executed_amount"`
	OrderTotal         string `json:"order_total"`
	BaseCoinPrecision  uint8  `json:"base_coin_precision"`
	QuoteCoinPrecision uint8  `json:"quote_coin_precision"`
}

func OpenOrderFromDB(openOrder db.OpenOrder) OpenOrder {
	side := uint8(1)
	if openOrder.IsBid {
		side = 0
	}
	return OpenOrder{
		ID:                 fmt.Sprintf("%d-%d-%d", openOrder.PoolID, openOrder.OrderID, side),
		OrderID:            openOrder.OrderID,
		PoolID:             openOrder.PoolID,
		PoolSymbol:         openOrder.PoolSymbol,
		PoolBaseCoin:       openOrder.PoolBaseCoin,
		PoolQuoteCoin:      openOrder.PoolQuoteCoin,
		ClientOrderID:      openOrder.ClientOrderID,
		Trader:             openOrder.Trader,
		OrderPrice:         openOrder.Price.String(),
		AvgPrice:           openOrder.Price.String(),
		OrderAmount:        openOrder.OriginalQuantity.String(),
		ExecutedAmount:     openOrder.ExecutedQuantity.String(),
		OrderTotal:         openOrder.OriginalQuantity.Mul(openOrder.Price).Round(int32(openOrder.QuoteCoinPrecision)).String(),
		OrderTime:          Time(openOrder.CreatedAt),
		Side:               side,
		Type:               1,
		BaseCoinPrecision:  openOrder.BaseCoinPrecision,
		QuoteCoinPrecision: openOrder.QuoteCoinPrecision,
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
	OrderTotal     string `json:"order_total"`
	FilledTotal    string `json:"filled_total"`
	// 0 open 1partially_filled 2full_filled 3.canceled
	Status             uint8 `json:"status"`
	History            bool  `json:"history"`
	BaseCoinPrecision  uint8 `json:"base_coin_precision"`
	QuoteCoinPrecision uint8 `json:"quote_coin_precision"`
	Unread             bool  `json:"unread"`
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
		OrderTime:          Time(order.CreatedAt),
		ID:                 fmt.Sprintf("%d-%d-%d", order.PoolID, order.OrderID, side),
		OrderID:            order.OrderID,
		PoolID:             order.PoolID,
		PoolSymbol:         order.PoolSymbol,
		PoolBaseCoin:       order.PoolBaseCoin,
		PoolQuoteCoin:      order.PoolQuoteCoin,
		ClientOrderID:      order.ClientOrderID,
		Trader:             order.Trader,
		Side:               side,
		Type:               orderType,
		OrderPrice:         order.Price.String(),
		AvgPrice:           order.AvgPrice.String(),
		OrderAmount:        order.OriginalQuantity.String(),
		ExecutedAmount:     order.ExecutedQuantity.String(),
		OrderTotal:         order.OriginalQuantity.Mul(order.AvgPrice).Round(int32(order.QuoteCoinPrecision)).String(),
		FilledTotal:        order.ExecutedQuantity.Mul(order.AvgPrice).Truncate(int32(order.QuoteCoinPrecision)).String(),
		Status:             uint8(order.Status),
		BaseCoinPrecision:  order.BaseCoinPrecision,
		QuoteCoinPrecision: order.QuoteCoinPrecision,
		History:            true,
		Unread:             false,
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
		OrderTotal:     order.OriginalQuantity.Mul(order.Price).Round(int32(order.QuoteCoinPrecision)).String(),
		FilledTotal:    order.ExecutedQuantity.Mul(order.Price).Truncate(int32(order.QuoteCoinPrecision)).String(),
		Status:         uint8(order.Status),
		History:        false,
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

type ReqMakeOrderAsRead struct {
	Trader string `json:"trader"`
	ID     string `json:"id"`
}

type RespUnreadOrder struct {
	HasUnread bool `json:"has_unread"`
}
