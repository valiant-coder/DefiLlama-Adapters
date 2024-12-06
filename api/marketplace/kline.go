package marketplace

import "github.com/gin-gonic/gin"

// @Summary Get kline data
// @Description Get kline data by pair id and interval
// @Tags kline
// @Accept json
// @Produce json
// @Param pair_id query string true "pair_id"
// @Param interval query string true "interval" Enums(1m,5m,15m,30m,1h,4h,1d,1w,1M)
// @Param start query integer false "start timestamp"
// @Param end query integer false "end timestamp"
// @Param limit query integer false "limit count"
// @Success 200 {array} entity.Kline "kline data"
// @Router /klines [get]
func getKlines(c *gin.Context) {

} 