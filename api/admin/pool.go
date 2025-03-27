package admin

import (
	"exapp-go/api"
	entity_admin "exapp-go/internal/entity/admin"
	"exapp-go/internal/services/admin"
	"exapp-go/pkg/queryparams"

	"github.com/gin-gonic/gin"
)

func queryPools(c *gin.Context) {

	resp, total, err := admin.New().QueryPools(c.Request.Context(), queryparams.NewQueryParams(c))
	if err != nil {
		api.Error(c, err)
		return
	}

	api.List(c, resp, total)
}

func updatePool(c *gin.Context) {

	var req entity_admin.ReqUpsertPool
	if err := c.ShouldBindJSON(&req); err != nil {
		api.Error(c, err)
		return
	}

	poolId := c.GetUint64("pool_id")
	pool, err := admin.New().UpdatePool(c.Request.Context(), req, poolId)
	if err != nil {
		api.Error(c, err)
		return
	}

	api.OK(c, pool)
}

func createPool(c *gin.Context) {

	var req entity_admin.ReqUpsertPool
	if err := c.ShouldBindJSON(&req); err != nil {
		api.Error(c, err)
		return
	}

	pool, err := admin.New().CreatePool(c.Request.Context(), &req)
	if err != nil {
		api.Error(c, err)
		return
	}

	api.OK(c, pool)
}
