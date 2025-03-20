package db

import (
	"context"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func init() {
	addMigrateFunc(func(r *Repo) error {
		return r.AutoMigrate(
			&Admin{},
			&AdminRole{},
			&AdminPermission{},
			&AdminPermissionAction{},
			&AdminPermissionGroup{})
	})
	addMigrateFunc(func(r *Repo) error {
		superAdminRole := &AdminRole{
			Name:        "admin",
			DisplayName: "super admin",
		}
		err := r.UpsertAdminRole(context.Background(), superAdminRole)
		if err != nil {
			return err
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)
		if err != nil {
			return err
		}

		admin := &Admin{
			Name:     "admin",
			Password: string(hashedPassword),
			RealName: "super admin",
			Locked:   false,
			Roles:    []*AdminRole{superAdminRole},
		}
		return r.UpsertAdmin(context.Background(), admin)
	})
}

type Admin struct {
	gorm.Model
	Name             string       `gorm:"type:varchar(30);uniqueIndex:idx_username"`
	Avatar           string       `gorm:"type:varchar(255)"`
	Password         string       `gorm:"type:varchar(255)"`
	RealName         string       `gorm:"type:varchar(255)"`
	Locked           bool         `gorm:"default:false"`
	GoogleAuthSecret string       `gorm:"type:varchar(255)"`
	FirstLogin       bool         `gorm:"default:true"`
	Roles            []*AdminRole `gorm:"many2many:admin_to_role;"`
	LastLoginAt      *time.Time   `gorm:"type:datetime"`
	Creator          string       `gorm:"type:varchar(30)"`
}

func (Admin) TableName() string {
	return "admins"
}

func (r *Repo) SaveAdmin(ctx context.Context, id uint, attributes map[string]interface{}) error {
	return r.DB.WithContext(ctx).Model(&Admin{}).Where("id = ?", id).Updates(attributes).Error
}

func (r *Repo) UpsertAdmin(ctx context.Context, admin *Admin) error {
	return r.DB.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "name"}},
		DoNothing: true,
	}).Create(admin).Error
}

type AdminRole struct {
	gorm.Model
	Name        string             `gorm:"type:varchar(30);uniqueIndex:idx_name"`
	DisplayName string             `gorm:"type:varchar(50)"`
	Permissions []*AdminPermission `gorm:"many2many:role_to_permission;"`
}

func (AdminRole) TableName() string {
	return "admin_roles"
}

type AdminPermission struct {
	gorm.Model
	GroupID     uint                     `gorm:"not null;"`
	Group       *AdminPermissionGroup    `gorm:"foreignKey:GroupID;"`
	Path        string                   `gorm:"type:varchar(50);not null;uniqueIndex:idx_path"`
	DisplayName string                   `gorm:"type:varchar(50);not null"`
	Actions     []*AdminPermissionAction `gorm:"foreignKey:PermissionID;"`
}

func (AdminPermission) TableName() string {
	return "admin_permissions"
}

type AdminPermissionGroup struct {
	gorm.Model
	Path        string             `gorm:"type:varchar(50);uniqueIndex:idx_path"`
	DisplayName string             `gorm:"type:varchar(50)"`
	Permissions []*AdminPermission `gorm:"foreignKey:GroupID;"`
}

func (AdminPermissionGroup) TableName() string {
	return "admin_permission_groups"
}

type AdminPermissionAction struct {
	gorm.Model
	PermissionID uint64           `gorm:"index"`
	Permission   *AdminPermission `gorm:"foreignKey:PermissionID;"`
	Method       string           `gorm:"type:varchar(10);not null;uniqueIndex:idx_method_uri"`
	Uri          string           `gorm:"type:varchar(100);not null;uniqueIndex:idx_method_uri"`
	Description  string           `gorm:"type:varchar(200)"`
}

func (AdminPermissionAction) TableName() string {
	return "admin_permission_actions"
}

func (r *Repo) GetAllAdmins(ctx context.Context) ([]*Admin, error) {
	var admins []*Admin
	if err := r.DB.WithContext(ctx).Preload("Roles").Find(&admins).Error; err != nil {
		return nil, err
	}
	return admins, nil
}

