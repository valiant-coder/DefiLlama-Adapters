package admin

import (
	"exapp-go/api"
	"exapp-go/config"
	entity_admin "exapp-go/internal/entity/admin"
	"exapp-go/internal/services/admin"
	"time"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
)

var (
	identityKey = "admin"
)

func jwtConfig() *jwt.GinJWTMiddleware {
	conf := config.Conf()
	return &jwt.GinJWTMiddleware{
		Realm: conf.JWT.Realm,
		Key:   []byte(conf.JWT.SecretKey),
		Timeout:         24 * time.Hour,
		MaxRefresh:      24 * time.Hour,
		IdentityKey:     identityKey,
		PayloadFunc:     payloadFunc(),
		IdentityHandler: identityHandler(),
		Authenticator:   authenticator(),
		Unauthorized:    unauthorized(),
		Authorizator:    authorizator(),
		LoginResponse:   loginResponse(),
		TokenHeadName:   "Bearer",
	}
}

func unauthorized() func(c *gin.Context, code int, message string) {
	return func(c *gin.Context, code int, message string) {
		api.Unauthorized(c, message)
	}
}

func loginResponse() func(c *gin.Context, code int, token string, expire time.Time) {
	return func(c *gin.Context, code int, token string, expire time.Time) {
		api.OK(c, gin.H{
			"token":      token,
			"expired_at": expire.Unix(),
		})
	}
}

func payloadFunc() func(data interface{}) jwt.MapClaims {
	return func(data interface{}) jwt.MapClaims {
		if v, ok := data.(string); ok {
			return jwt.MapClaims{
				identityKey: v,
			}
		}
		return jwt.MapClaims{}
	}
}

func identityHandler() func(c *gin.Context) interface{} {
	return func(c *gin.Context) interface{} {
		claims := jwt.ExtractClaims(c)
		return claims[identityKey]
	}
}

func authenticator() func(c *gin.Context) (interface{}, error) {
	return func(c *gin.Context) (interface{}, error) {
		var reqAuth entity_admin.ReqAuth
		if err := c.ShouldBind(&reqAuth); err != nil {
			return "", err
		}

		ctx := c.Request.Context()
		adminName, err := admin.New().Auth(ctx, &reqAuth)
		if err != nil {
			return "", err
		}
		return adminName, nil
	}
}

func authorizator() func(data interface{}, c *gin.Context) bool {
	return func(data interface{}, c *gin.Context) bool {
		v, ok := data.(string)
		if ok {
			isExist, err := admin.New().CheckAdminIsExist(c.Request.Context(), v)
			if err != nil {
				return false
			}
			return isExist
		}
		return false
	}
}
