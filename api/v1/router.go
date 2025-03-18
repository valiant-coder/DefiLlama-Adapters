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

	_ "exapp-go/docs/v1"
)

// @title exapp api v1
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
			config.InstanceName = "api_v1"
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
		api.Trace("api-v1"),
		gin.CustomRecovery(handleRecovery),
	)

	v1 := r.Group("/api/v1")

	v1.GET("/ping", ping)
	v1.GET("/pools", pools)
	v1.GET("/pools/:symbolOrId", getPoolDetail)
	v1.GET("/klines", klines)
	v1.GET("/depth", getDepth)
	v1.GET("/latest-trades", getLatestTrades)

	v1.GET("/system-info", getSystemInfo)

	v1.GET("/tokens", getSupportTokens)
	v1.GET("/token/:symbol", getToken)

	// Protected routes requiring API key authentication
	auth := v1.Group("/")
	auth.Use(AuthMiddleware())
	{
		auth.POST("/tx", sendTx)
		auth.GET("/info", getUserInfo)
		auth.GET("/open-orders", getOpenOrders)
		auth.GET("/history-orders", getHistoryOrders)
		auth.GET("/orders/:id", getOrderDetail)
		auth.GET("/balances", getUserBalances)
	}

	if config.Conf().HTTPS.Enabled {
		return r.RunTLS(addr,
			config.Conf().HTTPS.CertFile,
			config.Conf().HTTPS.KeyFile,
		)
	}
	return r.Run(addr)
}
