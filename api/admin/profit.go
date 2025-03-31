package admin

import (
	"exapp-go/api"
	"exapp-go/internal/services/admin"

	"github.com/gin-gonic/gin"
)

func getCoinBalances(c *gin.Context) {
	isEvmUser := c.GetBool("is_evm_user")

	resp, err := admin.New().GetCoinBalances(c.Request.Context(), isEvmUser)
	if err != nil {
		api.Error(c, err)
		return
	}

	api.OK(c, resp)
}
