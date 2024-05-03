package pkg

import (
	"fmt"
	"os"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// создание токена
func CreateToken(str string) (string, error) {
	// получение хэша пароля
	hesh, err := GeneratehashFromString(str)
	if err != nil {
		return "", err
	}
	// payload
	claims := jwt.MapClaims{
		"hesh": hesh, // хэш пароля
	}
	// token
	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		claims,
	)
	// ключ из env
	key := os.Getenv("JWT_SECRET")
	// signing token
	tokenStr, err := token.SignedString([]byte(key))
	if err != nil {
		return "", fmt.Errorf("failed to sign token with secret key")
	}
	return tokenStr, nil
}

// GeneratehashFromString создает хэш на основе строки используя алгоритм пакета bcrypt
func GeneratehashFromString(str string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(str), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

// IsValidHash проверяет соответсвие пароля и хеша
func IsValidHash(hash string, password string) bool {
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		return false
	}
	return true
}
