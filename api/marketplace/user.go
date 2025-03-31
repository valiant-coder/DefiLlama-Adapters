package marketplace

import (
	"errors"
	"exapp-go/api"
	"exapp-go/internal/entity"
	"exapp-go/internal/services/marketplace"

	"github.com/gin-gonic/gin"
)

// @Summary Login
// @Description Login
// @Tags user
// @Accept json
// @Produce json
// @Param req body entity.ReqUserLogin true "login params"
// @Success 200
// @Router /login [post]
func authenticator(c *gin.Context) (interface{}, error) {
	var req entity.ReqUserLogin
	if err := c.ShouldBind(&req); err != nil {
		return nil, err
	}
	userService := marketplace.NewUserService()
	uid, err := userService.Login(c.Request.Context(), req)
	if err != nil {
		return nil, err
	}
	return uid, nil
}

func authorizator(data interface{}, c *gin.Context) bool {
	uid := data.(string)
	userService := marketplace.NewUserService()
	exist, err := userService.IsUserExist(c.Request.Context(), uid)
	if err != nil {
		return false
	}
	if !exist {
		return false
	}
	return true
}

// @Summary Create user credentials
// @Description Create user credentials
// @Security ApiKeyAuth
// @Tags user
// @Accept json
// @Produce json
// @Param req body entity.UserCredential true "create user credential params"
// @Success 200
// @Router /credentials [post]
func createUserCredentials(c *gin.Context) {
	var req entity.UserCredential
	if err := c.ShouldBind(&req); err != nil {
		api.Error(c, err)
		return
	}
	userService := marketplace.NewUserService()
	if err := userService.CreateUserCredential(c.Request.Context(), req, c.GetString("uid")); err != nil {
		api.Error(c, err)
		return
	}
	api.OK(c, nil)
}

// @Summary Get user credentials
// @Description Get user credentials
// @Security ApiKeyAuth
// @Tags user
// @Accept json
// @Produce json
// @Success 200 {array} entity.RespUserCredential "user credentials"
// @Router /credentials [get]
func getUserCredentials(c *gin.Context) {
	userService := marketplace.NewUserService()
	credentials, err := userService.GetUserCredentials(c.Request.Context(), c.GetString("uid"))
	if err != nil {
		api.Error(c, err)
		return
	}
	api.OK(c, credentials)
}

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
// @Success 200 {array} entity.UserBalance "user balances"
// @Router /balances [get]
func getUserBalances(c *gin.Context) {
	userService := marketplace.NewUserService()
	balances, err := userService.FetchUserBalanceByUID(c.Request.Context(), c.GetString("uid"))
	if err != nil {
		api.Error(c, err)
		return
	}
	api.OK(c, balances)
}

// @Summary Delete user credential
// @Description Delete user credential by credential_id
// @Security ApiKeyAuth
// @Tags user
// @Accept json
// @Produce json
// @Param credential_id path string true "credential id"
// @Success 200
// @Router /credentials/{credential_id} [delete]
func deleteUserCredential(c *gin.Context) {
	credentialID := c.Param("credential_id")
	if credentialID == "" {
		api.Error(c, errors.New("credential_id is required"))
		return
	}
	userService := marketplace.NewUserService()
	if err := userService.DeleteUserCredential(c.Request.Context(), c.GetString("uid"), credentialID); err != nil {
		api.Error(c, err)
		return
	}
	api.OK(c, nil)
}
