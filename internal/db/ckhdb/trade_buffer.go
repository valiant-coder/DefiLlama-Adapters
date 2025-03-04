package ckhdb

import (
	"context"
	"log"
	"sync"
	"time"
)

type TradeBuffer struct {
	trades    []*Trade
	batchSize int
	mu        sync.Mutex
	ckhRepo   *ClickHouseRepo
}

func NewTradeBuffer(batchSize int, ckhRepo *ClickHouseRepo) *TradeBuffer {
	buffer := &TradeBuffer{
		trades:    make([]*Trade, 0, batchSize),
		batchSize: batchSize,
		ckhRepo:   ckhRepo,
	}
	go buffer.periodicFlush()
	return buffer
}

func (b *TradeBuffer) Add(trade *Trade) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.trades = append(b.trades, trade)
	if len(b.trades) >= b.batchSize {
		b.flush()
	}
}

func (b *TradeBuffer) flush() {
	if len(b.trades) == 0 {
		return
	}

	trades := make([]*Trade, len(b.trades))
	copy(trades, b.trades)
	b.trades = b.trades[:0]

	go func() {
		ctx := context.Background()
		if err := b.ckhRepo.BatchInsertTrades(ctx, trades); err != nil {
			log.Printf("批量插入交易记录失败: %v", err)
		}
	}()
}

func (b *TradeBuffer) periodicFlush() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for range ticker.C {
		b.mu.Lock()
		b.flush()
		b.mu.Unlock()
	}
}
