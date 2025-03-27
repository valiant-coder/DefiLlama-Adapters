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
