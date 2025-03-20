package admin

import (
	"context"
	"exapp-go/internal/db/db"
	entity_admin "exapp-go/internal/entity/admin"
	"exapp-go/pkg/queryparams"
)

func (s *AdminServices) QueryAdminPermissions(ctx context.Context, queryParams *queryparams.QueryParams) ([]*entity_admin.AdminPermission, int64, error) {
	var permissions []*db.AdminPermission
	queryParams.Preload = []string{"Actions", "Group"}
	total, err := s.repo.Query(ctx, &permissions, queryParams)
	if err != nil {
		return nil, 0, err
	}
	var resp []*entity_admin.AdminPermission
	for _, permission := range permissions {
		resp = append(resp, new(entity_admin.AdminPermission).Fill(permission))
	}
	return resp, total, nil

}

func (s *AdminServices) GetAdminPermission(ctx context.Context, id string) (*entity_admin.AdminPermission, error) {
	permission, err := s.repo.GetAdminPermission(ctx, id)
	if err != nil {
		return nil, err
	}
	return new(entity_admin.AdminPermission).Fill(permission), nil
}

func (s *AdminServices) CreateAdminPermission(ctx context.Context, req *entity_admin.ReqUpsertAdminPermission) (*entity_admin.AdminPermission, error) {
	actions := make([]*db.AdminPermissionAction, 0)
	for _, action := range req.Actions {
		actions = append(actions, &db.AdminPermissionAction{
			Method:      action.Method,
			Uri:         action.Uri,
			Description: action.Description,
		})
	}
	adminPermission := &db.AdminPermission{
		DisplayName: req.DisplayName,
		Path:        req.Path,
		GroupID:     req.GroupID,
		Actions:     actions,
	}
	err := s.repo.Insert(ctx, adminPermission)
	if err != nil {
		return nil, err
	}
	return new(entity_admin.AdminPermission).Fill(adminPermission), err

}

func (s *AdminServices) UpdateAdminPermission(ctx context.Context, req *entity_admin.ReqUpsertAdminPermission, id string) (*entity_admin.AdminPermission, error) {
	adminPermission, err := s.repo.GetAdminPermission(ctx, id)
	if err != nil {
		return nil, err
	}
	err = s.repo.DeleteAdminPermissionActions(ctx, id)
	if err != nil {
		return nil, err
	}

	actions := make([]*db.AdminPermissionAction, 0)
	for _, action := range req.Actions {
		actions = append(actions, &db.AdminPermissionAction{
			Method:      action.Method,
			Uri:         action.Uri,
			Description: action.Description,
		})
	}
	adminPermission.Actions = actions
	adminPermission.DisplayName = req.DisplayName
	adminPermission.Path = req.Path
	err = s.repo.Update(ctx, adminPermission)
	if err != nil {
		return nil, err
	}
	return new(entity_admin.AdminPermission).Fill(adminPermission), err

}
