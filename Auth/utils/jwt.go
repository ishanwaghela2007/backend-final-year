package utils

import (
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func GenerateToken(userID string, role string, email string, name string) (string, error) {
	secret := os.Getenv("JWT_SECRET")

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"role":    role,
		"email":   email,
		"name":    name,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	})

	return token.SignedString([]byte(secret))
}

func ParseToken(tokenStr string) (jwt.MapClaims, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		// Temporary debug log
		println("❌ ParseToken: JWT_SECRET is empty!")
	} else {
		// println("✅ ParseToken: JWT_SECRET loaded, len:", len(secret))
	}

	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})

	if err != nil || !token.Valid {
		println("❌ ParseToken Error:", err.Error())
		return nil, err
	}

	return token.Claims.(jwt.MapClaims), nil
}
