package repository

import (
	"DBManager/internal/shared/dto"
	"context"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"time"
)

// Errors:
var (
	RecordNotFound = errors.New("не удалось найти запись в Базе Данных")
)

type AuthRepo struct {
	db  *gorm.DB
	rDB *redis.Client
}

func NewAuthRepo(db *gorm.DB, rDB *redis.Client) *AuthRepo {
	return &AuthRepo{db, rDB}
}

// GetIDByEmail - получается ID пользователя по Email из базы.
func (ar *AuthRepo) GetIDByEmail(ctx context.Context, email string) (int, error) {
	var user dto.User

	if err := ar.db.WithContext(ctx).Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, err
		}
		return 0, err
	}

	return user.ID, nil
}

// GetHashByID - получает хэш пользователя из базы по id.
func (ar *AuthRepo) GetHashByID(ctx context.Context, userID int) (string, error) {
	var user dto.User

	if err := ar.db.WithContext(ctx).Where("id = ?", userID).First(&user).Error; err != nil {
		return "", err
	}

	return user.Hash, nil
}

// ChangeHashDB - Заменяет старый хэш пользователя по id на новый.
func (ar *AuthRepo) ChangeHashDB(ctx context.Context, userID int, hash string) error {
	if err := ar.db.WithContext(ctx).Model(&dto.User{}).Where("id = ?", userID).Update("hash", hash).Error; err != nil {
		return err
	}

	return nil
}

// AddUser - Создаёт запись с новым пользователем в БД.
func (ar *AuthRepo) AddUser(ctx context.Context, user *dto.User) error {
	if err := ar.db.WithContext(ctx).Create(user).Error; err != nil {
		return err
	}
	return nil
}

func (ar *AuthRepo) CreateRefreshToken(ctx context.Context, token *dto.RefreshToken) error {
	if err := ar.db.WithContext(ctx).Create(token).Error; err != nil {
		return err
	}
	return nil
}

func (ar *AuthRepo) GetRefreshTokenByHash(ctx context.Context, hash string) (*dto.RefreshToken, error) {
	var refreshToken dto.RefreshToken

	if err := ar.db.WithContext(ctx).Where("token_hash = ?", hash).First(&refreshToken).Error; err != nil {
		return nil, err
	}

	return &refreshToken, nil
}

// RevokeActiveRefreshTokens отзывает все активные refresh-токены пользователя
func (ar *AuthRepo) RevokeActiveRefreshTokens(ctx context.Context, userID int) (int, error) {
	result := ar.db.WithContext(ctx).Model(&dto.RefreshToken{}).
		Where("user_id = ? AND is_revoked = ? AND expires_at > ?", userID, false, time.Now()).
		Update("is_revoked", true)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return 0, RecordNotFound
		}
		return 0, result.Error
	}

	// Возвращаем количество отозванных токенов.
	return int(result.RowsAffected), nil
}

func (ar *AuthRepo) RevokeRefreshToken(ctx context.Context, hash string) error {
	if err := ar.db.WithContext(ctx).Model(&dto.RefreshToken{}).Where("token_hash = ?", hash).Update("is_revoked", true).Error; err != nil {
		return err
	}

	return nil
}

func (ar *AuthRepo) AddAccessToBlackList(ctx context.Context, jti string, expiresAt time.Duration) error {
	blackListKey := fmt.Sprintf("AccessBlock:%s", jti)

	err := ar.rDB.Set(ctx, blackListKey, nil, time.Minute*expiresAt).Err()
	if err != nil {
		return err
	}

	return nil
}
