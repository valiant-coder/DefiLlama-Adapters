package marketplace

import (
	"exapp-go/api"

	"exapp-go/internal/entity"
	"exapp-go/internal/services/marketplace"

	"exapp-go/config"

	"github.com/gin-gonic/gin"
)

// @Summary Get system information
// @Description Get system information
// @Tags system
// @Accept json
// @Produce json
// @Success 200 {object} entity.SystemInfo
// @Router /system-info [get]
func getSystemInfo(c *gin.Context) {
	sysInfo := entity.SystemInfo{
		Version: "1.0.0",
		PayCPU: entity.PayCPU{
			Account: config.Conf().Eos.PayerAccount,
		},
		VaultEVMAddress: config.Conf().Eos.Exapp.VaultEVMAddress,
		VaultEOSAddress: config.Conf().Eos.Exapp.AssetContract,
		TokenContract:   config.Conf().Eos.Exapp.TokenContract,
	}
	api.OK(c, sysInfo)
}

// @Summary Get system trade information
// @Description Get system trade information
// @Tags system
// @Accept json
// @Produce json
// @Success 200 {object} entity.SysTradeInfo
// @Router /sys-trade-info [get]
func getSysTradeInfo(c *gin.Context) {
	tradeService := marketplace.NewTradeService()
	tradeInfo, err := tradeService.GetTradeCountAndVolume(c.Request.Context())
	if err != nil {
		api.Error(c, err)
		return
	}
	api.OK(c, tradeInfo)
}
