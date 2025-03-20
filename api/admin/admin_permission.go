package admin

import (
	"exapp-go/api"
	entity_admin "exapp-go/internal/entity/admin"
	"exapp-go/internal/services/admin"
	"exapp-go/pkg/queryparams"

	"github.com/gin-gonic/gin"
)

// @tags admin
// @Security ApiKeyAuth
// @Summary query admin permissions
// @Success 200 {array} entity_admin.AdminPermission "Successful response"
// @Router /admin_permissions [get]
func getAdminPermissions(c *gin.Context) {
	
	resp, count, err := admin.New().QueryAdminPermissions(c.Request.Context(), queryparams.NewQueryParams(c))
	if err != nil {
		api.Error(c, err)
		return
	}

	api.List(c, resp, count)

}

// @tags admin
// @Security ApiKeyAuth
// @Summary get admin permission
// @Param id path string true "admin permission id"
// @Success 200 {object} entity_admin.AdminPermission "Successful response"
// @Router /admin_permission/{id} [get]
func getAdminPermission(c *gin.Context) {
	adminPermission, err := admin.New().GetAdminPermission(c.Request.Context(), c.Param("id"))
	if err != nil {
		api.Error(c, err)
		return
	}
	api.OK(c, adminPermission)
}

// @tags admin
// @Security ApiKeyAuth
// @Summary create admin permission
// @Param body body entity_admin.ReqUpsertAdminPermission false "request body"
// @Success 200 {object} entity_admin.AdminPermission "Successful response"
// @Router /admin_permission [post]
func createAdminPermission(c *gin.Context) {
	var req entity_admin.ReqUpsertAdminPermission
	if err := c.ShouldBind(&req); err != nil {
		api.Error(c, err)
		return
	}

	adminPermission, err := admin.New().CreateAdminPermission(c.Request.Context(), &req)
	if err != nil {
		api.Error(c, err)
		return
	}
	api.OK(c, adminPermission)

}

// @tags admin
// @Security ApiKeyAuth
// @Summary update admin permission
// @Param id path string true "admin permission id"
// @Param body body entity_admin.ReqUpsertAdminPermission false "request body"
// @Success 200 {object} entity_admin.AdminPermission "Successful response"
// @Router /admin_permission/{id} [put]
func updateAdminPermission(c *gin.Context) {
	var req entity_admin.ReqUpsertAdminPermission
	if err := c.ShouldBind(&req); err != nil {
		api.Error(c, err)
		return
	}

	adminPermission, err := admin.New().UpdateAdminPermission(c.Request.Context(), &req, c.Param("id"))
	if err != nil {
		api.Error(c, err)
		return
	}
	api.OK(c, adminPermission)	
}
