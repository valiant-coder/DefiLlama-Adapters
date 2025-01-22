package handler

import (
	"context"
	"exapp-go/internal/db/db"
	"exapp-go/internal/entity"
	"log"
	"time"
)

func (s *Service) updateDepth(ctx context.Context, params db.UpdateDepthParams) error {
	changes, err := s.repo.UpdateDepthV2(ctx, []db.UpdateDepthParams{params})
	if err != nil {
		log.Printf("update depth failed: %v", err)
		return err
	}

	bids, asks := [][]string{}, [][]string{}
	for _, change := range changes {
		var bid, ask []string
		if change.IsBuy {
			bid = append(bid, change.Price.String())
			bid = append(bid, change.Amount.String())
		} else {
			ask = append(ask, change.Price.String())
			ask = append(ask, change.Amount.String())
		}
		bids = append(bids, bid)
		asks = append(asks, ask)
	}
	depth := entity.Depth{
		PoolID:    params.PoolID,
		Timestamp: entity.Time(time.Now()),
		Bids:      bids,
		Asks:      asks,
	}

	go s.publisher.PublishDepthUpdate(depth)

	return nil
}
