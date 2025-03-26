package admin

import (
	"context"
	"exapp-go/internal/db/db"
	entity_admin "exapp-go/internal/entity/admin"
	"exapp-go/pkg/queryparams"
)

func (s *AdminServices) QueryTokens(ctx context.Context, queryParams *queryparams.QueryParams) ([]*entity_admin.RespToken, int64, error) {
	var tokens []*db.Token

	total, err := s.repo.Query(ctx, &tokens, queryParams, "symbol", "chains")
	if err != nil {
		return nil, 0, err
	}

	var resp []*entity_admin.RespToken
	for _, token := range tokens {
		resp = append(resp, new(entity_admin.RespToken).Fill(token))
	}
	return resp, total, nil
}
