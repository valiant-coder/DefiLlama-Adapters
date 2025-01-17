package marketplace

import (
	"context"

	"exapp-go/api"
	repair "exapp-go/internal/services/repair"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

func repairPool(c *gin.Context) {
	ctx := context.Background()
	server := repair.NewRepairServer()
	err := server.RepairPool(ctx, cast.ToUint64(c.Query("pool_id")))
	if err != nil {
		api.Error(c, err)
		return
	}
	api.NoContent(c)
}
