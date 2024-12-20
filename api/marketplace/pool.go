package marketplace

import (
	"github.com/gin-gonic/gin"
)

// @Summary List all trading pools
// @Description Get a list of all trading pools
// @Tags pools
// @Accept json
// @Produce json
// @Param base_symbol query string false "base coin symbol"
// @Param quote_symbol query string false "quote coin symbol"
// @Success 200 {array} entity.PoolInfo "pool info"
// @Router /api/v1/pools [get]
func pools(c *gin.Context) {



}


// @Summary Get pool details
// @Description Get detailed information about a specific trading pool
// @Tags pools
// @Accept json
// @Produce json
// @Param symbol path string true "pool symbol"
// @Success 200 {object} entity.Pool ""
// @Router /api/v1/pools/{symbol} [get]
func getPool(c *gin.Context) {


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
