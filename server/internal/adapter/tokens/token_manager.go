// Package service содержит реализацию парсера JWT-токенов.
package tokens

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/oegegr/gophkeeper/server/internal/config"

	// "github.com/oegegr/gophkeeper/server/internal/usecases"
	"github.com/samber/do/v2"
)

// var _ usecases.TokenManager = (*JWTTokenManager)(nil)

// Claims представляет структуру данных для хранения в JWT-токене.
type Claims struct {
	// UserID представляет идентификатор пользователя.
	UserID string `json:"user_id"`
	// RegisteredClaims представляет зарегистрированные данные в JWT-токене.
	jwt.RegisteredClaims
}

// JWTTokenManager представляет парсер JWT-токенов.
type JWTTokenManager struct {
	// jwtSecret представляет секретный ключ для подписи JWT-токенов.
	jwtSecret string
}

// ErrInvalidJWTToken представляет ошибку, которая возникает при невалидном JWT-токене.
var ErrInvalidJWTToken = errors.New("invalid token")

func ResolveJWTTokenManager(i do.Injector) *JWTTokenManager {
	appConfig, _ := do.Invoke[*config.Config](i)
	return NewJWTTokenManager(appConfig.TokenSecret)
}

// NewJWTParser возвращает новый экземпляр JWTParser.
// Эта функция принимает секретный ключ для подписи JWT-токенов и логгер.
func NewJWTTokenManager(jwtSecret string) *JWTTokenManager {
	return &JWTTokenManager{jwtSecret}
}

// GenerateToken создает новый JWT-токен для пользователя.
// Эта функция принимает идентификатор пользователя и возвращает сгенерированный JWT-токен.
func (v *JWTTokenManager) GenerateToken(userID string) (string, error) {
	claims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(v.jwtSecret))
}

// ValidateToken извлекает идентификатор пользователя из JWT-токена.
// Эта функция принимает JWT-токен и возвращает идентификатор пользователя, если токен валиден.
func (v *JWTTokenManager) ValidateToken(tokenString string) (string, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(v.jwtSecret), nil
	})

	if err != nil {
		return "", err
	}

	if !token.Valid {
		return "", ErrInvalidJWTToken
	}

	return claims.UserID, nil
}
