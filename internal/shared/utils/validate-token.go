package utils

import (
	"DBManager/internal/shared/dto"
	"errors"
	"github.com/golang-jwt/jwt/v5"
)

func ValidateAccessToken(tokenString, secret string) (*dto.AccessToken, error) {
	// Парсим токен с проверкой подписи
	token, err := jwt.ParseWithClaims(tokenString, &dto.AccessToken{},
		func(token *jwt.Token) (interface{}, error) {
			// Проверяем, что используется ожидаемый алгоритм подписи
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return []byte(secret), nil
		},
	)
	if err != nil {
		return nil, err
	}

	// Проверяем, что claims имеют правильный тип и токен валиден
	if claims, ok := token.Claims.(*dto.AccessToken); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}
