package marketplace

import (
	"exapp-go/api"
	"exapp-go/internal/services/marketplace"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

// @Summary Get latest trades
// @Description Get latest trades
// @Security ApiKeyAuth
// @Tags trade
// @Accept json
// @Produce json
// @Param pool_id query string false "pool_id"
// @Param limit query integer false "limit count"
// @Success 200 {array} entity.Trade "trade list"
// @Router /api/v1/latest-trades [get]
func getLatestTrades(c *gin.Context) {

	poolID := cast.ToUint64(c.Query("pool_id"))
	limit := cast.ToInt(c.Query("limit"))
	if limit == 0 {
		limit = 10
	}

	tradeService := marketplace.NewTradeService()
	trades, err := tradeService.GetLatestTrades(c.Request.Context(), poolID, limit)
	if err != nil {
		api.Error(c, err)
		return
	}
	api.OK(c, trades)

}
