package api

import (
	"log"
	"time"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
)

var (
	identityKey = "user_id"
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

	return &jwt.GinJWTMiddleware{
		Realm:       "exapp",
		Key:         []byte("secret key"),
		Timeout:     time.Hour,
		MaxRefresh:  time.Hour,
		IdentityKey: identityKey,
		PayloadFunc: payloadFunc(),

		IdentityHandler: identityHandler(),
		Authenticator:   authenticator(),
		Authorizator:    authorizator(),
		Unauthorized:    unauthorized(),
		TokenLookup:     "header: Authorization, query: token, cookie: jwt",
		// TokenLookup: "query:token",
		// TokenLookup: "cookie:token",
		TokenHeadName: "Bearer",
		TimeFunc:      time.Now,
	}
}

// payload set
func payloadFunc() func(data interface{}) jwt.MapClaims {
	return func(data interface{}) jwt.MapClaims {
		// if v, ok := data.(*User); ok {
		// 	return jwt.MapClaims{
		// 		identityKey: v.UserName,
		// 	}
		// }
		return jwt.MapClaims{}
	}
}


func identityHandler() func(c *gin.Context) interface{} {
	return func(c *gin.Context) interface{} {
		// claims := jwt.ExtractClaims(c)
		// return &User{
		// 	UserName: claims[identityKey].(string),
		// }
		return nil
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
