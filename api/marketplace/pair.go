package marketplace

import "github.com/gin-gonic/gin"

// @Summary Get trading pairs
// @Description Get all trading pairs
// @Tags pair
// @Accept json
// @Produce json
// @Success 200 {array} entity.Pair "pair list"
// @Router /pairs [get]
func getPairs(c *gin.Context) {

}

// @Summary Get trading pair detail
// @Description Get trading pair detail by id
// @Tags pair
// @Accept json
// @Produce json
// @Param pair_id path string true "pair_id"
// @Success 200 {object} entity.Pair "pair detail"
// @Router /pairs/{pair_id} [get]
func getPair(c *gin.Context) {

} 