func (r *Repo) GetAdminsByRoleID(ctx context.Context, roleID uint64) ([]*Admin, error) {
	var admins []*Admin
	if err := r.DB.WithContext(ctx).Preload("Roles").Where("role_id = ?", roleID).Find(&admins).Error; err != nil {
		return nil, err
	}
	return admins, nil
}

func (r *Repo) GetAdmin(ctx context.Context, id string) (*Admin, error) {
	var admin *Admin
	if err := r.DB.WithContext(ctx).Preload("Roles.Permissions.Actions").Preload("Roles.Permissions.Group").Where("id = ?", id).First(&admin).Error; err != nil {
		return nil, err
	}
	return admin, nil
}

func (r *Repo) ClearAdminRoles(ctx context.Context, adminID uint) error {
	return r.ExecSQL(ctx, "delete from admin_to_role where admin_id = ?", adminID)
}

func (r *Repo) GetAdminByName(ctx context.Context, name string) (*Admin, error) {
	var admin *Admin
	if err := r.DB.WithContext(ctx).Preload("Roles.Permissions.Actions").Preload("Roles.Permissions.Group").Where("name = ?", name).First(&admin).Error; err != nil {
		return nil, err
	}
	return admin, nil
}

func (r *Repo) IsAdminExists(ctx context.Context, name string) (bool, error) {
	var count int64
	err := r.DB.WithContext(ctx).Model(&Admin{}).Where("name = ?", name).Count(&count).Error
	return count > 0, err
}

func (r *Repo) DeleteAdmin(ctx context.Context, id uint) error {
	return r.DB.Select(clause.Associations).Delete(&Admin{Model: gorm.Model{ID: id}}).Error
}

func (r *Repo) GetAllAdminRoles(ctx context.Context) ([]*AdminRole, error) {

	var roles []*AdminRole
	if err := r.DB.WithContext(ctx).Preload("Permissions.Actions").Find(&roles).Error; err != nil {
		return nil, err
	}
	return roles, nil
}

func (r *Repo) GetAdminRole(ctx context.Context, id string) (*AdminRole, error) {
	var role *AdminRole
	if err := r.DB.WithContext(ctx).Preload("Permissions.Group").Preload("Permissions.Actions").Where("id = ?", id).First(&role).Error; err != nil {
		return nil, err
	}
	return role, nil
}

func (r *Repo) UpsertAdminRole(ctx context.Context, role *AdminRole) error {
	return r.DB.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "name"}},
		DoNothing: true,
	}).Create(role).Error
}

func (r *Repo) GetAdminRoles(ctx context.Context, ids []string) ([]*AdminRole, error) {
	var roles []*AdminRole
	if err := r.DB.WithContext(ctx).Preload("Permissions").Where("id in ?", ids).Find(&roles).Error; err != nil {
		return nil, err
	}
	return roles, nil
}

func (r *Repo) GetAdminPermissionsByIDs(ctx context.Context, permissionIDs []uint64) ([]*AdminPermission, error) {
	var permissions []*AdminPermission
	if err := r.DB.WithContext(ctx).Preload("Group").Where("id in ?", permissionIDs).Find(&permissions).Error; err != nil {
		return nil, err
	}
	return permissions, nil
}

func (r *Repo) GetAdminPermission(ctx context.Context, id string) (*AdminPermission, error) {
	var permission *AdminPermission
	if err := r.DB.WithContext(ctx).Preload("Actions").Preload("Group").Where("id = ?", id).First(&permission).Error; err != nil {
		return nil, err
	}
	return permission, nil
}

func (r *Repo) DeleteAdminPermissionActions(ctx context.Context, permissionID string) error {
	return r.WithContext(ctx).Unscoped().Where("permission_id = ?", permissionID).Delete(&AdminPermissionAction{}).Error
}

func (r *Repo) GetAdminPermissionGroup(ctx context.Context, id string) (*AdminPermissionGroup, error) {
	var group *AdminPermissionGroup
	if err := r.DB.WithContext(ctx).Preload("Permissions").Where("id = ?", id).First(&group).Error; err != nil {
		return nil, err
	}
	return group, nil
}
