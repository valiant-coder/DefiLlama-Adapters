package marketplace

import (
	"exapp-go/api"
	"exapp-go/pkg/log"
	"fmt"
	"os"
	"time"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title exapp-go marketplace api
// @version 1.0
// @host 127.0.0.1:8083
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
			// marketplace.SwaggerInfomarketplace.Host = swaggerHost
		}
	}

	corsConfig := cors.DefaultConfig()
	corsConfig.AllowAllOrigins = true
	corsConfig.AllowCredentials = true
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

	authMiddleware, err := jwt.New(api.InitParams())
	if err != nil {
		log.Logger().Errorf("[RegisterJwtMiddleWare] %s", err)
	}

	// register middleware
	r.Use(api.HandlerMiddleWare(authMiddleware))

	r.POST("/login", authMiddleware.LoginHandler)
	auth := r.Group("/", authMiddleware.MiddlewareFunc())
	auth.GET("/refresh_token", authMiddleware.RefreshHandler)

	return r.Run(addr)

}
