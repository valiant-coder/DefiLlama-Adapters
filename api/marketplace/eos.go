package marketplace

import (
	"exapp-go/api"
	"exapp-go/config"
	"exapp-go/internal/entity"
	"exapp-go/internal/errno"
	"exapp-go/internal/services/marketplace"
	"strings"

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

	if request.PublicKey != "" {
		userService := marketplace.NewUserService()
		if err := userService.UpdateUserCredentialUsage(c.Request.Context(), request.PublicKey, c.ClientIP()); err != nil {
			api.Error(c, err)
			return
		}
	}

	response, err := pkeos.SignAndBroadcastByPayer(
		c.Request.Context(),
		eos.New(config.Conf().Eos.NodeURL),
		request.SingleSignedTransaction,
		config.Conf().Eos.PayerPrivateKey,
	)
	if err != nil {
		errString := err.Error()
		errSplit := strings.Split(errString, ":")
		if len(errSplit) > 1 {
			isAssertionFailure := false
			for i, v := range errSplit {
				if strings.Contains(v, "assertion failure with message") {
					isAssertionFailure = true
					err = errno.DefaultParamsError(errSplit[i+1])
					break
				}
			}
			if !isAssertionFailure {
				err = errno.DefaultParamsError(errSplit[1])
			}
		} else {
			err = errno.DefaultParamsError(errString)
		}
		api.Error(c, err)
		return
	}
	api.OK(c, entity.RespPayCPU{
		TransactionID: response.TransactionID,
	})
}
