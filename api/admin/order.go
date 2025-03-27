package admin

import (
	"exapp-go/api"
	"exapp-go/internal/services/admin"
	"exapp-go/pkg/queryparams"

	"github.com/gin-gonic/gin"
)

func queryOpenOrders(c *gin.Context) {

	resp, total, err := admin.New().QueryOpenOrders(c.Request.Context(), queryparams.NewQueryParams(c))
	if err != nil {
		api.Error(c, err)
		return
	}

	api.List(c, resp, total)

}

func getOrdersCoinTotal(c *gin.Context) {
	startTime := c.Query("start_time")
	endTime := c.Query("end_time")

	resp, err := admin.New().GetOrdersCoinTotal(c.Request.Context(), startTime, endTime)
	if err != nil {
		api.Error(c, err)
		return
	}

	api.List(c, resp, 0)
}
