package service

import (
	"DBManager/internal/repository"
	errors2 "DBManager/internal/service/errors"
	"DBManager/internal/shared/config"
	"DBManager/internal/shared/dto"
	"DBManager/internal/shared/utils"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"golang.org/x/crypto/bcrypt"
	"log/slog"
	"regexp"
	"time"
)

type IAuth interface {
	Authentication(ctx context.Context, creds *dto.SignInRequest, deviceInfo, ipAddress string) (*dto.TokenPair, error)
	Registration(ctx context.Context, creds *dto.SignUpRequest, deviceInfo, ipAddress string) (*dto.TokenPair, error)
	LogOut(ctx context.Context, claims *dto.AccessToken) error
	RefreshTokens(ctx context.Context, jti string) (*dto.TokenPair, error)
}

type Auth struct {
	repo    IAuthRepository
	jwtRepo IJWTTokenRepository
}

func NewAuth(repo IAuthRepository) *Auth {
	return &Auth{repo: repo}
}

func (a *Auth) Authentication(ctx context.Context, creds *dto.SignInRequest, deviceInfo, ipAddress string) (*dto.TokenPair, error) {
	// Проверяем, соответствует ли отправленный email формату.
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(creds.Email) {
		return nil, errors2.InvalidEmailFormat
	}
	// Проверяем, соответствует ли отправленный пароль формату.
	passwordRegex := regexp.MustCompile(`^(?=.*[A-Za-z])(?=.*\d)(?=.*[@$!%*#?&]).{8,}$`)
	if passwordRegex.MatchString(creds.Password) {
		return nil, errors2.InvalidPasswordFormat
	}

	// Получаем ID из БД по email.
	userID, err := a.repo.GetIDByEmail(ctx, creds.Email)
	if err != nil {
		return nil, err
	}

	// Получаем Hash из БД по ID.
	hash, err := a.repo.GetHashByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Проверяю, подходит ли пароль, если нет - возвращаю ошибку, что пароль недействителен.
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(creds.Password)); err != nil {
		return nil, err
	}

	// Проверяем, есть ли у пользователя активный(ые) Refresh Token, если да - отзываем.
	countRevoke, err := a.jwtRepo.RevokeActiveRefreshTokens(ctx, userID)
	if err != nil {
		if !errors.Is(err, repository.RecordNotFound) {
			return nil, err
		}
		slog.Info("У пользователя id:", userID, " было отозвано ", countRevoke, " Refresh токен(ов).")
	}

	// Генерируем Access Token.
	newAccessToken, _, err := utils.GenerateAccessToken(userID, config.TokenConfig().AccessTTL, []byte(config.TokenConfig().AccessSecret))
	if err != nil {
		return nil, err
	}

	// Генерируем Refresh Token.
	newRefreshToken, hashRefreshToken := utils.GenerateRefreshToken()

	// Помещаем Refresh Token в БД.
	if err := a.jwtRepo.CreateRefreshToken(ctx, &dto.RefreshToken{
		UserID:     userID,
		TokenHash:  hashRefreshToken,
		DeviceInfo: deviceInfo,
		IPAddress:  ipAddress,
		ExpiresAt:  time.Now().Add(config.TokenConfig().RefreshTTL),
		CreatedAt:  time.Now(),
		IsRevoked:  false,
	}); err != nil {
		return nil, err
	}

	// Возвращаем пару токенов.
	return &dto.TokenPair{
		RefreshToken: newRefreshToken,
		AccessToken:  newAccessToken,
	}, nil
}

