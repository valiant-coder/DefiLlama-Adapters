package marketplace

import (
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
