package admin

import (
	"context"
	"exapp-go/internal/db/db"
	entity_admin "exapp-go/internal/entity/admin"
	"exapp-go/pkg/queryparams"
)

func (s *AdminServices) QueryAdminPermissionGroups(ctx context.Context, queryParams *queryparams.QueryParams) ([]*entity_admin.AdminPermissionGroup, int64, error) {
	var groups []*db.AdminPermissionGroup
	queryParams.Preload = []string{"Permissions"}
	total, err := s.repo.Query(ctx, &groups, queryParams)
	if err != nil {
		return nil, 0, err
	}
	var resp []*entity_admin.AdminPermissionGroup
	for _, group := range groups {
		resp = append(resp, new(entity_admin.AdminPermissionGroup).Fill(group))
	}
	return resp, total, nil

}

func (s *AdminServices) GetAdminPermissionGroup(ctx context.Context, id string) (*entity_admin.AdminPermissionGroup, error) {
	group, err := s.repo.GetAdminPermissionGroup(ctx, id)
	if err != nil {
		return nil, err
	}
	return new(entity_admin.AdminPermissionGroup).Fill(group), nil
}

func (s *AdminServices) CreateAdminPermissionGroup(ctx context.Context, req *entity_admin.ReqUpsertAdminPermissionGroup) (*entity_admin.AdminPermissionGroup, error) {
	adminPermissionGroup := &db.AdminPermissionGroup{
		DisplayName: req.DisplayName,
		Path:        req.Path,
	}
	err := s.repo.Insert(ctx, adminPermissionGroup)
	if err != nil {
		return nil, err
	}
	return new(entity_admin.AdminPermissionGroup).Fill(adminPermissionGroup), err

}

func (s *AdminServices) UpdateAdminPermissionGroup(ctx context.Context, req *entity_admin.ReqUpsertAdminPermissionGroup, groupID string) (*entity_admin.AdminPermissionGroup, error) {
	adminPermissionGroup, err := s.repo.GetAdminPermissionGroup(ctx, groupID)
	if err != nil {
		return nil, err
	}
	adminPermissionGroup.DisplayName = req.DisplayName
	adminPermissionGroup.Path = req.Path
	err = s.repo.Update(ctx, adminPermissionGroup)
	if err != nil {
		return nil, err
	}
	return new(entity_admin.AdminPermissionGroup).Fill(adminPermissionGroup), err

}
