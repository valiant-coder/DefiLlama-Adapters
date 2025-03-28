package admin

import (
	"exapp-go/api"
	"exapp-go/pkg/log"
	"fmt"
	"os"
	"time"

	admin "exapp-go/docs/admin"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/casbin/casbin/v2"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

var authMiddleware *jwt.GinJWTMiddleware

func registerJwtMiddleware() {

	var err error
	authMiddleware, err = jwt.New(jwtConfig())
	if err != nil {
		log.Logger().Errorf("[RegisterJwtMiddleWare] %s", err)
	}
}

// @title exapp-go admin api
// @version 1.0
// @host 127.0.0.1:8084
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
			config.InstanceName = "admin"
		}))
		swaggerHost := os.Getenv("SWAGGER_HOST")
		if swaggerHost != "" {
			admin.SwaggerInfoadmin.Host = swaggerHost
		}
	}

	corsConfig := cors.DefaultConfig()
	corsConfig.AllowAllOrigins = true
	corsConfig.AllowMethods = []string{"*"}
	corsConfig.AllowCredentials = true
	corsConfig.AllowHeaders = []string{"*"}
	corsConfig.MaxAge = 12 * time.Hour
	handleRecovery := func(c *gin.Context, err interface{}) {
		log.Logger().Errorf("[Recovery] %s", err)
		api.Error(c, fmt.Errorf("%v", err))
	}

	r.Use(
		cors.New(corsConfig),
		api.Logger(),
		api.Trace("exapp-go-admin"),
		gin.CustomRecovery(handleRecovery),
	)

	registerJwtMiddleware()

	r.POST("/login", login)
	r.POST("/auth", authMiddleware.LoginHandler)
	r.POST("/auth/reset_password", resetPassword)
	r.GET("/auth/google_auth_secret/:name", getGoogleAuthSecret)

	authorized := r.Group("/")

	enforcer, err := casbin.NewSyncedEnforcer("config/rbac_model.conf", &CasbinAdapter{})
	if err != nil {
		return err
	}
	defer enforcer.StopAutoLoadPolicy()
	enforcer.StartAutoLoadPolicy(time.Second * 10)
	authorized.Use(
		authMiddleware.MiddlewareFunc(),
	)
	authorized.GET("/admin/:name", getAdmin)
	authorized.Use(Authorized(enforcer))
	// accounts
	authorized.GET("/admins", queryAdmins)
	authorized.PUT("/admin/:name", updateAdmin)
	authorized.POST("/admin", createAdmin)
	authorized.DELETE("/admin/:name", deleteAdmin)

	// roles
	authorized.GET("/admin_roles", queryAdminRoles)
	authorized.GET("/admin_role/:id", getAdminRole)
	authorized.POST("/admin_role", createAdminRole)
	authorized.PUT("/admin_role/:id", updateAdminRole)

	// permission groups
	authorized.GET("/admin_permission_groups", getAdminPermissionGroups)
	authorized.GET("/admin_permission_group/:id", getAdminPermissionGroup)
	authorized.POST("/admin_permission_group", createAdminPermissionGroup)
	authorized.PUT("/admin_permission_group/:id", updateAdminPermissionGroup)

	// individual permissions
	authorized.GET("/admin_permissions", getAdminPermissions)
	authorized.GET("/admin_permission/:id", getAdminPermission)
	authorized.POST("/admin_permission", createAdminPermission)
	authorized.PUT("/admin_permission/:id", updateAdminPermission)

	// user
	authorized.GET("/users", queryUsers)
	authorized.GET("/user_passkeys/:uid", getUserPasskeys)
	authorized.GET("/users_statis", getUsersStatis)

	// transactions
	authorized.GET("/transactions_records", getTransactionsRecord)
	authorized.GET("/deposit_amount_total", getDepositAmountTotal)
	authorized.GET("/withdraw_amount_total", getWithdrawAmountTotal)

	// token
	authorized.GET("/tokens", queryTokens)
	authorized.POST("/token", createToken)
	authorized.PUT("/token/:id", updateToken)

	// pool
	authorized.GET("/pools", queryPools)
	authorized.PUT("/pool/:pool_id", updatePool)
	authorized.POST("/pool", createPool)

	// open_order
	authorized.GET("/history_orders", queryHistoryQrders)
	authorized.GET("/orders_coin_total", getOrdersCoinTotal)

	return r.Run(addr)

}
