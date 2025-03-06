package handler

import (
	"context"
	"exapp-go/internal/db/db"
	"exapp-go/internal/entity"
	"log"
	"sync"
	"time"
)

type DepthBuffer struct {
	repo      *db.Repo
	params    []db.UpdateDepthParams
	batchSize int
	mu        sync.Mutex
	publisher *NSQPublisher
}

func NewDepthBuffer(batchSize int, repo *db.Repo, publisher *NSQPublisher) *DepthBuffer {
	buffer := &DepthBuffer{
		params:    make([]db.UpdateDepthParams, 0, batchSize),
		repo:      repo,
		batchSize: batchSize,
		publisher: publisher,
	}
	go buffer.periodicFlush()
	return buffer
}

func (b *DepthBuffer) Add(updateParams db.UpdateDepthParams) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.params = append(b.params, updateParams)
	if len(b.params) >= b.batchSize {
		b.flush()
	}
}

func (b *DepthBuffer) flush() {
	if len(b.params) == 0 {
		return
	}

	params := make([]db.UpdateDepthParams, len(b.params))
	copy(params, b.params)
	b.params = b.params[:0]

	log.Printf("flush depth buffer, %d params", len(params))
	ctx := context.Background()
	changes, err := b.repo.UpdateDepthV2(ctx, params)
	if err != nil {
		log.Printf("update depth failed: %v", err)
		return
	}

	bids, asks := [][]string{}, [][]string{}
	depths := []entity.Depth{}
	for _, change := range changes {
		bid := make([]string, 0)
		ask := make([]string, 0)
		if change.IsBuy {
			bid = append(bid, change.Price.String())
			bid = append(bid, change.Amount.String())
		} else {
			ask = append(ask, change.Price.String())
			ask = append(ask, change.Amount.String())
		}
		bids = append(bids, bid)
		asks = append(asks, ask)
		depths = append(depths, entity.Depth{
			PoolID:    change.PoolID,
			Timestamp: uint64(time.Now().UnixMilli()),
			Bids:      bids,
			Asks:      asks,
			Precision: change.Precision,
		})
		bids, asks = [][]string{}, [][]string{}
	}

	for _, depth := range depths {
		go b.publisher.PublishDepthUpdate(depth)
	}

}

func (b *DepthBuffer) periodicFlush() {
	ticker := time.NewTicker(time.Millisecond * 50)
	defer ticker.Stop()

	for range ticker.C {
		b.mu.Lock()
		b.flush()
		b.mu.Unlock()
	}
}
