package ckhdb

import (
	"context"
	"time"

	"github.com/shopspring/decimal"
)


// Trade represents a trade record in the DEX
type Trade struct {
	TxID           string          `json:"tx_id"`
	PoolID         uint64          `json:"pool_id"`
	Taker          string          `json:"taker" `
	Maker          string          `json:"maker"`
	MakerOrderID   uint64          `json:"maker_order_id" `
	MakerOrderCID  string          `json:"maker_order_cid"`
	TakerOrderID   uint64          `json:"taker_order_id"`
	TakerOrderCID  string          `json:"taker_order_cid"`
	Price          decimal.Decimal `json:"price"`
	TakerIsBid     bool            `json:"taker_is_bid"`
	BaseQuantity   decimal.Decimal `json:"base_quantity" gorm:"type:Decimal(36,18)"`
	QuoteQuantity  decimal.Decimal `json:"quote_quantity" gorm:"type:Decimal(36,18)"`
	TakerFee       decimal.Decimal `json:"taker_fee" gorm:"type:Decimal(36,18)"`
	MakerFee       decimal.Decimal `json:"maker_fee" gorm:"type:Decimal(36,18)"`
	Timestamp      time.Time       `json:"time"`
	BlockNumber    uint64          `json:"block_number"`
	GlobalSequence uint64          `json:"global_sequence"`
	CreatedAt      time.Time       `json:"created_at"`
}

// TableName overrides the table name
func (Trade) TableName() string {
	return "trades"
}

func (r *ClickHouseRepo) InsertTrade(ctx context.Context, trade *Trade) error {
	return r.DB.WithContext(ctx).Create(trade).Error
}

func (r *ClickHouseRepo) GetLatestTrades(ctx context.Context, poolID uint64, limit int) ([]Trade, error) {
	trades := []Trade{}
	err := r.DB.WithContext(ctx).Where("pool_id = ?", poolID).Order("timestamp desc").Limit(limit).Find(&trades).Error
	return trades, err
}

func (r *ClickHouseRepo) GetTrades(ctx context.Context, orderID uint64) ([]Trade, error) {
	trades := []Trade{}
	err := r.DB.WithContext(ctx).Where("maker_order_id = ? OR taker_order_id = ?", orderID, orderID).Find(&trades).Error
	return trades, err
}

func (r *ClickHouseRepo) GetMaxBlockNumber(ctx context.Context) (uint64, error) {
	var blockNumber uint64
	err := r.DB.WithContext(ctx).Model(&Trade{}).Select("MAX(block_number)").Scan(&blockNumber).Error
	return blockNumber, err
}
