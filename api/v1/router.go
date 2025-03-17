package v1

import (
	"exapp-go/api"
	"exapp-go/config"
	"exapp-go/pkg/log"
	"fmt"
	"os"
	"time"

	"exapp-go/docs/marketplace"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title onedex v1 api
// @version 1.0
// @host 127.0.0.1:8080
// @BasePath /
// @schemes http https
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
func Run(addr string, release bool) error {
	fmt.Printf("run port: %s, release: %v\n", addr, release)
	if release {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()

	// ginSwagger
	if !release {
		r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler, func(config *ginSwagger.Config) {
			config.InstanceName = "v1-api"
		}))
		swaggerHost := os.Getenv("SWAGGER_HOST")
		if swaggerHost != "" {
			marketplace.SwaggerInfomarketplace.Host = swaggerHost
		}
	}

	corsConfig := cors.DefaultConfig()
	corsConfig.AllowAllOrigins = true
	corsConfig.AllowCredentials = true
	corsConfig.AllowHeaders = append(corsConfig.AllowHeaders, "Authorization")
	corsConfig.MaxAge = 12 * time.Hour
	handleRecovery := func(c *gin.Context, err interface{}) {
		log.Logger().Errorf("[Recovery] %s", err)
		api.Error(c, fmt.Errorf("%v", err))
	}

	r.Use(
		cors.New(corsConfig),
		api.Logger(),
		api.Trace("marketplace"),
		gin.CustomRecovery(handleRecovery),
	)

	v1 := r.Group("/api/v1")

	// Pool routes
	v1.GET("/pools", pools)
	v1.GET("/pools/:symbolOrId", getPoolDetail)

	v1.GET("/klines", klines)
	v1.GET("/depth", getDepth)
	v1.GET("/latest-trades", getLatestTrades)
	v1.GET("/open-orders", getOpenOrders)
	v1.GET("/history-orders", getHistoryOrders)
	v1.GET("/orders/:id", getOrderDetail)

	v1.GET("/balances", getUserBalances)

	r.GET("/system-info", getSystemInfo)
	r.POST("/eos/pay-cpu", payCPU)
	r.GET("/support-tokens", getSupportTokens)
	r.GET("/token/:symbol", getToken)
	// register middleware

	if config.Conf().HTTPS.Enabled {
		return r.RunTLS(addr,
			config.Conf().HTTPS.CertFile,
			config.Conf().HTTPS.KeyFile,
		)
	}
	return r.Run(addr)
}
