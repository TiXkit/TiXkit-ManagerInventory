package dto

import (
	"github.com/golang-jwt/jwt/v5"
	"time"
)

type RefreshToken struct {
	ID         int       `json:"id"`
	UserID     int       `json:"user_id"`
	TokenHash  string    `json:"token_hash"`
	DeviceInfo string    `json:"device_info"`
	IPAddress  string    `json:"ip_address"`
	ExpiresAt  time.Time `json:"expires_at"`
	IsRevoked  bool      `json:"is_revoked"`
	CreatedAt  time.Time `json:"created_at"`
}

type AccessToken struct {
	UserID int    `json:"user_id"`
	Jti    string `json:"jti"` // Уникальный идентификатор токена
	jwt.RegisteredClaims
}

type TokenPair struct {
	RefreshToken string `json:"refresh_token"`
	AccessToken  string `json:"access_token"`
}
