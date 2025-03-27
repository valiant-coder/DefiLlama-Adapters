package admin

import (
	"context"
	"exapp-go/internal/db/db"
	entity_admin "exapp-go/internal/entity/admin"
	"exapp-go/pkg/queryparams"
)

func (s *AdminServices) QueryTokens(ctx context.Context, queryParams *queryparams.QueryParams) ([]*entity_admin.RespToken, int64, error) {
	var tokens []*db.Token

	total, err := s.repo.Query(ctx, &tokens, queryParams, "symbol", "evm_contract_address", "chains")
	if err != nil {
		return nil, 0, err
	}

	var resp []*entity_admin.RespToken
	for _, token := range tokens {
		resp = append(resp, entity_admin.TokenFromDB(token))
	}
	return resp, total, nil
}

func (s *AdminServices) CreateToken(ctx context.Context, req *entity_admin.ReqUpsertToken) (*entity_admin.RespToken, error) {
	token := entity_admin.DBFromToken(req)
	err := s.repo.InsertToken(ctx, token)
	return entity_admin.TokenFromDB(token), err
}
