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

func (s *DepthService) GetDepth(ctx context.Context, poolID uint64) (entity.Depth, error) {
	depth, err := s.repo.GetDepth(ctx, poolID)
	if err != nil {
		return entity.Depth{}, err
	}
	return entity.Depth{
		PoolID:    depth.PoolID,
		Timestamp: entity.Time(time.Now()),
		Bids:      depth.Bids,
		Asks:      depth.Asks,
	}, nil
}

