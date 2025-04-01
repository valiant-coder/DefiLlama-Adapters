package admin

import (
	"exapp-go/api"
	entity_admin "exapp-go/internal/entity/admin"
	"exapp-go/internal/services/admin"
	"exapp-go/pkg/queryparams"

	"github.com/gin-gonic/gin"
)

func createUserPointsGrant(c *gin.Context) {
	var req entity_admin.ReqUserPointsGrant
	if err := c.ShouldBind(&req); err != nil {
		api.Error(c, err)
		return
	}

	operator := c.GetString("admin")

	grants, err := admin.New().CreateUserPointsGrant(c.Request.Context(), &req, operator)
	if err != nil {
		api.Error(c, err)
		return
	}
	api.OK(c, grants)
}

func queryUserPointsGrant(c *gin.Context) {
	resp, count, err := admin.New().QueryUserPointsGrant(c.Request.Context(), queryparams.NewQueryParams(c))
	if err != nil {
		api.Error(c, err)
		return
	}

	api.List(c, resp, count)
}
