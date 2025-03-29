package admin

import (
	"exapp-go/api"
	"exapp-go/internal/services/admin"
	"exapp-go/pkg/queryparams"

	"github.com/gin-gonic/gin"
)

func queryHistoryQrders(c *gin.Context) {

	resp, total, err := admin.New().QueryHistoryOrders(c.Request.Context(), queryparams.NewQueryParams(c))
	if err != nil {
		api.Error(c, err)
		return
	}

	api.List(c, resp, total)

}

func getOrdersCoinTotal(c *gin.Context) {

}

func getOrdersCoinQuantity(c *gin.Context) {
	startTime := c.Query("start_time")
	endTime := c.Query("end_time")

	resp, err := admin.New().GetOrdersCoinQuantity(c.Request.Context(), startTime, endTime)
	if err != nil {
		api.Error(c, err)
		return
	}

	api.List(c, resp, 0)
}

func getOrdersSymbolQuantity(c *gin.Context) {
	startTime := c.Query("start_time")
	endTime := c.Query("end_time")

	resp, err := admin.New().GetOrdersSymbolQuantity(c.Request.Context(), startTime, endTime)
	if err != nil {
		api.Error(c, err)
		return
	}

	api.List(c, resp, 0)
}

func getOrdersCoinFee(c *gin.Context) {
	startTime := c.Query("start_time")
	endTime := c.Query("end_time")

	resp, err := admin.New().GetOrdersCoinFee(c.Request.Context(), startTime, endTime)
	if err != nil {
		api.Error(c, err)
		return
	}

	api.List(c, resp, 0)
}

func getOrdersSymbolFee(c *gin.Context) {
	startTime := c.Query("start_time")
	endTime := c.Query("end_time")

	resp, err := admin.New().GetOrdersSymbolFee(c.Request.Context(), startTime, endTime)
	if err != nil {
		api.Error(c, err)
		return
	}

	api.List(c, resp, 0)
}
