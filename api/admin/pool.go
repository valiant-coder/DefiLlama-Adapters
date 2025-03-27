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
// @Summary query pool
// @Accept json
// @Produce json
// @Param base_symbol query string false "base_symbol"
// @Param base_contract query string false "base_contract"
// @Param quote_symbol query string false "quote_symbol"
// @Param quote_contract query string false "quote_contract"
// @Param limit query int false "limit"
// @Param offset query int false "offset"
// @Success 200 {array} entity_admin.RespPool "Successful response"
// @Router /pools [get]
func queryPools(c *gin.Context) {

	resp, total, err := admin.New().QueryPools(c.Request.Context(), queryparams.NewQueryParams(c))
	if err != nil {
		api.Error(c, err)
		return
	}

	api.List(c, resp, total)
}

// @tags pool
// @Security ApiKeyAuth
// @Summary update pool
// @Param id path string true "pool_id"
// @Param body body entity_admin.ReqUpsertPool false "request body"
// @Success 200 {object} entity_admin.RespPool "Successful response"
// @Router /token/{id} [put]
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

// @tags pool
// @Security ApiKeyAuth
// @summary create pool
// @Param body body entity_admin.ReqUpsertPool false "request body"
// @Success 200 {object} entity_admin.RespPool "Successful response"
// @Router /token [post]
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
