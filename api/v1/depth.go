package v1

import (
	"exapp-go/api"
	"exapp-go/internal/services/marketplace"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

// @Summary Get depth
// @Description Get order book by pool id
// @Tags depth
// @Accept json
// @Produce json
// @Param pool_id query string true "pool_id"
// @Param precision query string false "0.00000001 ~ 10000"
// @Param limit query int false "limit"
// @Success 200 {object} entity.Depth "order depth"
// @Router /api/v1/depth [get]
func getDepth(c *gin.Context) {
	poolID := c.Query("pool_id")
	precision := c.Query("precision")
	limit := c.Query("limit")
	depth, err := marketplace.NewDepthService().GetDepth(c.Request.Context(), cast.ToUint64(poolID), precision, cast.ToInt(limit))
	if err != nil {
		api.Error(c, err)
		return
	}
	api.OK(c, depth)
}
