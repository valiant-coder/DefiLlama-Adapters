package marketplace

import (
	"exapp-go/api"

	"exapp-go/internal/services/marketplace"

	"github.com/gin-gonic/gin"
)

// @Summary Get support tokens
// @Description Get support tokens
// @Tags token
// @Accept json
// @Produce json
// @Success 200 {array} entity.Token "token list"
// @Router /support-tokens [get]
func getSupportTokens(c *gin.Context) {
	tokenService := marketplace.NewTokenService()
	resp, err := tokenService.GetSupportTokens(c.Request.Context())
	if err != nil {
		api.Error(c, err)
		return
	}
	api.OK(c, resp)
}
