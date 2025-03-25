package marketplace

import (
	"exapp-go/api"
	"exapp-go/data"
	"exapp-go/internal/services/marketplace"
	
	"github.com/gin-gonic/gin"
)

// @Summary 获取积分信息
// @Description 获取积分信息
// @Tags user-points
// @Accept json
// @Produce json
// @Success 200 {object} api.Response "points info"
// @Router /user/points [get]
func getPointsInfo(c *gin.Context) {
	var param data.UPRecordListParam
	if err := c.ShouldBindJSON(&param); err != nil {
		api.Error(c, err)
		return
	}
	
	service := marketplace.NewUserPointsService()
	userPoints, err := service.GetUserPoints(c.Request.Context(), param.UID)
	if err != nil {
		api.Error(c, err)
		return
	}
	
	api.OK(c, userPoints)
}

// @Summary 获取积分记录
// @Description 获取积分记录
// @Tags user-points
// @Accept json
// @Produce json
// @Param request body data.UPRecordListParam true "request"
// @Success 200 {object} api.Response "points records"
// @Router /user/points/records [get]
func getPointsRecords(c *gin.Context) {
	var param data.UPRecordListParam
	if err := c.ShouldBindJSON(&param); err != nil {
		api.Error(c, err)
		return
	}
	
	service := marketplace.NewUserPointsService()
	result, err := service.GetUserPointsRecords(c.Request.Context(), param)
	if err != nil {
		api.Error(c, err)
		return
	}
	
	api.List(c, result.Array, result.Total)
}

// @Summary 获取积分配置
// @Description 获取积分配置
// @Tags user-points
// @Accept json
// @Produce json
// @Success 200 {object} api.Response "points conf"
// @Router /user/points/conf [get]
func getPointsConf(c *gin.Context) {
	service := marketplace.NewUserPointsService()
	conf, err := service.GetUserPointsConf(c.Request.Context())
	if err != nil {
		api.Error(c, err)
		return
	}
	
	api.OK(c, conf)
}

// @Summary 更新积分配置
// @Description 更新积分配置
// @Tags user-points
// @Accept json
// @Produce json
// @Param request body data.UserPointsConfParam true "request"
// @Success 200 {object} api.Response "points conf"
// @Router /user/points/conf [put]
func updatePointsConf(c *gin.Context) {
	var param data.UserPointsConfParam
	if err := c.ShouldBindJSON(&param); err != nil {
		api.Error(c, err)
		return
	}
	
	service := marketplace.NewUserPointsService()
	err := service.UpdateUserPointsConf(c.Request.Context(), &param)
	if err != nil {
		api.Error(c, err)
		return
	}
	
	api.OK(c, "success")
}
