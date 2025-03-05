package ckhdb

import (
	"context"
	"log"
	"sync"
	"time"
)

type OrderBuffer struct {
	orders    []*HistoryOrder
	batchSize int
	mu        sync.Mutex
	ckhRepo   *ClickHouseRepo
}

func NewOrderBuffer(batchSize int, ckhRepo *ClickHouseRepo) *OrderBuffer {
	buffer := &OrderBuffer{
		orders:    make([]*HistoryOrder, 0, batchSize),
		batchSize: batchSize,
		ckhRepo:   ckhRepo,
	}
	go buffer.periodicFlush()
	return buffer
}

func (b *OrderBuffer) Add(order *HistoryOrder) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.orders = append(b.orders, order)
	if len(b.orders) >= b.batchSize {
		b.flush()
	}
}

func (b *OrderBuffer) flush() {
	if len(b.orders) == 0 {
		return
	}

	orders := make([]*HistoryOrder, len(b.orders))
	copy(orders, b.orders)
	b.orders = b.orders[:0]

	go func() {
		err := b.ckhRepo.BatchInsertOrders(context.Background(), orders)
		if err != nil {
			log.Printf("insert orders failed: %v", err)
		}
	}()
}

func (b *OrderBuffer) periodicFlush() {
	ticker := time.NewTicker(time.Millisecond * 500)
	defer ticker.Stop()

	for range ticker.C {
		b.mu.Lock()
		b.flush()
		b.mu.Unlock()
	}
}
