package marketplace

import (
	"exapp-go/api"
	"exapp-go/internal/services/marketplace"
	"exapp-go/pkg/queryparams"

	"github.com/gin-gonic/gin"
)

// @Summary List all trading pools
// @Description Get a list of all trading pools
// @Tags pools
// @Accept json
// @Produce json
// @Param base_coin query string false "base coin"
// @Param quote_coin query string false "quote coin"
// @Success 200 {array} entity.PoolStats "pool info"
// @Router /api/v1/pools [get]
func pools(c *gin.Context) {
	queryParams := queryparams.NewQueryParams(c)
	poolService := marketplace.NewPoolService()
	pools, total, err := poolService.GetPools(c.Request.Context(), queryParams)
	if err != nil {
		api.Error(c, err)
		return
	}
	api.List(c, pools, total)
}

// @Summary Get pool details
// @Description Get detailed information about a specific trading pool
// @Tags pools
// @Accept json
// @Produce json
// @Param symbolOrId path string true "pool symbol or pool id"
// @Success 200 {object} entity.Pool ""
// @Router /api/v1/pools/{symbolOrId} [get]
func getPoolDetail(c *gin.Context) {
	symbolOrId := c.Param("symbolOrId")
	poolService := marketplace.NewPoolService()
	pool, err := poolService.GetPool(c.Request.Context(), symbolOrId)
	if err != nil {
		api.Error(c, err)
		return
	}
	api.OK(c, pool)
}







// @Summary Get kline data
// @Description Get kline data by pair id and interval
// @Tags kline
// @Accept json
// @Produce json
// @Param pool_id query string true "pool id"
// @Param interval query string true "interval" Enums(1m,5m,15m,30m,1h,4h,1d,1w,1M)
// @Param start query integer false "start timestamp"
// @Param end query integer false "end timestamp"
// @Param limit query integer false "limit count"
// @Success 200 {array} entity.Kline "kline data"
// @Router /api/v1/klines [get]
func klines(c *gin.Context) {

}
