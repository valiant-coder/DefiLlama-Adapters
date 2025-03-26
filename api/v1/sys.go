package v1

import (
	"exapp-go/api"
	"time"

	"exapp-go/internal/entity"

	"exapp-go/config"

	"github.com/gin-gonic/gin"
)

// @Summary Ping
// @Description Ping
// @Tags system
// @Accept json
// @Produce json
// @Success 200
// @Router /api/v1/ping [get]
func ping(c *gin.Context) {
	api.OK(c, map[string]uint64{"timestamp": uint64(time.Now().Unix())})
}

// @Summary Get system information
// @Description Get system information
// @Tags system
// @Accept json
// @Produce json
// @Success 200 {object} entity.RespSystemInfo
// @Router /api/v1/system-info [get]
func getSystemInfo(c *gin.Context) {
	sysInfo := entity.RespSystemInfo{
		Version:       "1.0.0",
		PayEOSAccount: config.Conf().Eos.PayerAccount,
		TokenContract: config.Conf().Eos.OneDex.TokenContract,
		AppContract:   config.Conf().Eos.CdexConfig.AppContract,
	}
	api.OK(c, sysInfo)
}
