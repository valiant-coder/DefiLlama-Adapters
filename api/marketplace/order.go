package marketplace

import "github.com/gin-gonic/gin"

// @Summary Get orders
// @Description Get orders
// @Tags order
// @Accept json
// @Produce json
// @Param pair_id query string false "pair_id"
// @Param order_type query string false "order_type"
// @Param status query string false "status"
// @Success 200 {array} entity.Order "order list"
// @Router /orders [get]
func getOrders(c *gin.Context) {



}
