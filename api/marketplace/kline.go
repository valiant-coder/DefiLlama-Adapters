package marketplace

import (
	"exapp-go/api"
	"exapp-go/internal/services/marketplace"
	"time"

	"github.com/gin-gonic/gin"
)

type getKlineReq struct {
	PoolID   uint64 `form:"pool_id"`
	Interval string `form:"interval" binding:"required"`
	Start    int64  `form:"start" binding:"required"`
	End      int64  `form:"end" binding:"required"`
}

// @Summary Get kline data
// @Description Get kline data by pool id and interval
// @Tags kline
// @Accept json
// @Produce json
// @Param pool_id query string true "pool id"
// @Param interval query string true "interval" Enums(1m,5m,15m,30m,1h,4h,1d,1w,1M)
// @Param start query integer false "start timestamp"
// @Param end query integer false "end timestamp"
// @Success 200 {array} entity.Kline "kline data"
// @Router /api/v1/klines [get]
func klines(c *gin.Context) {
	var req getKlineReq
	if err := c.ShouldBindQuery(&req); err != nil {
		api.Error(c, err)
		return
	}

	start := time.Unix(req.Start, 0)
	end := time.Unix(req.End, 0)
	klineService := marketplace.NewKlineService()
	klines, err := klineService.GetKline(c.Request.Context(), req.PoolID, req.Interval, start, end)
	if err != nil {
		api.Error(c, err)
		return
	}

	api.OK(c, klines)

}
