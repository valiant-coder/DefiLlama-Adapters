package marketplace

import (
	"exapp-go/api"

	"exapp-go/internal/entity"

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
	}
	api.OK(c, sysInfo)
}
