package ckhdb

import (
	"context"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
)

// Trade represents a trade record in the DEX
type Trade struct {
	TxID           string          `gorm:"column:tx_id;type:varchar(255)"`
	PoolID         uint64          `gorm:"column:pool_id;type:bigint(20)"`
	BaseCoin       string          `gorm:"column:base_coin;type:varchar(255)"`
	QuoteCoin      string          `gorm:"column:quote_coin;type:varchar(255)"`
	Symbol         string          `gorm:"column:symbol;type:varchar(255)"`
	Taker          string          `gorm:"column:taker;type:varchar(255)"`
	Maker          string          `gorm:"column:maker;type:varchar(255)"`
	MakerApp       string          `gorm:"column:maker_app;type:varchar(255)"`
	TakerApp       string          `gorm:"column:taker_app;type:varchar(255)"`
	MakerOrderTag  string          `gorm:"column:maker_order_tag;type:varchar(255)"`
	MakerOrderID   uint64          `gorm:"column:maker_order_id;type:bigint(20)"`
	MakerOrderCID  string          `gorm:"column:maker_order_cid;type:varchar(255)"`
	TakerOrderID   uint64          `gorm:"column:taker_order_id;type:bigint(20)"`
	TakerOrderCID  string          `gorm:"column:taker_order_cid;type:varchar(255)"`
	TakerOrderTag  string          `gorm:"column:taker_order_tag;type:varchar(255)"`
	Price          decimal.Decimal `gorm:"type:Decimal(36,18)"`
	TakerIsBid     bool            `gorm:"column:taker_is_bid;type:tinyint(1)"`
	BaseQuantity   decimal.Decimal `gorm:"type:Decimal(36,18)"`
	QuoteQuantity  decimal.Decimal `gorm:"type:Decimal(36,18)"`
	TakerFee       decimal.Decimal `gorm:"type:Decimal(36,18)"`
	MakerFee       decimal.Decimal `gorm:"type:Decimal(36,18)"`
	TakerAppFee    decimal.Decimal `gorm:"type:Decimal(36,18)"`
	MakerAppFee    decimal.Decimal `gorm:"type:Decimal(36,18)"`
	Time           time.Time       `gorm:"column:time;type:datetime"`
	BlockNumber    uint64          `gorm:"column:block_number;type:bigint(20)"`
	GlobalSequence uint64          `gorm:"column:global_sequence;type:bigint(20)"`
	CreatedAt      time.Time       `gorm:"column:created_at;type:datetime"`
}

// TableName overrides the table name
func (Trade) TableName() string {
	return "trades"
}

func (r *ClickHouseRepo) InsertTradeIfNotExist(ctx context.Context, trade *Trade) error {
	var makerSide, takerSide uint8
	if trade.TakerIsBid {
		takerSide = 0
		makerSide = 1
	} else {
		takerSide = 1
		makerSide = 0
	}
	trade.MakerOrderTag = fmt.Sprintf("%d-%d-%d", trade.PoolID, trade.MakerOrderID, makerSide)
	trade.TakerOrderTag = fmt.Sprintf("%d-%d-%d", trade.PoolID, trade.TakerOrderID, takerSide)
	return r.DB.WithContext(ctx).Model(&Trade{}).Where("global_sequence = ?", trade.GlobalSequence).FirstOrCreate(trade).Error
}

func (r *ClickHouseRepo) GetLatestTrades(ctx context.Context, poolID uint64, limit int) ([]Trade, error) {
	trades := []Trade{}
	err := r.DB.WithContext(ctx).Where("pool_id = ?", poolID).Order("global_sequence desc").Limit(limit).Find(&trades).Error
	return trades, err
}

func (r *ClickHouseRepo) GetTrades(ctx context.Context, orderTag string) ([]Trade, error) {
	trades := []Trade{}
	err := r.DB.WithContext(ctx).Where("maker_order_tag = ? OR taker_order_tag = ?", orderTag, orderTag).Find(&trades).Error
	return trades, err
}

func (r *ClickHouseRepo) GetMaxBlockNumber(ctx context.Context) (uint64, error) {
	var blockNumber *uint64
	err := r.DB.WithContext(ctx).Model(&Trade{}).Select("COALESCE(MAX(block_number), 0)").Scan(&blockNumber).Error
	if err != nil {
		return 0, err
	}
	if blockNumber == nil {
		return 0, nil
	}
	return *blockNumber, nil
}
