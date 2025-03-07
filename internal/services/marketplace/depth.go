package marketplace

import (
	"context"
	"exapp-go/internal/db/db"
	"exapp-go/internal/entity"
	"time"
)

func NewDepthService() *DepthService {
	return &DepthService{repo: db.New()}
}

type DepthService struct {
	repo *db.Repo
}

func (s *DepthService) GetDepth(ctx context.Context, poolID uint64, precision string, limit int) (entity.Depth, error) {
	if precision == "" {
		precision = "0.000000001"
	}
	if precision == "0.00000001" && limit == 1 {
		precision = "0.000000001"
	}
	if limit == 0 {
		limit = 100
	}
	depth, err := s.repo.GetDepthV2(ctx, poolID, precision, limit)
	if err != nil {
		return entity.Depth{
			PoolID:    poolID,
			Timestamp: uint64(time.Now().UnixMilli()),
			Bids:      make([][]string, 0),
			Asks:      make([][]string, 0),
		}, err
	}
	if depth.Bids == nil {
		depth.Bids = make([][]string, 0)
	}
	if depth.Asks == nil {
		depth.Asks = make([][]string, 0)
	}
	return entity.Depth{
		PoolID:    depth.PoolID,
		Timestamp: uint64(time.Now().UnixMilli()),
		Bids:      depth.Bids,
		Asks:      depth.Asks,
		Precision: precision,
	}, nil
}
