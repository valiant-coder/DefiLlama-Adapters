package marketplace

import (
	"exapp-go/api"
	"exapp-go/internal/entity"
	"exapp-go/internal/services/marketplace"
	"exapp-go/pkg/queryparams"

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

// @Summary Get deposit history
// @Description Get deposit history
// @Security ApiKeyAuth
// @Tags deposit
// @Accept json
// @Produce json
// @Success 200 {array} entity.RespDepositRecord "deposit records"
// @Router /deposit-history [get]
func getDepositHistory(c *gin.Context) {
	depositWithdrawalService := marketplace.NewDepositWithdrawalService()
	queryParams := queryparams.NewQueryParams(c)
	resp, total, err := depositWithdrawalService.GetDepositRecords(c.Request.Context(), c.GetString("uid"), queryParams)
	if err != nil {
		api.Error(c, err)
		return
	}
	api.List(c, resp, total)
}

// @Summary Get withdrawal history
// @Description Get withdrawal history
// @Security ApiKeyAuth
// @Tags withdrawal
// @Accept json
// @Produce json
// @Success 200 {array} entity.RespWithdrawRecord "withdrawal records"
// @Router /withdrawal-history [get]
func getWithdrawalHistory(c *gin.Context) {
	depositWithdrawalService := marketplace.NewDepositWithdrawalService()
	queryParams := queryparams.NewQueryParams(c)
	resp, total, err := depositWithdrawalService.GetWithdrawalRecords(c.Request.Context(), c.GetString("uid"), queryParams)
	if err != nil {
		api.Error(c, err)
		return
	}
	api.List(c, resp, total)
}
