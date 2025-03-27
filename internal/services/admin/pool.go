package admin

import (
	"context"
	"errors"
	"exapp-go/internal/db/db"
	entity_admin "exapp-go/internal/entity/admin"
	"exapp-go/pkg/queryparams"
)

func (s *AdminServices) QueryPools(ctx context.Context, queryParams *queryparams.QueryParams) ([]*entity_admin.RespPool, int64, error) {
	var pools []*db.Pool

	total, err := s.repo.Query(ctx, &pools, queryParams, "base_symbol", "base_contract", "quote_symbol", "quote_contract")
	if err != nil {
		return nil, 0, err
	}

	var resp []*entity_admin.RespPool
	for _, pool := range pools {
		resp = append(resp, new(entity_admin.RespPool).Fill(pool))
	}

	return resp, total, nil
}

func (s *AdminServices) UpdatePool(ctx context.Context, req entity_admin.ReqUpsertPool, poolId uint64) (*entity_admin.RespPool, error) {
	pool, err := s.repo.GetPoolByID(ctx, poolId)
	if err != nil {
		return nil, err
	}
	if pool.Status != 0 {
		return nil, errors.New("pool status is not closed")
	}

	pool.BaseContract = req.BaseContract
	pool.BaseSymbol = req.BaseSymbol
	pool.QuoteContract = req.QuoteContract
	pool.QuoteSymbol = req.QuoteSymbol
	pool.Symbol = req.Symbol
	pool.Visible = req.Visible
	pool.Status = req.Status

	if err := s.repo.UpdatePool(ctx, pool); err != nil {
		return nil, err
	}

	return new(entity_admin.RespPool).Fill(pool), nil
}

func (s *AdminServices) CreatePool(ctx context.Context, req *entity_admin.ReqUpsertPool) (*entity_admin.RespPool, error) {

	pool := &db.Pool{
		PoolID:        req.PoolID,
		BaseContract:  req.BaseContract,
		BaseSymbol:    req.BaseSymbol,
		QuoteContract: req.QuoteContract,
		QuoteSymbol:   req.QuoteSymbol,
		Visible:       req.Visible,
		Status:        req.Status,
	}
	if err := s.repo.CreatePoolIfNotExist(ctx, pool); err != nil {
		return nil, err
	}
	return new(entity_admin.RespPool).Fill(pool), nil
}
