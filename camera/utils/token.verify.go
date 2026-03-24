package utils
import (
	"os"
	"github.com/golang-jwt/jwt/v5"
)
func ParseToken(tokenStr string) (jwt.MapClaims, error) {
	secret := os.Getenv("JWT_SECRET")

	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})

	if err != nil || !token.Valid {
		return nil, err
	}
	return token.Claims.(jwt.MapClaims), nil
}
