package admin

import (
	"exapp-go/api"
	"exapp-go/internal/services/admin"
	"exapp-go/pkg/queryparams"

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

func getUserBalanceStat(c *gin.Context) {

	isEvmUser := c.GetBool("is_evm_user")
	minValue := c.GetInt64("min_value")
	maxValue := c.GetInt64("max_value")
	rangeCount := c.GetInt("range_count")

	resp, err := admin.New().GetUserBalanceStat(c.Request.Context(), isEvmUser, minValue, maxValue, rangeCount)
	if err != nil {
		api.Error(c, err)
		return
	}

	api.OK(c, resp)
}

func queryUserBalance(c *gin.Context) {

	resp, err := admin.New().QueryUserBalance(c.Request.Context(), queryparams.NewQueryParams(c))
	if err != nil {
		api.Error(c, err)
		return
	}

	api.OK(c, resp)
}
