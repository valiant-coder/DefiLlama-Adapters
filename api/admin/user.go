package admin

import (
	"errors"
	"exapp-go/api"
	entity_admin "exapp-go/internal/entity/admin"
	"exapp-go/internal/services/admin"
	"exapp-go/pkg/queryparams"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

// @tags admin
// @Security ApiKeyAuth
// @Summary query users
// @Accept json
// @Produce json
// @Param username query string false "username"
// @Param uid query string false "uid"
// @Param login_method query string false "login_method"
// @Param start_time query string false "start_time"
// @Param end_time query string false "end_time"
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

// @tags admin
// @Security ApiKeyAuth
// @Summary get user passkeys
// @Success 200 {array} entity_admin.RespPasskey "Successful response"
// @Router /user-passkeys/{uid} [get]
func getPasskeys(c *gin.Context) {
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

func getUsersStatis(c *gin.Context) {

	timeDimension := c.Query("time_dimension")
	dataType := c.Query("data_type")
	amount := c.Query("amount")

	if !entity_admin.IsValidTimeDimension(timeDimension) || !entity_admin.IsValidDataType(dataType) {
		api.Error(c, errors.New("time_dimension or data_type is invalid"))
		return
	}

	resp, total, err := admin.New().GetUsersStatis(c.Request.Context(), timeDimension, dataType, cast.ToInt(amount))
	if err != nil {
		api.Error(c, err)
		return
	}

	api.List(c, resp, total)
}

// @tags admin
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
