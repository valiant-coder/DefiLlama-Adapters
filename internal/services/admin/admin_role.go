package admin

import (
	"context"
	"exapp-go/internal/db/db"
	entity_admin "exapp-go/internal/entity/admin"
	"exapp-go/pkg/queryparams"
)

func (s *AdminServices) QueryAdminRoles(ctx context.Context, queryParams *queryparams.QueryParams) ([]*entity_admin.AdminRole, int64, error) {
	var adminRoles []*db.AdminRole

	total, err := s.repo.Query(ctx, &adminRoles, queryParams)
	if err != nil {
		return nil, 0, err
	}
	var resp []*entity_admin.AdminRole
	for _, admin := range adminRoles {
		resp = append(resp, new(entity_admin.AdminRole).Fill(admin))
	}
	return resp, total, nil

}

func (s *AdminServices) GetAdminRole(ctx context.Context, id string) (*entity_admin.AdminRole, error) {
	adminRole, err := s.repo.GetAdminRole(ctx, id)
	if err != nil {
		return nil, err
	}
	return new(entity_admin.AdminRole).Fill(adminRole), nil
}

func (s *AdminServices) CreateAdminRole(ctx context.Context, req *entity_admin.ReqUpsertAdminRole) (*entity_admin.AdminRole, error) {
	permissions, err := s.repo.GetAdminPermissionsByIDs(ctx, req.PermissionIDs)
	if err != nil {
		return nil, err
	}
	adminRole := &db.AdminRole{
		Name:        req.Name,
		DisplayName: req.DisplayName,
		Permissions: permissions,
	}
	err = s.repo.Insert(ctx, adminRole)
	if err != nil {
		return nil, err
	}
	return new(entity_admin.AdminRole).Fill(adminRole), err
}

func (s *AdminServices) UpdateAdminRole(ctx context.Context, req *entity_admin.ReqUpsertAdminRole, roleID string) (*entity_admin.AdminRole, error) {
	adminRole, err := s.repo.GetAdminRole(ctx, roleID)
	if err != nil {
		return nil, err
	}

	permissions, err := s.repo.GetAdminPermissionsByIDs(ctx, req.PermissionIDs)
	if err != nil {
		return nil, err
	}

	adminRole.Name = req.Name
	adminRole.DisplayName = req.DisplayName
	adminRole.Permissions = permissions
	err = s.repo.Update(ctx, adminRole)
	if err != nil {
		return nil, err
	}
	return new(entity_admin.AdminRole).Fill(adminRole), err
}
