package v1

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

// @Summary Get token info
// @Description Get token info
// @Tags token
// @Accept json
// @Produce json
// @Param symbol path string true "coin symbol,ps BTC"
// @Success 200 {object} entity.Token "token info"
// @Router /token/{symbol} [get]
func getToken(c *gin.Context) {
	symbol := c.Param("symbol")
	tokenService := marketplace.NewTokenService()
	resp, err := tokenService.GetToken(c.Request.Context(), symbol)
	if err != nil {
		api.Error(c, err)
		return
	}
	api.OK(c, resp)
}
