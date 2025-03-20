package admin

import (
	"exapp-go/api"
	"exapp-go/internal/services/admin"
	"exapp-go/pkg/queryparams"

	entity_admin "exapp-go/internal/entity/admin"

	"github.com/gin-gonic/gin"
)

// @tags admin
// @Security ApiKeyAuth
// @Summary query admins
// @Success 200 {array} entity_admin.RespAdmin "Successful response"
// @Router /admins [get]
func queryAdmins(c *gin.Context) {
	resp, count, err := admin.New().QueryAdmins(c.Request.Context(), queryparams.NewQueryParams(c))
	if err != nil {
		api.Error(c, err)
		return
	}

	api.List(c, resp, count)
}

// @tags admin
// @Security ApiKeyAuth
// @Summary get admin info
// @Param name path string true "admin name"
// @Success 200 {object} entity_admin.RespAdmin "Successful response"
// @Router /admin/{name} [get]
func getAdmin(c *gin.Context) {
	name := c.Param("name")
	operator := c.GetString("admin")
	if name != operator {
		api.NoPermission(c)
		return
	}
	admin, err := admin.New().GetAdmin(c.Request.Context(), name)
	if err != nil {
		api.Error(c, err)
		return
	}
	api.OK(c, admin)

}

// @tags admin
// @Security ApiKeyAuth
// @Summary create admin
// @Param body body entity_admin.ReqUpsertAdmin false "request body"
// @Success 200 {object} entity_admin.RespAdmin "Successful response"
// @Router /admin [post]
func createAdmin(c *gin.Context) {
	var req entity_admin.ReqUpsertAdmin
	if err := c.ShouldBind(&req); err != nil {
		api.Error(c, err)
		return
	}
	operator := c.GetString("admin")

	admin, err := admin.New().CreateAdmin(c.Request.Context(), &req, operator)
	if err != nil {
		api.Error(c, err)
		return
	}
	api.OK(c, admin)
}

// @tags admin
// @Security ApiKeyAuth
// @Summary update admin
// @Param name path string true "admin name"
// @Param body body entity_admin.ReqUpsertAdmin false "request body"
// @Success 200 {object} entity_admin.RespAdmin "Successful response"
// @Router /admin/{name} [put]
func updateAdmin(c *gin.Context) {
	var req entity_admin.ReqUpsertAdmin
	if err := c.ShouldBind(&req); err != nil {
		api.Error(c, err)
		return
	}
	name := c.Param("name")
	operator := c.GetString("admin")
	admin, err := admin.New().UpdateAdmin(c.Request.Context(), &req, name, operator)
	if err != nil {
		api.Error(c, err)
		return
	}
	api.OK(c, admin)

}

// @tags admin
// @Security ApiKeyAuth
// @Summary delete admin
// @Param name path string true "admin name"
// @Success 204
// @Router /admin/{name} [delete]
func deleteAdmin(c *gin.Context) {

	name := c.Param("name")
	err := admin.New().DeleteAdmin(c.Request.Context(), name)
	if err != nil {
		api.Error(c, err)
		return
	}
	api.NoContent(c)

}
