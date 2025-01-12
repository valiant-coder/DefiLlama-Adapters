package api

import (
	"log"
	"time"

	"exapp-go/config"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
)

var (
	identityKey = "uid"
)

func HandlerMiddleWare(authMiddleware *jwt.GinJWTMiddleware) gin.HandlerFunc {
	return func(context *gin.Context) {
		errInit := authMiddleware.MiddlewareInit()
		if errInit != nil {
			log.Fatal("authMiddleware.MiddlewareInit() Error:" + errInit.Error())
		}
	}
}

func InitParams() *jwt.GinJWTMiddleware {
	cfg := config.Conf()

	return &jwt.GinJWTMiddleware{
		Realm:       cfg.JWT.Realm,
		Key:         []byte(cfg.JWT.SecretKey),
		Timeout:     time.Duration(cfg.JWT.Timeout) * time.Hour,
		MaxRefresh:  time.Duration(cfg.JWT.Timeout) * time.Hour,
		IdentityKey: identityKey,
		PayloadFunc: payloadFunc(),

		IdentityHandler: identityHandler(),
		Authenticator:   authenticator(),
		Authorizator:    authorizator(),
		Unauthorized:    unauthorized(),
		LoginResponse:   loginResponse(),
		TokenLookup:     "header: Authorization, query: token, cookie: jwt",
		TokenHeadName:   "Bearer",
		TimeFunc:        time.Now,
	}
}

// payload set
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
		// var loginVals login
		// if err := c.ShouldBind(&loginVals); err != nil {
		// 	return "", jwt.ErrMissingLoginValues
		// }
		// userID := loginVals.Username
		// password := loginVals.Password

		// if (userID == "admin" && password == "admin") || (userID == "test" && password == "test") {
		// 	return &User{
		// 		UserName:  userID,
		// 		LastName:  "Bo-Yi",
		// 		FirstName: "Wu",
		// 	}, nil
		// }
		// return nil, jwt.ErrFailedAuthentication
		return nil, nil
	}
}

func authorizator() func(data interface{}, c *gin.Context) bool {
	return func(data interface{}, c *gin.Context) bool {
		// if v, ok := data.(*User); ok && v.UserName == "admin" {
		// 	return true
		// }
		// return false
		return true
	}
}

func unauthorized() func(c *gin.Context, code int, message string) {
	return func(c *gin.Context, code int, message string) {
		c.JSON(code, gin.H{
			"code":    code,
			"message": message,
		})
	}
}

func loginResponse() func(c *gin.Context, code int, token string, expire time.Time) {
	return func(c *gin.Context, code int, token string, expire time.Time) {
		OK(c, gin.H{
			"token":  token,
			"expire": expire.Unix(),
		})
	}
}