func (a *Auth) Registration(ctx context.Context, creds *dto.SignUpRequest, deviceInfo, ipAddress string) (*dto.TokenPair, error) {
	// Проверяем, соответствует ли отправленный email формату.
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(creds.Email) {
		return nil, errors2.InvalidEmailFormat
	}
	// Проверяем, соответствует ли отправленный пароль формату.
	passwordRegex := regexp.MustCompile(`^(?=.*[A-Za-z])(?=.*\d)(?=.*[@$!%*#?&]).{8,}$`)
	if passwordRegex.MatchString(creds.Password) {
		return nil, errors2.InvalidPasswordFormat
	}

	// Проверяем существует ли пользователь с данным email в БД.
	_, err := a.repo.GetIDByEmail(ctx, creds.Email)
	if err != nil {
		// Если ошибка произошла не из-за отсутствия записи в БД - возвращаем дальше ошибку.
		if !errors.Is(err, repository.RecordNotFound) {
			return nil, err
		}
		// Если запись найдена - возвращаем ошибку, что данный email уже зарегистрирован.
	} else {
		return nil, errors2.EmailAlreadyExist
	}

	// Генерируем хэш из пароля, с константным значением 10 - дефолт.
	hash, err := bcrypt.GenerateFromPassword([]byte(creds.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// заполняем структуру пользователя
	var newUser dto.User
	newUser.FirstName = creds.FirstName
	newUser.LastName = creds.LastName
	newUser.Email = creds.Email
	newUser.Hash = string(hash) // передаём в БД хэш вместо пароля.
	newUser.CreatedAt = time.Now()
	newUser.UpdatedAt = time.Now()

	// Создаём нового пользователя.
	if err := a.repo.AddUser(ctx, &newUser); err != nil {
		return nil, err
	}

	// Генерируем Access Token.
	newAccessToken, _, err := utils.GenerateAccessToken(newUser.ID, config.TokenConfig().AccessTTL, []byte(config.TokenConfig().AccessSecret))
	if err != nil {
		return nil, err
	}

	// Генерируем Refresh Token.
	newRefreshToken, hashRefreshToken := utils.GenerateRefreshToken()

	// Помещаем Refresh Token в БД.
	if err := a.jwtRepo.CreateRefreshToken(ctx, &dto.RefreshToken{
		UserID:     newUser.ID,
		TokenHash:  hashRefreshToken,
		DeviceInfo: deviceInfo,
		IPAddress:  ipAddress,
		ExpiresAt:  time.Now().Add(config.TokenConfig().RefreshTTL),
		CreatedAt:  time.Now(),
		IsRevoked:  false,
	}); err != nil {
		return nil, err
	}

	// Возвращаем пару токенов.
	return &dto.TokenPair{
		RefreshToken: newRefreshToken,
		AccessToken:  newAccessToken,
	}, nil
}

func (a *Auth) LogOut(ctx context.Context, claims *dto.AccessToken) error {
	countRevoke, err := a.jwtRepo.RevokeActiveRefreshTokens(ctx, claims.UserID)
	if err != nil {
		if !errors.Is(err, repository.RecordNotFound) {
			return err
		}
		slog.Info("У пользователя id:", claims.UserID, " было отозвано ", countRevoke, " Refresh токен(ов).")
	}

	timeLife := claims.ExpiresAt.Sub(time.Now())
	if err := a.jwtRepo.AddAccessToBlackList(ctx, claims.Jti, timeLife); err != nil {
		return err
	}

	return nil
}

// RefreshTokens обновляет Refresh & Access токены.
func (a *Auth) RefreshTokens(ctx context.Context, jti string) (*dto.TokenPair, error) {
	// 1. Хэшируем токен.
	hash := sha256.Sum256([]byte(jti))
	tokenHash := hex.EncodeToString(hash[:])

	// 2. Получаем информацию о токене из БД по хэшу.
	tokenFromDB, err := a.jwtRepo.GetRefreshTokenByHash(ctx, tokenHash)
	if err != nil {
		return nil, err
	}

	// 3. Проверяем, не отозван ли токен
	if tokenFromDB.IsRevoked {
		return nil, errors2.ErrRefreshTokenRevoked
	}

	// 4. Проверяем срок действия
	if time.Now().After(tokenFromDB.ExpiresAt) {
		return nil, errors2.ErrRefreshTokenExpired
	}

	// 5. Отзываем старый refresh токен
	if err := a.jwtRepo.RevokeRefreshToken(ctx, tokenHash); err != nil {
		return nil, err
	}

	// 6. Генерируем Access Token.
	newAccessToken, _, err := utils.GenerateAccessToken(tokenFromDB.UserID, config.TokenConfig().AccessTTL, []byte(config.TokenConfig().AccessSecret))
	if err != nil {
		return nil, err
	}

	// 7. Генерируем Refresh Token.
	newRefreshToken, hashRefreshToken := utils.GenerateRefreshToken()

	// 8. Помещаем Refresh Token в БД.
	if err := a.jwtRepo.CreateRefreshToken(ctx, &dto.RefreshToken{
		UserID:     tokenFromDB.UserID,
		TokenHash:  hashRefreshToken,
		DeviceInfo: tokenFromDB.DeviceInfo,
		IPAddress:  tokenFromDB.IPAddress,
		ExpiresAt:  time.Now().Add(config.TokenConfig().RefreshTTL),
		CreatedAt:  time.Now(),
		IsRevoked:  false,
	}); err != nil {
		return nil, err
	}

	// 9. Возвращаем пару токенов.
	return &dto.TokenPair{
		RefreshToken: newRefreshToken,
		AccessToken:  newAccessToken,
	}, nil
}
