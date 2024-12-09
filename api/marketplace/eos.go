package marketplace

import (
	"exapp-go/api"
	"exapp-go/config"
	"exapp-go/internal/entity"

	pkeos "exapp-go/pkg/eos"

	"github.com/eoscanada/eos-go"

	"github.com/gin-gonic/gin"
)

// payCPU Pay CPU
// @Summary pay cpu
// @Description pay cpu for user tx
// @Tags eos
// @Accept json
// @Produce json
// @Param request body entity.ReqPayCPU true "signed tx"
// @Success 200 {object} entity.RespPayCPU "txid"
// @Router /eos/pay-cpu [post]
func payCPU(c *gin.Context) {
	request := entity.ReqPayCPU{}
	if err := c.ShouldBindJSON(&request); err != nil {
		api.Error(c, err)
		return
	}
	response, err := pkeos.SignAndBroadcastByPayer(
		c.Request.Context(),
		eos.New(config.Conf().EOS.NodeURL),
		request.SingleSignedTransaction,
		config.Conf().EOS.PayerPrivateKey,
	)
	if err != nil {
		api.Error(c, err)
		return
	}
	api.OK(c, entity.RespPayCPU{
		TransactionID: response.TransactionID,
	})
}
