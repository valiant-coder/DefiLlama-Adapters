package v1

import (
	"exapp-go/api"
	"exapp-go/internal/db/db"
	"exapp-go/internal/errno"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	// ContextKeySubAccount is the key used to store sub-account info in context
	ContextKeySubAccount = "sub_account"
)

// AuthMiddleware authenticates requests using API key from Authorization header
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get API key from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			api.Error(c, errno.DefaultParamsError("missing API key"))
			c.Abort()
			return
		}

		// Remove any "Bearer " prefix if present
		apiKey := strings.TrimPrefix(authHeader, "Bearer ")
		apiKey = strings.TrimSpace(apiKey)

		if apiKey == "" {
			api.Error(c, errno.DefaultParamsError("invalid API key format"))
			c.Abort()
			return
		}

		// Get sub-account by API key
		subAccount, err := db.New().GetUserSubAccountByAPIKey(c.Request.Context(), apiKey)
		if err != nil {
			api.Error(c, errno.DefaultParamsError("invalid API key"))
			c.Abort()
			return
		}

		// Store sub-account in context for later use
		c.Set(ContextKeySubAccount, subAccount)
		c.Next()
	}
}

// GetSubAccountFromContext retrieves the authenticated sub-account from context
func GetSubAccountFromContext(c *gin.Context) *db.UserSubAccount {
	value, exists := c.Get(ContextKeySubAccount)
	if !exists {
		return nil
	}
	if subAccount, ok := value.(*db.UserSubAccount); ok {
		return subAccount
	}
	return nil
}
