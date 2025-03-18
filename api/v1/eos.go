package v1

import (
	"exapp-go/api"
	"exapp-go/config"
	"exapp-go/internal/entity"
	"exapp-go/internal/errno"
	"strings"

	pkeos "exapp-go/pkg/eos"

	"github.com/eoscanada/eos-go"

	"github.com/gin-gonic/gin"
)

// @Summary send tx
// @Description send  tx
// @Tags tx
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body entity.ReqSendTx true "signed tx"
// @Success 200 {object} entity.RespSendTx "txid"
// @Router /api/v1/tx [post]
func sendTx(c *gin.Context) {
	request := entity.ReqSendTx{}
	if err := c.ShouldBindJSON(&request); err != nil {
		api.Error(c, err)
		return
	}

	subAccount := GetSubAccountFromContext(c)
	isLegit := pkeos.CheckIsLegitTransaction(request.SingleSignedTransaction, config.Conf().Eos.PayerAccount, subAccount.EOSAccount, subAccount.Permission)
	if !isLegit {
		api.Error(c, errno.DefaultParamsError("transaction is not legit"))
		return
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
	api.OK(c, entity.RespSendTx{
		TransactionID: response.TransactionID,
	})
}
