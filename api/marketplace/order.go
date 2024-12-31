package marketplace

import (
	"exapp-go/api"
	"exapp-go/internal/services/marketplace"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

// @Summary Get open orders
// @Description Get open orders
// @Tags order
// @Accept json
// @Produce json
// @Param pool_id query string false "pool_id"
// @Param account query string false "eos account name"
// @Success 200 {array} entity.Order "open order list"
// @Router /api/v1/open-orders [get]
func getOpenOrders(c *gin.Context) {

}

// @Summary Get history orders
// @Description Get history orders
// @Tags order
// @Accept json
// @Produce json
// @Param pool_id query string false "pool_id"
// @Param account query string false "eos account name"
// @Param order_type query string false "order_type"
// @Param status query string false "status"
// @Success 200 {array} entity.Order "order list"
// @Router /api/v1/orders [get]
func getOrders(c *gin.Context) {

}

// @Summary Get order detail
// @Description Get order
// @Tags order
// @Accept json
// @Produce json
// @Param id path int true "Order ID"
// @Success 200 {object} entity.OrderDetail "order detail"
// @Router /api/v1/orders/{id} [get]
func getOrder(c *gin.Context) {

}


// @Summary Get depth
// @Description Get order book by pool id
// @Tags depth
// @Accept json
// @Produce json
// @Param pool_id query string true "pool_id"
// @Success 200 {object} entity.Depth "order depth"
// @Router /api/v1/depth [get]
func getDepth(c *gin.Context) {
	poolID := c.Query("pool_id")
	depth, err := marketplace.NewDepthService().GetDepth(c.Request.Context(), cast.ToUint64(poolID))
	if err != nil {
		api.Error(c, err)
		return
	}
	api.OK(c, depth)
}
