package marketplace

import (
	"exapp-go/api"
	"exapp-go/internal/entity"
	"exapp-go/internal/services/marketplace"

	"github.com/gin-gonic/gin"
)

// @Summary First deposit
// @Description First deposit
// @Security ApiKeyAuth
// @Tags deposit
// @Accept json
// @Produce json
// @Param req body entity.ReqFirstDeposit true "first deposit params"
// @Success 200 {object} entity.RespFirstDeposit ""
// @Router /first-deposit [post]
func firstDeposit(c *gin.Context) {
	depositWithdrawalService := marketplace.NewDepositWithdrawalService()
	var req entity.ReqFirstDeposit
	if err := c.ShouldBind(&req); err != nil {
		api.Error(c, err)
		return
	}
	resp, err := depositWithdrawalService.FirstDeposit(c.Request.Context(), c.GetString("uid"), &req)
	if err != nil {
		api.Error(c, err)
		return
	}
	api.OK(c, resp)
}

// @Summary Deposit
// @Description Deposit
// @Security ApiKeyAuth
// @Tags deposit
// @Accept json
// @Produce json
// @Param req body entity.ReqDeposit true "deposit params"
// @Success 200 {object} entity.RespDeposit ""
// @Router /deposit [post]
func deposit(c *gin.Context) {
	depositWithdrawalService := marketplace.NewDepositWithdrawalService()
	var req entity.ReqDeposit
	if err := c.ShouldBind(&req); err != nil {
		api.Error(c, err)
		return
	}
	resp, err := depositWithdrawalService.Deposit(c.Request.Context(), c.GetString("uid"), &req)
	if err != nil {
		api.Error(c, err)
		return
	}
	api.OK(c, resp)
}


func withdrawal(c *gin.Context) {

}

func getDepositHistory(c *gin.Context) {

}

func getWithdrawalHistory(c *gin.Context) {

}
