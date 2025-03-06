package db

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

type OpenOrderBuffer struct {
	insertOrders []*OpenOrder
	updateOrders []*OpenOrder
	deleteOrders []*OpenOrder
	cache        map[string]*OpenOrder // key: poolID-orderID-isBid
	batchSize    int
	mu           sync.RWMutex
	repo         *Repo
}

func NewOpenOrderBuffer(batchSize int, repo *Repo) *OpenOrderBuffer {
	buffer := &OpenOrderBuffer{
		insertOrders: make([]*OpenOrder, 0, batchSize),
		updateOrders: make([]*OpenOrder, 0, batchSize),
		deleteOrders: make([]*OpenOrder, 0, batchSize),
		cache:        make(map[string]*OpenOrder),
		batchSize:    batchSize,
		repo:         repo,
	}
	go buffer.periodicFlush()
	return buffer
}

func (b *OpenOrderBuffer) cleanExpiredOrders() {
	b.mu.Lock()
	defer b.mu.Unlock()

	now := time.Now()
	expiredTime := now.Add(-30 * time.Minute)

	for key, order := range b.cache {
		if order.CreatedAt.Before(expiredTime) {
			delete(b.cache, key)
		}
	}
}

func (b *OpenOrderBuffer) getCacheKey(poolID uint64, orderID uint64, isBid bool) string {
	bidStr := "0"
	if isBid {
		bidStr = "1"
	}
	return fmt.Sprintf("%d-%d-%s", poolID, orderID, bidStr)
}

func (b *OpenOrderBuffer) Add(order *OpenOrder) {
	b.mu.Lock()
	defer b.mu.Unlock()

	key := b.getCacheKey(order.PoolID, order.OrderID, order.IsBid)
	b.cache[key] = order
	b.insertOrders = append(b.insertOrders, order)

	if len(b.insertOrders) >= b.batchSize {
		b.flush()
	}
}

func (b *OpenOrderBuffer) Update(order *OpenOrder) {
	b.mu.Lock()
	defer b.mu.Unlock()

	key := b.getCacheKey(order.PoolID, order.OrderID, order.IsBid)
	b.cache[key] = order
	b.updateOrders = append(b.updateOrders, order)

	if len(b.updateOrders) >= b.batchSize {
		b.flush()
	}
}

func (b *OpenOrderBuffer) Get(poolID uint64, orderID uint64, isBid bool) (*OpenOrder, error) {
	b.mu.RLock()
	key := b.getCacheKey(poolID, orderID, isBid)
	if order, ok := b.cache[key]; ok {
		b.mu.RUnlock()
		return order, nil
	}
	b.mu.RUnlock()

	order, err := b.repo.GetOpenOrder(context.Background(), poolID, orderID, isBid)
	if err != nil {
		return nil, err
	}

	b.mu.Lock()
	b.cache[key] = order
	b.mu.Unlock()

	return order, nil
}

func (b *OpenOrderBuffer) Delete(poolID uint64, orderID uint64, isBid bool) {
	b.mu.Lock()
	defer b.mu.Unlock()

	key := b.getCacheKey(poolID, orderID, isBid)
	if _, ok := b.cache[key]; ok {
		delete(b.cache, key)
	}

	b.deleteOrders = append(b.deleteOrders, &OpenOrder{
		PoolID:  poolID,
		OrderID: orderID,
		IsBid:   isBid,
	})

	if len(b.deleteOrders) >= b.batchSize {
		b.flush()
	}
}

func (b *OpenOrderBuffer) flush() {
	if len(b.insertOrders) == 0 && len(b.updateOrders) == 0 && len(b.deleteOrders) == 0 {
		return
	}

	insertOrders := make([]*OpenOrder, len(b.insertOrders))
	copy(insertOrders, b.insertOrders)
	b.insertOrders = b.insertOrders[:0]

	updateOrders := make([]*OpenOrder, len(b.updateOrders))
	copy(updateOrders, b.updateOrders)
	b.updateOrders = b.updateOrders[:0]

	deleteOrders := make([]*OpenOrder, len(b.deleteOrders))
	copy(deleteOrders, b.deleteOrders)
	b.deleteOrders = b.deleteOrders[:0]

	go func() {
		ctx := context.Background()
		if len(insertOrders) > 0 {
			err := b.repo.BatchInsertOpenOrders(ctx, insertOrders)
			if err != nil {
				log.Printf("batch insert open orders failed: %v", err)
			}
		}

		if len(updateOrders) > 0 {
			err := b.repo.BatchUpdateOpenOrders(ctx, updateOrders)
			if err != nil {
				log.Printf("batch update open orders failed: %v", err)
			}
		}

		if len(deleteOrders) > 0 {
			err := b.repo.BatchDeleteOpenOrders(ctx, deleteOrders)
			if err != nil {
				log.Printf("batch delete open orders failed: %v", err)
			}
		}
	}()
}

func (b *OpenOrderBuffer) periodicFlush() {
	ticker := time.NewTicker(time.Millisecond * 200)
	cleanTicker := time.NewTicker(time.Minute) 
	defer ticker.Stop()
	defer cleanTicker.Stop()

	for {
		select {
		case <-ticker.C:
			b.mu.Lock()
			b.flush()
			b.mu.Unlock()
		case <-cleanTicker.C:
			b.cleanExpiredOrders()
		}
	}
}
