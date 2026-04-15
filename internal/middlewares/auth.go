package middlewares

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware checks if the request has a valid authorization token
func AuthMiddleware(secretToken string) gin.HandlerFunc {
	// Added hardcoded value check but we will do JWT validation in actual production
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token != secretToken {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}
		c.Next()
	}
}
