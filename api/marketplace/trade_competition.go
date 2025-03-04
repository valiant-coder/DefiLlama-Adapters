package marketplace

import (
	"errors"
	"exapp-go/api"
	"exapp-go/config"
	"exapp-go/internal/errno"
	"exapp-go/internal/services/marketplace"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

// @Summary Get user daily profit ranking
// @Description Get top 20 users profit ranking for specified date, and current user's ranking and profit
// @Tags trade competition
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param timestamp query int true "UTC0 timezone date start timestamp"
// @Success 200 {object} entity.UserProfitRank
// @Router /trade-competition/day-ranking [get]
func getDayProfitRanking(c *gin.Context) {
	claims, _ := authMiddleware.GetClaimsFromJWT(c)
	var uid string
	if claims["uid"] != nil {
		uid = claims["uid"].(string)
	}

	dayTimestamp := c.Query("timestamp")
	if dayTimestamp == "" {
		api.Error(c, errors.New("timestamp is empty"))
		return
	}
	dayTime := time.Unix(cast.ToInt64(dayTimestamp), 0)
	if dayTime.Before(config.Conf().TradingCompetition.BeginTime) {
		api.Error(c, errno.DefaultParamsError("timestamp is before trading competition begin time"))
		return
	}
	if dayTime.After(config.Conf().TradingCompetition.EndTime) {
		api.Error(c, errno.DefaultParamsError("timestamp is after trading competition end time"))
		return
	}
	tradeCompetitionService := marketplace.NewTradeCompetitionService()
	result, err := tradeCompetitionService.GetDayProfitRanking(c.Request.Context(), dayTime, uid)
	if err != nil {
		api.Error(c, err)
		return
	}

	api.OK(c, result)
}

// @Summary Get user accumulated profit ranking
// @Description Get top 20 users accumulated profit ranking for specified time range, and current user's ranking and profit
// @Tags trade competition
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param begin query int true "begin timestamp"
// @Param end query int true "end timestamp"
// @Success 200 {object} entity.UserProfitRank
// @Router /trade-competition/accumulated-ranking [get]
func getAccumulatedProfitRanking(c *gin.Context) {
	claims, _ := authMiddleware.GetClaimsFromJWT(c)
	var uid string
	if claims["uid"] != nil {
		uid = claims["uid"].(string)
	}

	beginTimestamp := c.Query("begin")
	endTimestamp := c.Query("end")
	var beginTime, endTime time.Time
	if beginTimestamp == "" || endTimestamp == "" {
		beginTime = config.Conf().TradingCompetition.BeginTime
		endTime = config.Conf().TradingCompetition.EndTime
	} else {
		beginTime = time.Unix(cast.ToInt64(beginTimestamp), 0)
		endTime = time.Unix(cast.ToInt64(endTimestamp), 0)
	}

	tradeCompetitionService := marketplace.NewTradeCompetitionService()
	result, err := tradeCompetitionService.GetAccumulatedProfitRanking(c.Request.Context(), beginTime, endTime, uid)
	if err != nil {
		api.Error(c, err)
		return
	}

	api.OK(c, result)
}

// @Summary Get total trade stats
// @Description Get total trade stats
// @Tags trade competition
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Success 200 {object} entity.TotalTradeStats
// @Router /trade-competition/total-trade-stats [get]
func getTotalTradeStats(c *gin.Context) {
	claims, _ := authMiddleware.GetClaimsFromJWT(c)
	var uid string
	if claims["uid"] != nil {
		uid = claims["uid"].(string)
	}
	tradeCompetitionService := marketplace.NewTradeCompetitionService()
	result, err := tradeCompetitionService.GetTotalTradeStats(c.Request.Context(), uid)
	if err != nil {
		api.Error(c, err)
		return
	}

	api.OK(c, result)
}
