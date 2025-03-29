package admin

import (
	"exapp-go/api"
	"exapp-go/internal/services/admin"
	"exapp-go/pkg/queryparams"

	"github.com/gin-gonic/gin"
)

// @tags order
// @Security ApiKeyAuth
// @Summary query history orders
// @Accept json
// @Produce json
// @Param pool_base_coin query string false "pool_base_coin"
// @Param pool_symbol query string false "pool_symbol"
// @Param app query string false "app"
// @Param trader query string false "trader"
// @Param start_time query string false "start_time"
// @Param end_time query string false "end_time"
// @Param limit query int false "limit"
// @Param offset query int false "offset"
// @Success 200 {array} entity_admin.RespHistoryOrder "Successful response"
// @Router /history_orders [get]

func queryHistoryOrders(c *gin.Context) {

	resp, total, err := admin.New().QueryHistoryOrders(c.Request.Context(), queryparams.NewQueryParams(c))
	if err != nil {
		api.Error(c, err)
		return
	}

	api.List(c, resp, total)

}

// @tags order
// @Security ApiKeyAuth
// @Summary get orders coin total
// @Accept json
// @Produce json
// @Param start_time query string false "start_time"
// @Param end_time query string false "end_time"
// @Success 200 {object} decimal.Decimal "Successful response"
// @Router /orders_coin_total [get]
func getOrdersCoinTotal(c *gin.Context) {
	startTime := c.Query("start_time")
	endTime := c.Query("end_time")

	resp, err := admin.New().GetOrdersCoinTotal(c.Request.Context(), startTime, endTime)
	if err != nil {
		api.Error(c, err)
		return
	}

	api.OK(c, resp)
}

// @tags order
// @Security ApiKeyAuth
// @Summary get orders coin quantity
// @Accept json
// @Produce json
// @Param start_time query string false "start_time"
// @Param end_time query string false "end_time"
// @Success 200 {array} entity_admin.RespOrdersCoinQuantity "Successful response"
// @Router /orders_coin_quantity [get]
func getOrdersCoinQuantity(c *gin.Context) {
	startTime := c.Query("start_time")
	endTime := c.Query("end_time")

	resp, err := admin.New().GetOrdersCoinQuantity(c.Request.Context(), startTime, endTime)
	if err != nil {
		api.Error(c, err)
		return
	}

	api.OK(c, resp)
}

// @tags order
// @Security ApiKeyAuth
// @Summary get orders symbol quantity
// @Accept json
// @Produce json
// @Param start_time query string false "start_time"
// @Param end_time query string false "end_time"
// @Success 200 {array} entity_admin.RespOrdersSymbolQuantity "Successful response"
// @Router /orders_symbol_quantity [get]
func getOrdersSymbolQuantity(c *gin.Context) {
	startTime := c.Query("start_time")
	endTime := c.Query("end_time")

	resp, err := admin.New().GetOrdersSymbolQuantity(c.Request.Context(), startTime, endTime)
	if err != nil {
		api.Error(c, err)
		return
	}

	api.OK(c, resp)
}

// @tags order
// @Security ApiKeyAuth
// @Summary get orders fee total
// @Accept json
// @Produce json
// @Param start_time query string false "start_time"
// @Param end_time query string false "end_time"
// @Success 200 {object} decimal.Decimal "Successful response"
// @Router /orders_fee_total [get]
func getOrdersFeeTotal(c *gin.Context) {
	startTime := c.Query("start_time")
	endTime := c.Query("end_time")

	resp, err := admin.New().GetOrdersFeeTotal(c.Request.Context(), startTime, endTime)
	if err != nil {
		api.Error(c, err)
		return
	}

	api.OK(c, resp)
}

// @tags order
// @Security ApiKeyAuth
// @Summary get orders coin fee
// @Accept json
// @Produce json
// @Param start_time query string false "start_time"
// @Param end_time query string false "end_time"
// @Success 200 {array} entity_admin.RespOrdersCoinFee "Successful response"
// @Router /orders_coin_fee [get]
func getOrdersCoinFee(c *gin.Context) {
	startTime := c.Query("start_time")
	endTime := c.Query("end_time")

	resp, err := admin.New().GetOrdersCoinFee(c.Request.Context(), startTime, endTime)
	if err != nil {
		api.Error(c, err)
		return
	}

	api.OK(c, resp)
}

// @tags order
// @Security ApiKeyAuth
// @Summary get orders symbol fee
// @Accept json
// @Produce json
// @Param start_time query string false "start_time"
// @Param end_time query string false "end_time"
// @Success 200 {array} entity_admin.RespOrdersSymbolFee "Successful response"
// @Router /orders_symbol_fee [get]
func getOrdersSymbolFee(c *gin.Context) {
	startTime := c.Query("start_time")
	endTime := c.Query("end_time")

	resp, err := admin.New().GetOrdersSymbolFee(c.Request.Context(), startTime, endTime)
	if err != nil {
		api.Error(c, err)
		return
	}

	api.OK(c, resp)
}
