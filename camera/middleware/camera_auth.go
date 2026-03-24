package middleware

import (
	"camera/db"
	"camera/utils"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gocql/gocql"
)

func CameraAccess() gin.HandlerFunc {
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

		role, ok := claims["role"].(string)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			c.Abort()
			return
		}

		email, ok := claims["email"].(string)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			c.Abort()
			return
		}

		var isLoggedIn bool
		var userRole string
		if err := db.Session.Query(`SELECT role, isloggedin FROM users WHERE email = ? LIMIT 1`, email).Consistency(gocql.One).Scan(&userRole, &isLoggedIn); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User validation failed"})
			c.Abort()
			return
		}

		if !isLoggedIn {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Session ended"})
			c.Abort()
			return
		}


		c.Set("user_id", claims["user_id"])
		c.Set("role", role)
		c.Next()
	}
}