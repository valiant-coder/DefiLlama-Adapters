package admin

import (
	"context"
	"exapp-go/internal/db/db"
	entity_admin "exapp-go/internal/entity/admin"
)

func (s *AdminServices) CreateUserPointsGrant(ctx context.Context, req *entity_admin.ReqUserPointsGrant, operator string) (*entity_admin.RespUserPointsGrant, error) {
	grant := &db.UserPointsGrant{
		Admin:  operator,
		UID:    req.UID,
		Amount: req.Amount,
	}
	err := s.repo.Insert(ctx, grant)
	return new(entity_admin.RespUserPointsGrant).Fill(grant), err
}
