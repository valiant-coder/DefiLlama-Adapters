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
// @Summary query admin permission groups
// @Success 200 {array} entity_admin.AdminPermissionGroup "Successful response"
// @Router /admin_permission_groups [get]
func getAdminPermissionGroups(c *gin.Context) {

	resp, count, err := admin.New().QueryAdminPermissionGroups(c.Request.Context(), queryparams.NewQueryParams(c))
	if err != nil {
		api.Error(c, err)
		return
	}

	api.List(c, resp, count)

}

// @tags admin
// @Security ApiKeyAuth
// @Summary get admin permission group
// @Param id path string true "admin permission group id"
// @Success 200 {object} entity_admin.AdminPermissionGroup "Successful response"
// @Router /admin_permission_group/{id} [get]
func getAdminPermissionGroup(c *gin.Context) {
	adminPermissionGroup, err := admin.New().GetAdminPermissionGroup(c.Request.Context(), c.Param("id"))
	if err != nil {
		api.Error(c, err)
		return
	}
	api.OK(c, adminPermissionGroup)
}

// @tags admin
// @Security ApiKeyAuth
// @Summary create admin permission group
// @Param body body entity_admin.ReqUpsertAdminPermissionGroup false "request body"
// @Success 200 {object} entity_admin.AdminPermissionGroup "Successful response"
// @Router /admin_permission_group [post]
func createAdminPermissionGroup(c *gin.Context) {
	var req entity_admin.ReqUpsertAdminPermissionGroup
	if err := c.ShouldBind(&req); err != nil {
		api.Error(c, err)
		return
	}

	adminPermissionGroup, err := admin.New().CreateAdminPermissionGroup(c.Request.Context(), &req)
	if err != nil {
		api.Error(c, err)
		return
	}
	api.OK(c, adminPermissionGroup)

}

// @tags admin
// @Security ApiKeyAuth
// @Summary update admin permission group
// @Param id path string true "admin permission group id"
// @Param body body entity_admin.ReqUpsertAdminPermissionGroup false "request body"
// @Success 200 {object} entity_admin.AdminPermissionGroup "Successful response"
// @Router /admin_permission_group/{id} [put]
func updateAdminPermissionGroup(c *gin.Context) {
	var req entity_admin.ReqUpsertAdminPermissionGroup
	if err := c.ShouldBind(&req); err != nil {
		api.Error(c, err)
		return
	}

	adminPermissionGroup, err := admin.New().UpdateAdminPermissionGroup(c.Request.Context(), &req, c.Param("id"))
	if err != nil {
		api.Error(c, err)
		return
	}
	api.OK(c, adminPermissionGroup)	


}
