package marketplace

import "github.com/gin-gonic/gin"

// @Summary Get trades
// @Description Get trades
// @Security ApiKeyAuth
// @Tags trade
// @Accept json
// @Produce json
// @Success 200 {array} entity.Trade "trade list"
// @Router /trades [get]
func getTrades(c *gin.Context) {

}
