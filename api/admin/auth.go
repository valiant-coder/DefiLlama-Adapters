package admin

import (
	"exapp-go/api"
	entity_admin "exapp-go/internal/entity/admin"
	"exapp-go/internal/services/admin"

	"github.com/gin-gonic/gin"
)

// @tags auth
// @Summary login
// @Param body body entity_admin.ReqLogin true "request body"
// @Success 200 {object} entity_admin.RespLogin "Successful response"
// @Router /login [post]
func login(c *gin.Context) {
	var req entity_admin.ReqLogin
	if err := c.ShouldBind(&req); err != nil {
		api.Error(c, err)
		return
	}

	resp, err := admin.New().Login(c.Request.Context(), &req)
	if err != nil {
		api.Error(c, err)
		return
	}
	api.OK(c, resp)

}

// @tags auth
// @Summary two step auth
// @Param body body entity_admin.ReqAuth true "request body"
// @Success 200 {object} entity_admin.RespAuth "Successful response"
// @Router /auth [post]
func auth(c *gin.Context) {
}

// @tags auth
// @Summary reset password
// @Param body body entity_admin.ReqResetPassword true "request body"
// @Success 204
// @Router /auth/reset_password [post]
func resetPassword(c *gin.Context) {
	var req entity_admin.ReqResetPassword
	if err := c.ShouldBind(&req); err != nil {
		api.Error(c, err)
		return
	}

	err := admin.New().ResetPassword(c.Request.Context(), &req)
	if err != nil {
		api.Error(c, err)
		return
	}
	api.NoContent(c)

}

// @tags auth
// @Summary get google auth secret
// @Param name path string true "name"
// @Success 200 {object} entity_admin.RespGoogleAuth "Successful response"
// @Router /auth/google_auth_secret/{name} [get]
func getGoogleAuthSecret(c *gin.Context) {
	name := c.Param("name")
	resp, err := admin.New().GetGoogleAuthSecret(c.Request.Context(), name)
	if err != nil {
		api.Error(c, err)
		return
	}
	api.OK(c, resp)

}
