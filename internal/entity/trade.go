package entity

import "exapp-go/internal/db/ckhdb"

// Trade represents a trade record in the DEX
type TradeDetail struct {
	ID            uint64 `json:"id"`
	PoolID        uint64 `json:"pool_id"`
	TxHash        string `json:"tx_hash"`
	Taker         string `json:"taker" `
	Maker         string `json:"maker"`
	MakerOrderID  uint64 `json:"maker_order_id" `
	MakerOrderCID string `json:"maker_order_cid"`
	TakerOrderID  uint64 `json:"taker_order_id"`
	TakerOrderCID string `json:"taker_order_cid"`
	Price         uint64 `json:"price"`
	TakerIsBid    bool   `json:"taker_is_bid"`
	BaseQuantity  string `json:"base_quantity"`
	QuoteQuantity string `json:"quote_quantity"`
	TakerFee      string `json:"taker_fee"`
	MakerFee      string `json:"maker_fee"`
	Timestamp     Time   `json:"time"`
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
