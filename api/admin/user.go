package admin

import (
	"errors"
	"exapp-go/api"
	"exapp-go/internal/services/admin"
	"exapp-go/pkg/queryparams"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

// @tags user
// @Security ApiKeyAuth
// @Summary query users
// @Accept json
// @Produce json
// @Param username query string false "username"
// @Param uid query string false "uid"
// @Param login_method query string false "login_method"
// @Param start_time query string false "start_time"
// @Param end_time query string false "end_time"
// @Param limit query int false "limit"
// @Param offset query int false "offset"
// @Success 200 {array} entity_admin.RespUser "Successful response"
// @Router /users [get]
func queryUsers(c *gin.Context) {

	resp, total, err := admin.New().QueryUsers(c.Request.Context(), queryparams.NewQueryParams(c))
	if err != nil {
		api.Error(c, err)
		return
	}

	api.List(c, resp, total)
}

// @tags user
// @Security ApiKeyAuth
// @Summary get user passkeys
// @Success 200 {array} entity_admin.RespPasskey "Successful response"
// @Router /user-passkeys/{uid} [get]
func getUserPasskeys(c *gin.Context) {
	uid := c.GetString("uid")
	if uid == "" {
		api.Error(c, errors.New("uid is empty"))
		return
	}
	queryParams := queryparams.NewQueryParams(c)

	resp, total, err := admin.New().GetPasskeys(c.Request.Context(), queryParams)
	if err != nil {
		api.Error(c, err)
		return
	}

	api.List(c, resp, total)
}

// @Tags user
// @Summary Get users statis
// @Description Get users statis
// @Accept json
// @Produce json
// @Param time_dimension query string true "month week day"
// @Param data_type query string true "add_user_count add_passkey_count add_evm_count add_deposit_count"
// @Param amount query string false "amount"
// @Success 200 {object} db.UsersStatis "users statis"
// @Router /users_statis [get]
func getUsersStatis(c *gin.Context) {

	timeDimension := c.Query("time_dimension")
	dataType := c.Query("data_type")
	amount := c.Query("amount")

	resp, total, err := admin.New().GetUsersStatis(c.Request.Context(), timeDimension, dataType, cast.ToInt(amount))
	if err != nil {
		api.Error(c, err)
		return
	}

	api.List(c, resp, total)
}

// @tags user
// @Security ApiKeyAuth
// @Summary query transactions record
// @Accept json
// @Produce json
// @Param symbol query string false "symbol"
// @Param chain_name query string false "chain_name"
// @Param uid query string false "uid"
// @Param tx_hash query string false "tx_hash"
// @Param login_method query string false "login_method"
// @Param start_time query string false "start_time"
// @Param end_time query string false "end_time"
// @Success 200 {array} db.TransactionsRecord "Successful response"
// @Router /user_transactions_records [get]
func getTransactionsRecord(c *gin.Context) {
	queryParams := queryparams.NewQueryParams(c)

	resp, total, err := admin.New().GetTransactionsRecord(c.Request.Context(), queryParams)
	if err != nil {
		api.Error(c, err)
		return
	}

	api.List(c, resp, total)
}

// @Tags User
// @Summary Get deposit amount total
// @Description Get deposit amount total
// @Accept json
// @Produce json
// @Param start_time query string false "2006-01-06 00:00:00"
// @Param end_time query string false "2006-01-06 23:59:59"
// @Success 200 {array} entity_admin.RespGetDepositWithdrawTotal "Successful response"
// @Router /deposit_amount_total [get]
func getDepositAmountTotal(c *gin.Context) {
	startTime := c.Query("start_time")
	endTime := c.Query("end_time")

	resp, err := admin.New().GetDepositAmountTotal(c.Request.Context(), startTime, endTime)
	if err != nil {
		api.Error(c, err)
		return
	}

	api.List(c, resp, 0)
}

// @Tags User
// @Summary Get withdraw amount total
// @Description Get withdraw amount total
// @Accept json
// @Produce json
// @Param start_time query string false "2006-01-06 00:00:00"
// @Param end_time query string false "2006-01-06 23:59:59"
// @Success 200 {array} entity_admin.RespGetDepositWithdrawTotal "Successful response"
// @Router /withdraw_amount_total [get]
func getWithdrawAmountTotal(c *gin.Context) {
	startTime := c.Query("start_time")
	endTime := c.Query("end_time")

	resp, err := admin.New().GetWithdrawAmountTotal(c.Request.Context(), startTime, endTime)
	if err != nil {
		api.Error(c, err)
		return
	}

	api.List(c, resp, 0)
}
