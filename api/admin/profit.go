package admin

import (
	"exapp-go/api"
	"exapp-go/internal/services/admin"

	"github.com/gin-gonic/gin"
)

func getBalances(c *gin.Context) {
	isEvmUser := c.GetBool("is_evm_user")

	resp, err := admin.New().GetBalances(c.Request.Context(), isEvmUser)
	if err != nil {
		api.Error(c, err)
		return
	}

	api.OK(c, resp)
}

func getCoinBalances(c *gin.Context) {
	isEvmUser := c.GetBool("is_evm_user")

	resp, err := admin.New().GetCoinBalances(c.Request.Context(), isEvmUser)
	if err != nil {
		api.Error(c, err)
		return
	}

	api.OK(c, resp)
}
