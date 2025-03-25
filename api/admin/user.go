package admin

import (
	"exapp-go/api"
	"exapp-go/internal/services/admin"
	"exapp-go/pkg/queryparams"

	"github.com/gin-gonic/gin"
)

// @tags admin
// @Security ApiKeyAuth
// @Summary query admins
// @Success 200 {array} entity_admin.RespUser "Successful response"
// @Router /users [get]
func queryUsers(c *gin.Context) {

	resp, count, err := admin.New().QueryUsers(c.Request.Context(), queryparams.NewQueryParams(c))
	if err != nil {
		api.Error(c, err)
		return
	}

	api.List(c, resp, count)
}
