package admin

import (
	"exapp-go/api"
	entity_admin "exapp-go/internal/entity/admin"
	"exapp-go/internal/services/admin"
	"exapp-go/pkg/queryparams"

	"github.com/gin-gonic/gin"
)

// @Tags admin
// @Security ApiKeyAuth
// @Summary query tokens
// @Accept json
// @Produce json
// @Param symbol query string false "symbol"
// @Param evm_contract_address query string false "evm_contract_address"
// @Param chains query string false "chains"
// @Param limit query int false "limit"
// @Param offset query int false "offset"
// @Success 200 {array} entity_admin.RespToken "Successful response"
// @Router /tokens [get]

func queryTokens(c *gin.Context) {

	resp, total, err := admin.New().QueryTokens(c.Request.Context(), queryparams.NewQueryParams(c))
	if err != nil {
		api.Error(c, err)
		return
	}

	api.List(c, resp, total)
}

// @tags token
// @Security ApiKeyAuth
// @summary create token
// @Param body body entity_admin.ReqCreateToken false "request body"
// @Success 200 {object} entity_admin.RespToken "Successful response"
// @Router /token [post]
func createToken(c *gin.Context) {

	var req entity_admin.ReqCreateToken
	if err := c.ShouldBind(&req); err != nil {
		api.Error(c, err)
		return
	}

	token, err := admin.New().CreateToken(c.Request.Context(), &req)
	if err != nil {
		api.Error(c, err)
		return
	}
	api.OK(c, token)
}

// @tags token
// @Security ApiKeyAuth
// @Summary update token
// @Param id path string true "id"
// @Param body body entity_admin.ReqUpdateToken false "request body"
// @Success 200 {object} entity_admin.RespToken "Successful response"
// @Router /token/{id} [put]

func updateToken(c *gin.Context) {
	var req entity_admin.ReqUpdateToken
	if err := c.ShouldBind(&req); err != nil {
		api.Error(c, err)
		return
	}

	id := c.GetUint("id")
	token, err := admin.New().UpdateToken(c.Request.Context(), &req, id)
	if err != nil {
		api.Error(c, err)
		return
	}
	api.OK(c, token)
}
