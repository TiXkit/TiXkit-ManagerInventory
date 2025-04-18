package service

import (
	"DBManager/internal/shared/dto"
	"context"
	"time"
)

type IAuthRepository interface {
	GetIDByEmail(ctx context.Context, email string) (int, error)
	ChangeHashDB(ctx context.Context, userID int, hash string) error
	GetHashByID(ctx context.Context, userID int) (string, error)
	AddUser(ctx context.Context, repo *dto.User) error
}

type IManagerRepository interface {
	DeleteRecord(ctx context.Context, recordID int) error
}

type IJWTTokenRepository interface {
	CreateRefreshToken(ctx context.Context, token *dto.RefreshToken) error
	GetRefreshTokenByHash(ctx context.Context, hash string) (*dto.RefreshToken, error)
	RevokeActiveRefreshTokens(ctx context.Context, userID int) (int, error)
	RevokeRefreshToken(ctx context.Context, hash string) error
	AddAccessToBlackList(ctx context.Context, jti string, expiresAt time.Duration) error
}
