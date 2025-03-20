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
// @Summary query admin roles
// @Success 200 {array} entity_admin.AdminRole "Successful response"
// @Router /admin_roles [get]
func queryAdminRoles(c *gin.Context) {
	resp, count, err := admin.New().QueryAdminRoles(c.Request.Context(), queryparams.NewQueryParams(c))
	if err != nil {
		api.Error(c, err)
		return
	}

	api.List(c, resp, count)
}

// @tags admin
// @Security ApiKeyAuth
// @Summary get admin role
// @Param id path string true "admin role id"
// @Success 200 {object} entity_admin.AdminRole "Successful response"
// @Router /admin_role/{id} [get]
func getAdminRole(c *gin.Context) {
	id := c.Param("id")
	adminRole, err := admin.New().GetAdminRole(c.Request.Context(), id)
	if err != nil {
		api.Error(c, err)
		return
	}
	api.OK(c, adminRole)

}

// @tags admin
// @Security ApiKeyAuth
// @Summary create admin role
// @Param body body entity_admin.ReqUpsertAdminRole false "request body"
// @Success 200 {object} entity_admin.AdminRole "Successful response"
// @Router /admin_role [post]
func createAdminRole(c *gin.Context) {
	var req entity_admin.ReqUpsertAdminRole
	if err := c.ShouldBind(&req); err != nil {
		api.Error(c, err)
		return
	}

	adminRole, err := admin.New().CreateAdminRole(c.Request.Context(), &req)
	if err != nil {
		api.Error(c, err)
		return
	}
	api.OK(c, adminRole)

}

// @tags admin
// @Security ApiKeyAuth
// @Summary update admin role
// @Param id path string true "admin role id"
// @Param body body entity_admin.ReqUpsertAdminRole false "request body"
// @Success 200 {object} entity_admin.AdminRole "Successful response"
// @Router /admin_role/{id} [put]
func updateAdminRole(c *gin.Context) {
	var req entity_admin.ReqUpsertAdminRole
	if err := c.ShouldBind(&req); err != nil {
		api.Error(c, err)
		return
	}
	roleID := c.Param("id")
	adminRole, err := admin.New().UpdateAdminRole(c.Request.Context(), &req,roleID)
	if err != nil {
		api.Error(c, err)
		return
	}
	api.OK(c, adminRole)

}
