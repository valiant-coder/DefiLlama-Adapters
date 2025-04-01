package admin

import (
	"exapp-go/api"
	"exapp-go/internal/services/admin"
	"exapp-go/pkg/queryparams"

	"github.com/gin-gonic/gin"
)

// @tags profit
// @Security ApiKeyAuth
// @Summary get balances
// @Accept json
// @Produce json
// @Param is_evm_user query bool true "is_evm_user"
// @Success 200 {object} decimal.Decimal "Successful response"
// @Router /balances [get]
func getBalances(c *gin.Context) {
	isEvmUser := c.GetBool("is_evm_user")

	resp, err := admin.New().GetBalances(c.Request.Context(), isEvmUser)
	if err != nil {
		api.Error(c, err)
		return
	}

	api.OK(c, resp)
}

// @tags profit
// @Security ApiKeyAuth
// @Summary get coin balances
// @Accept json
// @Produce json
// @Param is_evm_user query bool true "is_evm_user"
// @Success 200 {object} entity_admin.RespCoinBalance "Successful response"
// @Router /coin_balances [get]
func getCoinBalances(c *gin.Context) {
	isEvmUser := c.GetBool("is_evm_user")

	resp, err := admin.New().GetCoinBalances(c.Request.Context(), isEvmUser)
	if err != nil {
		api.Error(c, err)
		return
	}

	api.OK(c, resp)
}

// @tags profit
// @Security ApiKeyAuth
// @Summary get user balance stat
// @Accept json
// @Produce json
// @Param is_evm_user query bool true "is_evm_user"
// @Param min_value query int64 true "min_value"
// @Param max_value query int64 true "max_value"
// @Param range_count query int true "range_count"
// @Success 200 {object} db.BalanceRange "Successful response"
// @Router /user_balance_stat [get]
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

// @tags profit
// @Security ApiKeyAuth
// @Summary query user balance
// @Accept json
// @Produce json
// @Param username query string false "username"
// @Param uid query string false "uid"
// @Param limit query int false "limit"
// @Param offset query int false "offset"
// @Success 200 {array} entity_admin.RespUserBalance "Successful response"
// @Router /user_balances [get]
func queryUserBalance(c *gin.Context) {

	resp, err := admin.New().QueryUserBalance(c.Request.Context(), queryparams.NewQueryParams(c))
	if err != nil {
		api.Error(c, err)
		return
	}

	api.OK(c, resp)
}

// @tags profit
// @Security ApiKeyAuth
// @Summary get user coin balance
// @Accept json
// @Produce json
// @Param uid path string true "uid"
// @Success 200 {object} entity_admin.RespUserCoinBalanceAndUsdtAmount "Successful response"
// @Router /user_coin_balance/{uid} [get]
func getUserCoinBalance(c *gin.Context) {
	uid := c.Param("uid")

	resp, err := admin.New().GetUserCoinBalance(c.Request.Context(), uid)
	if err != nil {
		api.Error(c, err)
		return
	}

	api.OK(c, resp)
}
