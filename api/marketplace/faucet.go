package marketplace

import (
	"exapp-go/api"
	"exapp-go/internal/entity"
	"exapp-go/internal/services/marketplace"

	"github.com/gin-gonic/gin"
)

// @Summary Claim faucet
// @Description Claim faucet
// @Tags faucet
// @Accept json
// @Produce json
// @Param request body entity.ReqClaimFaucet true "request"
// @Success 200 {object} entity.RespClaimFaucet "txid"
// @Router /faucet [post]
func claimFaucet(c *gin.Context) {
	request := entity.ReqClaimFaucet{}
	if err := c.ShouldBindJSON(&request); err != nil {
		api.Error(c, err)
		return
	}
	faucetService := marketplace.NewFaucetService()
	claimFaucet, err := faucetService.ClaimFaucet(c.Request.Context(), &request)
	if err != nil {
		api.Error(c, err)
		return
	}
	api.OK(c, claimFaucet)
}
