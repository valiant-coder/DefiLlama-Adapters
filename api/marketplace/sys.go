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
		EVMAgentContract: config.Conf().Eos.OneDex.EVMAgentContract,
		VaultEVMAddress:  config.Conf().Eos.Exsat.BridgeExtensionEVMAddress,
		VaultEOSAddress:  config.Conf().Eos.OneDex.PortalContract,
		TokenContract:    config.Conf().Eos.OneDex.TokenContract,
		AppContract:      config.Conf().Eos.CdexConfig.AppContract,
		ExsatNetwork: entity.ExsatNetwork{
			CurrencySymbol:   config.Conf().Evm.ExsatNetwork.CurrencySymbol,
			NetworkUrl:       config.Conf().Evm.ExsatNetwork.NetworkUrl,
			ChainId:          config.Conf().Evm.ExsatNetwork.ChainId,
			NetworkName:      config.Conf().Evm.ExsatNetwork.NetworkName,
			BlockExplorerUrl: config.Conf().Evm.ExsatNetwork.BlockExplorerUrl,
		},

		TradingCompetition: entity.TradingCompetition{
			BeginTime:         entity.Time(config.Conf().TradingCompetition.BeginTime),
			EndTime:           entity.Time(config.Conf().TradingCompetition.EndTime),
			DailyPoints:       config.Conf().TradingCompetition.DailyPoints,
			AccumulatedPoints: config.Conf().TradingCompetition.AccumulatedPoints,
		},
		BTCBridgeContract:  config.Conf().Eos.Exsat.BTCBridgeContract,
		BridgeContract:     config.Conf().Eos.Exsat.BridgeContract,
		DepositRecipientAd: config.Conf().Eos.Exsat.BridgeExtensionEVMAddress,
		MakerAgentContract: config.Conf().Eos.OneDex.MakerAgentContract,
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
