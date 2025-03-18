package marketplace

import (
	"errors"
	"exapp-go/api"
	"exapp-go/internal/entity"
	"exapp-go/internal/services/marketplace"

	"github.com/gin-gonic/gin"
)

// @Summary Add sub-account
// @Description Add a new sub-account for the user
// @Security ApiKeyAuth
// @Tags user
// @Accept json
// @Produce json
// @Param req body entity.ReqAddSubAccount true "add sub-account params"
// @Success 200 {object} entity.RespAddSubAccount
// @Router /sub-accounts [post]
func addSubAccount(c *gin.Context) {
	var req entity.ReqAddSubAccount
	if err := c.ShouldBind(&req); err != nil {
		api.Error(c, err)
		return
	}

	uid := c.GetString("uid")
	if uid == "" {
		api.Error(c, errors.New("unauthorized"))
		return
	}

	userService := marketplace.NewUserService()
	resp, err := userService.AddSubAccount(c.Request.Context(), uid, req)
	if err != nil {
		api.Error(c, err)
		return
	}

	api.OK(c, resp)
}

// @Summary Get sub-accounts
// @Description Get all sub-accounts for the user
// @Security ApiKeyAuth
// @Tags user
// @Accept json
// @Produce json
// @Success 200 {object} entity.RespGetSubAccounts
// @Router /sub-accounts [get]
func getSubAccounts(c *gin.Context) {
	uid := c.GetString("uid")
	if uid == "" {
		api.Error(c, errors.New("unauthorized"))
		return
	}

	userService := marketplace.NewUserService()
	resp, err := userService.GetSubAccounts(c.Request.Context(), uid)
	if err != nil {
		api.Error(c, err)
		return
	}

	api.OK(c, resp)
}

// @Summary Delete sub-account
// @Description Delete a sub-account by name
// @Security ApiKeyAuth
// @Tags user
// @Accept json
// @Produce json
// @Param req body entity.ReqDeleteSubAccount true "delete sub-account params"
// @Success 200 {object} entity.RespDeleteSubAccount
// @Router /sub-accounts [delete]
func deleteSubAccount(c *gin.Context) {
	var req entity.ReqDeleteSubAccount
	if err := c.ShouldBind(&req); err != nil {
		api.Error(c, err)
		return
	}

	uid := c.GetString("uid")
	if uid == "" {
		api.Error(c, errors.New("unauthorized"))
		return
	}

	userService := marketplace.NewUserService()
	resp, err := userService.DeleteSubAccount(c.Request.Context(), uid, req)
	if err != nil {
		api.Error(c, err)
		return
	}

	api.OK(c, resp)
}
