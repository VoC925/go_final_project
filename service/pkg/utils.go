package pkg

import (
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	lifeTimeToken = time.Hour * 5
)

// создание токена
func CreateToken() (string, error) {
	// payload
	claims := jwt.MapClaims{
		"expires": time.Now().Add(lifeTimeToken).Unix(), // время жизни токена
	}
	// token
	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		claims,
	)
	// ключ
	key := os.Getenv("JWT_SECRET")
	// signing token
	tokenStr, err := token.SignedString([]byte(key))
	if err != nil {
		return "", fmt.Errorf("failed to sign token with secret key")
	}
	return tokenStr, nil
}
