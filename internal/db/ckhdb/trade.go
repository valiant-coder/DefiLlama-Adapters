package ckhdb

import (
	"context"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
)

// Trade represents a trade record in the DEX
type Trade struct {
	PoolID        uint64          `json:"pool_id"`
	Taker         string          `json:"taker" `
	Maker         string          `json:"maker"`
	MakerOrderID  uint64          `json:"maker_order_id" `
	MakerOrderCID string          `json:"maker_order_cid"`
	TakerOrderID  uint64          `json:"taker_order_id"`
	TakerOrderCID string          `json:"taker_order_cid"`
	Price         uint64          `json:"price"`
	TakerIsBid    bool            `json:"taker_is_bid"`
	BaseQuantity  decimal.Decimal `json:"base_quantity" gorm:"type:Decimal(36,18)"`
	QuoteQuantity decimal.Decimal `json:"quote_quantity" gorm:"type:Decimal(36,18)"`
	TakerFee      decimal.Decimal `json:"taker_fee" gorm:"type:Decimal(36,18)"`
	MakerFee      decimal.Decimal `json:"maker_fee" gorm:"type:Decimal(36,18)"`
	Timestamp     time.Time       `json:"time"`
	BlockNumber   uint64          `json:"block_number"`
}

// TableName overrides the table name
func (Trade) TableName() string {
	return "trades"
}

type TradeStat struct {
	High        uint64          `json:"high"`
	Low         uint64          `json:"low"`
	Trades      uint64          `json:"trades"`
	Volume      decimal.Decimal `json:"volume" gorm:"type:Decimal(36,18)"`
	QuoteVolume decimal.Decimal `json:"quote_volume" gorm:"type:Decimal(36,18)"`
}

func (r *ClickHouseRepo) GetTradeStat(ctx context.Context, poolID uint64) (*TradeStat, error) {
	query := fmt.Sprintf("SELECT MAX(price) AS high, MIN(price) AS low, COUNT(*) AS trades, SUM(base_quantity) AS volume, SUM(quote_quantity) AS quote_volume FROM trades WHERE pool_id = %d AND timestamp >= NOW() - INTERVAL 24 HOUR", poolID)
	row := r.DB.WithContext(ctx).Raw(query).Row()
	var stat TradeStat
	err := row.Scan(&stat.High, &stat.Low, &stat.Trades, &stat.Volume, &stat.QuoteVolume)
	return &stat, err
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

