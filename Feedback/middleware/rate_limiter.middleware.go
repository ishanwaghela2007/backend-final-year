package middleware

import (
	"context"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis_rate/v10"
	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()

// NewRedisUserRateLimiter applies Redis rate limiting per user ID (from JWT).
// limit is a redis_rate.Limit, e.g., redis_rate.PerMinute(10)
func NewRedisUserRateLimiter(rdb *redis.Client, rateLimit redis_rate.Limit) gin.HandlerFunc {
	limiter := redis_rate.NewLimiter(rdb)

	return func(c *gin.Context) {
		key := extractUserID(c)
		if key == "" {
			// fallback to IP for unauthenticated requests
			key = c.ClientIP()
		}

		res, err := limiter.Allow(ctx, key, rateLimit)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "rate limiter error"})
			return
		}

		if res.Allowed == 0 {
			resetIn := res.ResetAfter.Seconds() // FIX ✅
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":   "Too many requests",
				"message": "Rate limit exceeded. Try again later.",
				"retryIn": resetIn,
			})
			return
		}

		// continue
		c.Next()
	}
}

func extractUserID(c *gin.Context) string {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return ""
	}
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return ""
	}
	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
	secret := []byte(os.Getenv("JWT_SECRET"))
	if len(secret) == 0 {
		secret = []byte("supersecretkey")
	}

	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		return secret, nil
	})
	if err != nil || !token.Valid {
		return ""
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		// your token uses "user_id"
		if uid, ok := claims["user_id"].(string); ok {
			return uid
		}
		// fallback to sub if present
		if sub, ok := claims["sub"].(string); ok {
			return sub
		}
	}
	return ""
}
