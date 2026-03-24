package routes

import (
	"Auth/utils"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func ProfileRoutes(router *gin.Engine) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing token"})
			c.Abort()
			return
		}
		token := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := utils.ParseToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Profile fetched successfully",
			"user":    claims,
		})
	}
}
