package marketplace

import (
	"exapp-go/api"
	"exapp-go/config"
	"exapp-go/pkg/log"
	"fmt"
	"os"
	"time"

	"exapp-go/docs/marketplace"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

var authMiddleware *jwt.GinJWTMiddleware

func registerJwtMiddleware() {
	var err error
	jwtParams := api.InitParams()
	jwtParams.Authenticator = authenticator
	jwtParams.Authorizator = authorizator

	authMiddleware, err = jwt.New(jwtParams)
	if err != nil {
		log.Logger().Errorf("[RegisterJwtMiddleWare] %s", err)
	}
}

// @title exapp-go marketplace api
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
			config.InstanceName = "marketplace"
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

	r.GET("/system-info", getSystemInfo)
	r.POST("/eos/pay-cpu", payCPU)
	r.GET("/support-tokens", getSupportTokens)
	r.GET("/token/:symbol", getToken)
	r.GET("/repair-pool", repairPool)
	r.GET("/sys-trade-info", getSysTradeInfo)
	r.POST("/faucet", claimFaucet)
	r.GET("/trade-competition/day-ranking", getDayProfitRanking)
	r.GET("/trade-competition/accumulated-ranking", getAccumulatedProfitRanking)
	r.GET("/trade-competition/total-trade-stats", getTotalTradeStats)
	// register middleware
	registerJwtMiddleware()
	r.Use(api.HandlerMiddleWare(authMiddleware))
	r.POST("/login", authMiddleware.LoginHandler)
	auth := r.Group("/", authMiddleware.MiddlewareFunc())
	auth.GET("/refresh_token", authMiddleware.RefreshHandler)

	auth.POST("/credentials", createUserCredentials)
	auth.GET("/credentials", getUserCredentials)
	auth.DELETE("/credentials/:credential_id", deleteUserCredential)

	auth.GET("/user-info", getUserInfo)

	// orders
	auth.GET("/open-orders", getOpenOrders)
	auth.GET("/history-orders", getHistoryOrders)
	auth.GET("/orders/:id", getOrderDetail)
	auth.GET("/unread-orders", checkUnreadOrders)
	auth.POST("/orders/clear-unread", clearAllUnreadOrders)

	// user balance
	auth.GET("/balances", getUserBalances)

	// sub-account routes
	auth.POST("/sub-accounts", addSubAccount)
	auth.GET("/sub-accounts", getSubAccounts)
	auth.DELETE("/sub-accounts", deleteSubAccount)

	auth.POST("/first-deposit", firstDeposit)
	auth.POST("/deposit", deposit)

	auth.GET("/deposit-history", getDepositHistory)
	auth.GET("/withdrawal-history", getWithdrawalHistory)

	// User Invitation
	auth.GET("/user/invitation", getInvitationInfo)
	auth.GET("/user/invites", getInviteUsers)
	auth.GET("/user/invitation/links", getInvitationLinks)
	auth.POST("/user/invitation/link", createInvitationLink)
	auth.GET("/user/invitation/link/:code", getInvitationLinkByCode)
	auth.DELETE("/user/invitation/link/:code", deleteInvitationLink)

	// User Points
	auth.GET("/user/points", getPointsInfo)
	auth.GET("/user/points/records", getPointsRecords)
	auth.GET("/user/points/conf", getPointsConf)
	auth.PUT("/user/points/conf", updatePointsConf)

	if config.Conf().HTTPS.Enabled {
		return r.RunTLS(addr,
			config.Conf().HTTPS.CertFile,
			config.Conf().HTTPS.KeyFile,
		)
	}
	return r.Run(addr)
}
