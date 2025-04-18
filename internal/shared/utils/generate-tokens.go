package utils

import (
	"DBManager/internal/shared/dto"
	"crypto/sha256"
	"encoding/hex"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"time"
)

// GenerateAccessToken - генерирует Access token и вшивает в него данные, применяя метод хеша sha256.
// Возвращает 3 переменные - сформированный токен в строке, уникальный идентификатор access токена и возможную ошибку.
func GenerateAccessToken(userID int, expiresIn time.Duration, secret []byte) (string, string, error) {
	// Формируем уникальный идентификатор токена.
	jti := uuid.NewString()

	claims := dto.AccessToken{
		UserID: userID,
		Jti:    jti,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresIn)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ID:        jti,
		},
	}

	// Зашиваем данные в токен и указываем дальнейший метод хеша.
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Подписываем токен с использованием секретного ключа и хэшируем.
	tokenString, err := token.SignedString(secret)
	return tokenString, jti, err // Возвращаем сформированный токен в виде строки...,
	//уникальный идентификатор токена и ошибку, если есть.
}

// GenerateRefreshToken - генерирует Refresh token, применяя метод хеша sha256.
// Возвращает 2 переменные - refresh token и refresh token, только хэшированный.
func GenerateRefreshToken() (string, string) {
	// генерируем случайный UUID в качестве токена
	token := uuid.NewString()

	// Создаем SHA-256 хеш от токена для безопасного хранения в БД
	hash := sha256.Sum256([]byte(token))
	hashToken := hex.EncodeToString(hash[:])

	// Возвращаем токен и хеш от токена. Первый - пользователю, второй - в БД.
	return token, hashToken
}
