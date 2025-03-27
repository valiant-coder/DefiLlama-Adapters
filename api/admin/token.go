package admin

import (
	"exapp-go/api"
	entity_admin "exapp-go/internal/entity/admin"
	"exapp-go/internal/services/admin"
	"exapp-go/pkg/queryparams"

	"github.com/gin-gonic/gin"
)

func queryTokens(c *gin.Context) {

	resp, total, err := admin.New().QueryTokens(c.Request.Context(), queryparams.NewQueryParams(c))
	if err != nil {
		api.Error(c, err)
		return
	}

	api.List(c, resp, total)
}

func createToken(c *gin.Context) {

	var req entity_admin.ReqUpsertToken
	if err := c.ShouldBind(&req); err != nil {
		api.Error(c, err)
		return
	}

	token, err := admin.New().CreateToken(c.Request.Context(), &req)
	if err != nil {
		api.Error(c, err)
		return
	}
	api.OK(c, token)
}
