package admin

import (
	"context"
	"errors"
	"exapp-go/pkg/queryparams"

	"exapp-go/internal/db/db"
	entity_admin "exapp-go/internal/entity/admin"
	"exapp-go/internal/errno"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func (s *AdminServices) QueryAdmins(ctx context.Context, queryParams *queryparams.QueryParams) ([]*entity_admin.RespAdmin, int64, error) {
	var admins []*db.Admin
	queryParams.Preload = []string{"Roles"}

	total, err := s.repo.Query(ctx, &admins, queryParams)
	if err != nil {
		return nil, 0, err
	}
	var resp []*entity_admin.RespAdmin
	for _, admin := range admins {
		resp = append(resp, new(entity_admin.RespAdmin).Fill(admin))
	}
	return resp, total, nil

}

func (s *AdminServices) GetAdmin(ctx context.Context, name string) (*entity_admin.RespAdmin, error) {
	admin, err := s.repo.GetAdminByName(ctx, name)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errno.DefaultParamsError("Admin does not exist")
		}
		return nil, err
	}
	return new(entity_admin.RespAdmin).Fill(admin), nil
}

func (s *AdminServices) UpdateAdmin(ctx context.Context, req *entity_admin.ReqUpsertAdmin, name, operator string) (*entity_admin.RespAdmin, error) {
	if req.Name == "admin" && operator != "admin" {
		err := errno.DefaultParamsError("Cannot modify super admin information")
		return nil, err
	}
	oldAdmin, err := s.repo.GetAdminByName(ctx, name)
	if err != nil {
		return nil, err
	}
	err = s.repo.ClearAdminRoles(ctx, oldAdmin.ID)
	if err != nil {
		return nil, err
	}
	roles, err := s.repo.GetAdminRoles(ctx, req.RoleIDs)
	if err != nil {
		return nil, err
	}
	oldAdmin.Name = req.Name
	oldAdmin.RealName = req.RealName
	oldAdmin.Avatar = req.Avatar

	if req.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}
		oldAdmin.Password = string(hashedPassword)
	}

	oldAdmin.Roles = roles
	err = s.repo.Update(ctx, oldAdmin)

	return new(entity_admin.RespAdmin).Fill(oldAdmin), err

}

func (s *AdminServices) CreateAdmin(ctx context.Context, req *entity_admin.ReqUpsertAdmin, operator string) (*entity_admin.RespAdmin, error) {
	roles, err := s.repo.GetAdminRoles(ctx, req.RoleIDs)
	if err != nil {
		return nil, err
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	admin := &db.Admin{
		Name:     req.Name,
		Password: string(hashedPassword),
		RealName: req.RealName,
		Avatar:   req.Avatar,
		Roles:    roles,
		Creator:  operator,
	}
	err = s.repo.Insert(ctx, admin)
	return new(entity_admin.RespAdmin).Fill(admin), err
}

func (s *AdminServices) DeleteAdmin(ctx context.Context, name string) error {
	admin, err := s.repo.GetAdminByName(ctx, name)
	if err != nil {
		return err
	}
	if admin.Name == "admin" {
		return errno.DefaultParamsError("Cannot delete super admin")
	}
	return s.repo.DeleteAdmin(ctx, admin.ID)
}
