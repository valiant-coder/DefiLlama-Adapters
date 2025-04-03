package admin

import (
	"context"
	"exapp-go/internal/db/db"
	entity_admin "exapp-go/internal/entity/admin"
	"exapp-go/pkg/queryparams"
	"fmt"
	"time"

	"github.com/spf13/cast"
)

func (s *AdminServices) QueryUsers(ctx context.Context, params *queryparams.QueryParams) ([]*entity_admin.RespUser, int64, error) {

	if username := params.Query.Values["username"]; username != nil {
		params.CustomQuery["username"] = []interface{}{username}
	}
	if uid := params.Query.Values["uid"]; uid != nil {
		params.CustomQuery["uid"] = []interface{}{uid}
	}
	if startTime := params.Query.Values["start_time"]; startTime != nil {
		start := time.Unix(cast.ToInt64(startTime), 0)
		params.CustomQuery["start_time"] = []interface{}{start}
	}
	if endTime := params.Query.Values["end_time"]; endTime != nil {
		end := time.Unix(cast.ToInt64(endTime), 0)
		params.CustomQuery["end_time"] = []interface{}{end}
	}

	users, total, err := s.repo.QueryUserList(ctx, params)
	if err != nil {
		return nil, 0, err
	}

	var resp []*entity_admin.RespUser
	for _, user := range users {
		resp = append(resp, new(entity_admin.RespUser).Fill(user))
	}

	return resp, total, nil
}

func (s *AdminServices) GetPasskeys(ctx context.Context, queryParams *queryparams.QueryParams) ([]*entity_admin.RespPasskey, int64, error) {
	var passkeys []*db.UserCredential

	total, err := s.repo.Query(ctx, &passkeys, queryParams, "uid")
	if err != nil {
		return nil, 0, err
	}
	var resp []*entity_admin.RespPasskey
	for _, passkey := range passkeys {
		resp = append(resp, new(entity_admin.RespPasskey).Fill(passkey))
	}

	return resp, total, nil
}

func (s *AdminServices) GetUsersStatis(ctx context.Context, timeDimension, dataType string, amount int) ([]*db.UsersStatis, int64, error) {
	if amount == 0 {
		amount = 10
	}

	switch dataType {
	case entity_admin.DataTypeAddUserCount:
		return s.repo.GetStatisAddUserCount(ctx, timeDimension, amount)

	case entity_admin.DataTypeAddPasskeyCount:
		return s.repo.GetStatisAddPasskeyCount(ctx, timeDimension, amount)

	case entity_admin.DataTypeAddEvmCount:
		return s.repo.GetStatisAddEvmCount(ctx, timeDimension, amount)

	case entity_admin.DateTypeAddDepositCount:
		return s.repo.GetStatisAddDepositCount(ctx, timeDimension, amount)

	default:
		return nil, 0, fmt.Errorf("data_type is invalid")
	}
}
