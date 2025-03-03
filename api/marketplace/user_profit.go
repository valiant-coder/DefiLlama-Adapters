package marketplace

import (
	"errors"
	"exapp-go/api"
	"exapp-go/internal/services/marketplace"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

// @Summary Get user daily profit ranking
// @Description Get top 20 users profit ranking for specified date, and current user's ranking and profit
// @Tags user profit
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param timestamp query int true "UTC0 timezone date start timestamp"
// @Success 200 {object} entity.UserProfitRank
// @Router /profit/day-ranking [get]
func getDayProfitRanking(c *gin.Context) {
	dayTimestamp := c.Query("timestamp")
	if dayTimestamp == "" {
		api.Error(c, errors.New("timestamp is empty"))
		return
	}
	dayTime := time.Unix(cast.ToInt64(dayTimestamp), 0)
	uid := c.GetString("uid")
	userProfitService := marketplace.NewUserProfitService()
	result, err := userProfitService.GetDayProfitRanking(c.Request.Context(), dayTime, uid)
	if err != nil {
		api.Error(c, err)
		return
	}

	api.OK(c, result)
}

// @Summary Get user accumulated profit ranking
// @Description Get top 20 users accumulated profit ranking for specified time range, and current user's ranking and profit
// @Tags user profit
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param begin query int true "begin timestamp"
// @Param end query int true "end timestamp"
// @Success 200 {object} entity.UserProfitRank
// @Router /profit/accumulated-ranking [get]
func getAccumulatedProfitRanking(c *gin.Context) {
	beginTimestamp := c.Query("begin")
	endTimestamp := c.Query("end")
	if beginTimestamp == "" || endTimestamp == "" {
		api.Error(c, errors.New("begin or end timestamp is empty"))
		return
	}

	uid := c.GetString("uid")
	beginTime := time.Unix(cast.ToInt64(beginTimestamp), 0)
	endTime := time.Unix(cast.ToInt64(endTimestamp), 0)

	userProfitService := marketplace.NewUserProfitService()
	result, err := userProfitService.GetAccumulatedProfitRanking(c.Request.Context(), beginTime, endTime, uid)
	if err != nil {
		api.Error(c, err)
		return
	}

	api.OK(c, result)
}
