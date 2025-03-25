package v1

import (
	"exapp-go/api"
	"exapp-go/internal/entity"
	"exapp-go/internal/services/marketplace"

	"github.com/gin-gonic/gin"
)

// @Summary Get user info
// @Description Get user info
// @Security ApiKeyAuth
// @Tags user
// @Accept json
// @Produce json
// @Success 200 {object} entity.RespV1UserInfo "user info"
// @Router /api/v1/info [get]
func getUserInfo(c *gin.Context) {
	subAccount := GetSubAccountFromContext(c)
	api.OK(c, entity.RespV1UserInfo{
		EOSAccount: subAccount.EOSAccount,
		Permission: subAccount.Permission,
	})
}

// @Summary Get user balances
// @Description Get user balances
// @Tags user
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {array} entity.SubAccountBalance "user balances"
// @Router /api/v1/balances [get]
func getUserBalances(c *gin.Context) {
	subAccount := GetSubAccountFromContext(c)

	userService := marketplace.NewUserService()
	balances, err := userService.GetUserSubaccountBalances(c.Request.Context(), subAccount.EOSAccount, subAccount.Permission)
	if err != nil {
		api.Error(c, err)
		return
	}
	api.OK(c, balances)
}
