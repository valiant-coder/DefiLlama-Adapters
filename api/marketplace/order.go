package marketplace

import (
	"exapp-go/api"
	"exapp-go/internal/entity"
	"exapp-go/internal/services/marketplace"

	"exapp-go/pkg/queryparams"

	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

// @Summary Get open orders
// @Description Get open orders
// @Tags order
// @Accept json
// @Produce json
// @Param pool_id query string false "pool_id"
// @Param trader query string true "eos account name"
// @Param side    query string false "0 buy 1 sell"
// @Success 200 {array} entity.OpenOrder "open order list"
// @Router /api/v1/open-orders [get]
func getOpenOrders(c *gin.Context) {
	queryParams := queryparams.NewQueryParams(c)

	orders, total, err := marketplace.NewOrderService().GetOpenOrders(c.Request.Context(), queryParams)
	if err != nil {
		api.Error(c, err)
		return
	}
	api.List(c, orders, total)
}

// @Summary Get history orders
// @Description Get history orders
// @Tags order
// @Accept json
// @Produce json
// @Param pool_id query string false "pool_id"
// @Param trader query string false "eos account name"
// @Param side  query string false "0 buy 1 sell"
// @Param type  query string false "0 market 1limit"
// @Param status query string false "status"
// @Success 200 {array} entity.Order "history orders"
// @Router /api/v1/history-orders [get]
func getHistoryOrders(c *gin.Context) {
	queryParams := queryparams.NewQueryParams(c)

	orders, total, err := marketplace.NewOrderService().GetHistoryOrders(c.Request.Context(), queryParams)
	if err != nil {
		api.Error(c, err)
		return
	}
	api.List(c, orders, total)
}

// @Summary Get history order detail
// @Description Get history order detail
// @Tags order
// @Accept json
// @Produce json
// @Param id path string true "pool_id+order_id+side,ps:0-1-0 pool_id = 0,order_id = 1,side = buy"
// @Success 200 {object} entity.OrderDetail "history order detail"
// @Router /api/v1/orders/{id} [get]
func getOrderDetail(c *gin.Context) {
	id := c.Param("id")
	order, err := marketplace.NewOrderService().GetOrderDetail(c.Request.Context(), id)
	if err != nil {
		api.Error(c, err)
		return
	}
	api.OK(c, order)
}

// @Summary Get depth
// @Description Get order book by pool id
// @Tags depth
// @Accept json
// @Produce json
// @Param pool_id query string true "pool_id"
// @Param precision query string false "0.00000001 ~ 10000"
// @Param limit query int false "limit"
// @Success 200 {object} entity.Depth "order depth"
// @Router /api/v1/depth [get]
func getDepth(c *gin.Context) {
	poolID := c.Query("pool_id")
	precision := c.Query("precision")
	limit := c.Query("limit")
	depth, err := marketplace.NewDepthService().GetDepth(c.Request.Context(), cast.ToUint64(poolID), precision, cast.ToInt(limit))
	if err != nil {
		api.Error(c, err)
		return
	}
	api.OK(c, depth)
}

// @Summary Mark order as read
// @Description Mark order as read and remove notification dot
// @Tags order
// @Accept json
// @Produce json
// @Param req body entity.ReqMakeOrderAsRead true "Request to mark order as read"
// @Success 200
// @Router /api/v1/orders/read [post]
func markOrderAsRead(c *gin.Context) {
	var req entity.ReqMakeOrderAsRead
	if err := c.ShouldBindJSON(&req); err != nil {
		api.Error(c, err)
		return
	}

	orderService := marketplace.NewOrderService()
	err := orderService.MarkOrderAsRead(c.Request.Context(), req.Trader, req.ID)
	if err != nil {
		api.Error(c, err)
		return
	}

	api.OK(c, nil)
}

// @Summary Check for unread orders
// @Description Check if user has any unread completed orders
// @Tags order
// @Accept json
// @Produce json
// @Param trader query string true "trader eos account name"
// @Success 200 {object} entity.RespUnreadOrder "Response for unread status"
// @Router /api/v1/unread-orders [get]
func checkUnreadOrders(c *gin.Context) {
	trader := c.Query("trader")

	if trader == "" {
		api.Error(c, fmt.Errorf("trader is required"))
		return
	}

	hasUnread, err := marketplace.NewOrderService().CheckUnreadFilledOrders(c.Request.Context(), trader)
	if err != nil {
		api.Error(c, err)
		return
	}

	api.OK(c, entity.RespUnreadOrder{HasUnread: hasUnread})
}

// @Summary Clear all unread orders
// @Description Clear all unread orders for a trader
// @Tags order
// @Accept json
// @Produce json
// @Param trader query string true "trader eos account name"
// @Success 200
// @Router /api/v1/orders/clear-unread [post]
func clearAllUnreadOrders(c *gin.Context) {
	trader := c.Query("trader")

	if trader == "" {
		api.Error(c, fmt.Errorf("trader is required"))
		return
	}

	err := marketplace.NewOrderService().ClearAllUnreadOrders(c.Request.Context(), trader)
	if err != nil {
		api.Error(c, err)
		return
	}

	api.OK(c, nil)
}
