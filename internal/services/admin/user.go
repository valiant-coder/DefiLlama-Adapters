package admin

import (
	"context"
	"exapp-go/internal/db/db"
	entity_admin "exapp-go/internal/entity/admin"
	"exapp-go/pkg/queryparams"
)

func (s *AdminServices) QueryUsers(ctx context.Context, params *queryparams.QueryParams) ([]*entity_admin.RespUser, int64, error) {

	if username := params.Query.Values["username"]; username != nil {
		params.CustomQuery["username"] = []interface{}{username}
	}
	if uid := params.Query.Values["uid"]; uid != nil {
		params.CustomQuery["uid"] = []interface{}{uid}
	}
	if startTime := params.Query.Values["start_time"]; startTime != nil {
		params.CustomQuery["start_time"] = []interface{}{startTime}
	}
	if endTime := params.Query.Values["end_time"]; endTime != nil {
		params.CustomQuery["end_time"] = []interface{}{endTime}
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
