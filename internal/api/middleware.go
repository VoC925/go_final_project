package api

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/VoC925/go_final_project/internal/httpResponse"
	"github.com/VoC925/go_final_project/pkg"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// auth - middleware для аутентификации пользователя
func auth(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var (
			startTime = time.Now()
			pass      = os.Getenv("TODO_PASSWORD")
			jwt       string // JWT-токен из куки
		)
		// получаем куку
		cookie, err := r.Cookie("token")
		if err == nil {
			jwt = cookie.Value
		}
		// валидация токена
		claims, err := validateJWTToken(jwt)
		if err != nil {
			logrus.Debug("token invalid")
			httpResponse.Error(w, httpResponse.NewLogInfo("auth", r, nil, time.Since(startTime),
				httpResponse.NewError(
					http.StatusUnauthorized,
					errors.Wrap(err, "валидация токена"),
				),
			))
			return
		}
		// проверка хеша пароля
		hashedPassword, ok := claims["hesh"].(string)
		if !ok {
			logrus.Debug("hesh invalid")
			httpResponse.Error(w, httpResponse.NewLogInfo("auth", r, nil, time.Since(startTime),
				httpResponse.NewError(
					http.StatusUnauthorized,
					fmt.Errorf("отсутствует хэш пароля в токене"),
				),
			))
			return
		}
		if !pkg.IsValidHash(hashedPassword, pass) {
			logrus.Debug("token already expires")
			httpResponse.Error(w, httpResponse.NewLogInfo("auth", r, nil, time.Since(startTime),
				httpResponse.NewError(
					http.StatusUnauthorized,
					fmt.Errorf("хэш пароля не равен хэшу в токене"),
				),
			))
			return
		}
		logrus.Debug("token validated successfully")
		next(w, r)
	})
}

// validateJWTToken получает мапу jwt.MapClaims из токена tokenStr
func validateJWTToken(tokenStr string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		key := os.Getenv("JWT_SECRET")
		return []byte(key), nil
	})
	if err != nil {
		fmt.Println("failed to parse token")
		return nil, err
	}
	if !token.Valid {
		fmt.Println("not valid token")
		return nil, err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("Anauthorized")
	}
	return claims, nil
}

// logUUID - middleware для формирования уникального ID для логирования
func logUUID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cid := uuid.New()
		ctx := context.WithValue(r.Context(), "log_uuid", cid.String())
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
