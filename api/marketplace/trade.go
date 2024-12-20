package marketplace

import "github.com/gin-gonic/gin"

// @Summary Get trades
// @Description Get trades
// @Security ApiKeyAuth
// @Tags trade
// @Accept json
// @Produce json
// @Param pool_id query string false "pool_id"
// @Param start query integer false "start timestamp"
// @Param end query integer false "end timestamp"
// @Param limit query integer false "limit count"
// @Success 200 {array} entity.Trade "trade list"
// @Router /api/v1/trades [get]
func getTrades(c *gin.Context) {

}
