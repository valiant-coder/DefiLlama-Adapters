package admin

import (
	"context"
	"exapp-go/internal/db/db"
	entity_admin "exapp-go/internal/entity/admin"
	"exapp-go/pkg/queryparams"
)

func (s *AdminServices) GetTransactionsRecord(ctx context.Context, params *queryparams.QueryParams) ([]*db.TransactionsRecord, int64, error) {

	if symbol := params.Query.Values["symbol"]; symbol != nil {
		params.CustomQuery["symbol"] = []interface{}{symbol}
	}
	if chainName := params.Query.Values["chain_name"]; chainName != nil {
		params.CustomQuery["chain_name"] = []interface{}{chainName}
	}
	if uid := params.Query.Values["uid"]; uid != nil {
		params.CustomQuery["uid"] = []interface{}{uid}
	}
	if txHash := params.Query.Values["tx_hash"]; txHash != nil {
		params.CustomQuery["tx_hash"] = []interface{}{txHash}
	}
	if startTime := params.Query.Values["start_time"]; startTime != nil {
		params.CustomQuery["start_time"] = []interface{}{startTime}
	}
	if endTime := params.Query.Values["end_time"]; endTime != nil {
		params.CustomQuery["end_time"] = []interface{}{endTime}
	}

	record, total, err := s.repo.QueryTransactionsRecord(ctx, params)
	if err != nil {
		return nil, 0, err
	}

	return record, total, nil
}

func (s *AdminServices) GetDepositAmountTotal(ctx context.Context, startTime, endTime string) ([]*entity_admin.RespGetDepositWithdrawal, error) {

	records, err := s.repo.GetDepositAmountTotal(ctx, startTime, endTime)
	if err != nil {
		return nil, err
	}

	var resp []*entity_admin.RespGetDepositWithdrawal
	for _, record := range records {
		resp = append(resp, new(entity_admin.RespGetDepositWithdrawal).Fill(record))
	}

	return resp, nil
}
