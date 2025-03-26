package marketplace

import (
	"exapp-go/api"
	"exapp-go/data"
	"exapp-go/internal/services/marketplace"

	"github.com/gin-gonic/gin"
)

// @Summary 获取邀请信息
// @Description Get invitation info
// @Tags user-invitation
// @Accept json
// @Produce json
// @Router /user/invitation [get]
func getInvitationInfo(c *gin.Context) {

	service := marketplace.NewUserInvitationService()
	userInvitation, err := service.GetUserInvitation(c.Request.Context(), c.GetString("uid"))
	if err != nil {
		api.Error(c, err)
		return
	}

	api.OK(c, userInvitation)
}

// @Summary 获取邀请用户列表
// @Description Get invite users
// @Tags user-invitation
// @Accept json
// @Produce json
// @Param request body data.UserInvitationListParam true "request"
// @Router /user/invites [get]
func getInviteUsers(c *gin.Context) {

	var param data.UserInvitationListParam
	if err := c.ShouldBindJSON(&param); err != nil {
		api.Error(c, err)
		return
	}

	service := marketplace.NewUserInvitationService()
	result, err := service.GetInviteUsers(c.Request.Context(), param)
	if err != nil {
		api.Error(c, err)
		return
	}

	api.List(c, result.Array, result.Total)
}

// @Summary 获取邀请链接列表
// @Description Get invitation links
// @Tags user-invitation
// @Accept json
// @Produce json
// @Param request body data.UILinkListParam true "request"
// @Success 200 {object} api.Response "invitation links"
// @Router /user/invitation/links [get]
func getInvitationLinks(c *gin.Context) {

	var param data.UILinkListParam
	if err := c.ShouldBindJSON(&param); err != nil {
		api.Error(c, err)
		return
	}

	service := marketplace.NewUserInvitationService()
	result, err := service.GetUserInvitationLinks(c.Request.Context(), param)
	if err != nil {
		api.Error(c, err)
		return
	}

	api.List(c, result.Array, result.Total)
}

// @Summary 创建邀请链接
// @Description Create invitation link
// @Tags user-invitation
// @Accept json
// @Produce json
// @Param req body data.UILinkParam true "create invitation link params"
// @Success 200 {object} api.Response "invitation link"
// @Router /user/invitation/link [post]
func createInvitationLink(c *gin.Context) {

	var param data.UILinkParam
	if err := c.ShouldBindJSON(&param); err != nil {
		api.Error(c, err)
		return
	}

	service := marketplace.NewUserInvitationService()
	err := service.CreateUILink(c.Request.Context(), c.GetString("uid"), &param)
	if err != nil {
		api.Error(c, err)
		return
	}

	api.OK(c, "success")
}

// @Summary 删除邀请链接
// @Description Delete invitation link
// @Tags user-invitation
// @Accept json
// @Produce json
// @Param link_id path string true "link id"
// @Success 200
// @Router /user/invitation/link/{link_id} [delete]
func deleteInvitationLink(c *gin.Context) {

	service := marketplace.NewUserInvitationService()
	err := service.DeleteInvitationLink(c.Request.Context(), c.Param("code"))
	if err != nil {

		api.Error(c, err)
		return
	}

	api.OK(c, "success")
}

// @Summary 获取邀请链接信息
// @Description Get invitation link by code
// @Tags user-invitation
// @Accept json
// @Produce json
// @Param code path string true "code"
func getInvitationLinkByCode(c *gin.Context) {

	service := marketplace.NewUserInvitationService()
	link, err := service.GetInvitationLinkByCode(c.Request.Context(), c.Param("code"))
	if err != nil {
		api.Error(c, err)
		return
	}

	api.OK(c, link)
}
