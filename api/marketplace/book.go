package marketplace

import "github.com/gin-gonic/gin"

// @Summary Get order book
// @Description Get order book by pair id
// @Tags book
// @Accept json
// @Produce json
// @Param pair_id query string true "pair_id"
// @Success 200 {object} entity.OrderBook "order book"
// @Router /book [get]
func getOrderBook(c *gin.Context) {

} 