package entity_admin

import (
	"exapp-go/internal/db/db"
	"exapp-go/internal/entity"
	"time"
)

type RespAdmin struct {
	ID          uint         `json:"id"`
	Name        string       `json:"name"`
	Avatar      string       `json:"avatar"`
	RealName    string       `json:"real_name"`
	Roles       []*AdminRole `json:"roles"`
	CreatedAt   entity.Time  `json:"created_at"`
	Creator     string       `json:"creator"`
	LastLoginAt entity.Time  `json:"last_login"`
}

func (r *RespAdmin) Fill(a *db.Admin) *RespAdmin {
	r.ID = a.ID
	r.Name = a.Name
	r.Avatar = a.Avatar
	r.RealName = a.RealName
	r.Roles = make([]*AdminRole, 0)
	for _, role := range a.Roles {
		r.Roles = append(r.Roles, new(AdminRole).Fill(role))
	}
	r.CreatedAt = entity.Time(a.CreatedAt)
	r.Creator = a.Creator
	if a.LastLoginAt == nil {
		r.LastLoginAt = entity.Time(time.Unix(0, 0))
	} else {
		r.LastLoginAt = entity.Time(*a.LastLoginAt)
	}

	return r

}

type ReqUpsertAdmin struct {
	Name     string   `json:"name"`
	Avatar   string   `json:"avatar"`
	RealName string   `json:"real_name"`
	Password string   `json:"password"`
	RoleIDs  []string `json:"role_ids"`
}

type AdminRole struct {
	ID          uint               `json:"id"`
	DisplayName string             `json:"display_name"`
	Permissions []*AdminPermission `json:"permissions"`
}

func (r *AdminRole) Fill(a *db.AdminRole) *AdminRole {
	r.ID = a.ID
	r.DisplayName = a.DisplayName
	r.Permissions = make([]*AdminPermission, 0)
	for _, permission := range a.Permissions {
		r.Permissions = append(r.Permissions, new(AdminPermission).Fill(permission))
	}
	return r
}

type ReqUpsertAdminRole struct {
	Name          string   `json:"name"`
	DisplayName   string   `json:"display_name"`
	PermissionIDs []uint64 `json:"permission_ids"`
}

type AdminPermission struct {
	ID          uint                     `json:"id"`
	Group       *AdminPermissionGroup    `json:"group"`
	Path        string                   `json:"path"`
	DisplayName string                   `json:"display_name"`
	Actions     []*AdminPermissionAction `json:"actions"`
}

func (r *AdminPermission) Fill(a *db.AdminPermission) *AdminPermission {
	r.ID = a.ID
	if a.Group != nil {
		r.Group = &AdminPermissionGroup{ID: a.Group.ID, Path: a.Group.Path, DisplayName: a.Group.DisplayName}
	}

	r.Path = a.Path
	r.DisplayName = a.DisplayName
	r.Actions = make([]*AdminPermissionAction, 0)
	for _, action := range a.Actions {
		r.Actions = append(r.Actions, &AdminPermissionAction{
			Method:      action.Method,
			Uri:         action.Uri,
			Description: action.Description,
		})
	}
	return r
}

type ReqUpsertAdminPermission struct {
	GroupID     uint                     `json:"group_id"`
	Path        string                   `json:"path"`
	DisplayName string                   `json:"display_name"`
	Actions     []*AdminPermissionAction `json:"actions"`
}

type AdminPermissionGroup struct {
	ID          uint               `json:"id"`
	Path        string             `json:"path"`
	DisplayName string             `json:"display_name"`
	Permissions []*AdminPermission `json:"permissions,omitempty"`
}

func (r *AdminPermissionGroup) Fill(a *db.AdminPermissionGroup) *AdminPermissionGroup {
	r.ID = a.ID
	r.Path = a.Path
	r.DisplayName = a.DisplayName
	r.Permissions = make([]*AdminPermission, 0)
	for _, permission := range a.Permissions {
		r.Permissions = append(r.Permissions, new(AdminPermission).Fill(permission))
	}
	return r
}

type ReqUpsertAdminPermissionGroup struct {
	Path        string `json:"path"`
	DisplayName string `json:"display_name"`
}

type AdminPermissionAction struct {
	Method      string `json:"method"`
	Uri         string `json:"uri"`
	Description string `json:"description"`
}
