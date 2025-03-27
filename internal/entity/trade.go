package entity

import "exapp-go/internal/db/ckhdb"

// Trade represents a trade record in the DEX
type TradeDetail struct {
	PoolID          uint64 `json:"pool_id"`
	BaseCoin        string `json:"base_coin"`
	QuoteCoin       string `json:"quote_coin"`
	TxID            string `json:"tx_id"`
	Taker           string `json:"taker"`
	TakerPermission string `json:"taker_permission"`
	Maker           string `json:"maker"`
	MakerPermission string `json:"maker_permission"`
	MakerOrderID    uint64 `json:"maker_order_id"`
	MakerOrderCID   string `json:"maker_order_cid"`
	TakerOrderID    uint64 `json:"taker_order_id"`
	TakerOrderCID   string `json:"taker_order_cid"`
	Price           string `json:"price"`
	TakerIsBid      bool   `json:"taker_is_bid"`
	BaseQuantity    string `json:"base_quantity"`
	QuoteQuantity   string `json:"quote_quantity"`
	TakerFee        string `json:"taker_fee"`
	MakerFee        string `json:"maker_fee"`
	Timestamp       Time   `json:"timestamp"`
	GlobalSeq       uint64 `json:"global_seq"`
}

func DBTradeToTradeDetail(trade ckhdb.Trade) TradeDetail {
	return TradeDetail{
		PoolID:          trade.PoolID,
		BaseCoin:        trade.BaseCoin,
		QuoteCoin:       trade.QuoteCoin,
		TxID:            trade.TxID,
		Taker:           trade.Taker,
		TakerPermission: trade.TakerPermission,
		Maker:           trade.Maker,
		MakerPermission: trade.MakerPermission,
		MakerOrderID:    trade.MakerOrderID,
		MakerOrderCID:   trade.MakerOrderCID,
		TakerOrderID:    trade.TakerOrderID,
		TakerOrderCID:   trade.TakerOrderCID,
		Price:           trade.Price.String(),
		TakerIsBid:      trade.TakerIsBid,
		BaseQuantity:    trade.BaseQuantity.String(),
		QuoteQuantity:   trade.QuoteQuantity.String(),
		TakerFee:        trade.TakerFee.Add(trade.TakerAppFee).String(),
		MakerFee:        trade.MakerFee.Add(trade.MakerAppFee).String(),
		Timestamp:       Time(trade.Time),
		GlobalSeq:       trade.GlobalSequence,
	}
}

type TradeSide string

const (
	TradeSideBuy  TradeSide = "buy"
	TradeSideSell TradeSide = "sell"
)

type Trade struct {
	PoolID   uint64    `json:"pool_id"`
	Buyer    string    `json:"buyer"`
	Seller   string    `json:"seller"`
	Quantity string    `json:"quantity"`
	Price    string    `json:"price"`
	TradedAt Time      `json:"traded_at"`
	Side     TradeSide `json:"side"`
}

func DbTradeToTrade(dbTrade ckhdb.Trade) Trade {
	var buyer, seller string
	side := TradeSideBuy
	if dbTrade.TakerIsBid {
		buyer = dbTrade.Taker
		seller = dbTrade.Maker
	} else {
		buyer = dbTrade.Maker
		seller = dbTrade.Taker
		side = TradeSideSell
	}
	return Trade{
		PoolID:   dbTrade.PoolID,
		Buyer:    buyer,
		Seller:   seller,
		Quantity: dbTrade.BaseQuantity.String(),
		Price:    dbTrade.Price.String(),
		TradedAt: Time(dbTrade.Time),
		Side:     side,
	}
}
