package admin

import (
	"context"
	"exapp-go/internal/db/db"
	entity_admin "exapp-go/internal/entity/admin"
	"exapp-go/pkg/queryparams"
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

func (s *AdminServices) QueryUserPointsGrant(ctx context.Context, params *queryparams.QueryParams) ([]*entity_admin.RespUserPointsGrant, int64, error) {
	var grants []*db.UserPointsGrant

	total, err := s.repo.Query(ctx, &grants, params, "admin", "uid", "status")
	if err != nil {
		return nil, 0, err
	}
	var resp []*entity_admin.RespUserPointsGrant
	for _, grant := range grants {
		resp = append(resp, new(entity_admin.RespUserPointsGrant).Fill(grant))
	}
	return resp, total, nil
}
