package v1

import (
	"exapp-go/api"
	"exapp-go/internal/services/marketplace"

	"github.com/gin-gonic/gin"
)

// @Summary Get user info
// @Description Get user info
// @Security ApiKeyAuth
// @Tags user
// @Accept json
// @Produce json
// @Success 200 {object} entity.RespUserInfo "user info"
// @Router /user-info [get]
func getUserInfo(c *gin.Context) {
	userService := marketplace.NewUserService()
	userInfo, err := userService.GetUserInfo(c.Request.Context(), c.GetString("uid"))
	if err != nil {
		api.Error(c, err)
		return
	}
	api.OK(c, userInfo)
}

// @Summary Get user balances
// @Description Get user balances
// @Tags user
// @Accept json
// @Produce json
// @Param account query string false "eos account name"
// @Success 200 {array} entity.UserBalance "user balances"
// @Router /api/v1/balances [get]
func getUserBalances(c *gin.Context) {
	accountName := c.Query("account")
	userService := marketplace.NewUserService()
	balances, err := userService.GetUserBalance(c.Request.Context(), accountName)
	if err != nil {
		api.Error(c, err)
		return
	}
	api.OK(c, balances)
}
