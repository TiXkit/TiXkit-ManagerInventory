package transport

import (
	"DBManager/internal/service"
	errors2 "DBManager/internal/service/errors"
	"DBManager/internal/shared/config"
	"DBManager/internal/shared/dto"
	"DBManager/internal/shared/utils"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"
)

type Controller struct {
	service.IAuth
}

func NewController(auth service.IAuth) *Controller {
	return &Controller{auth}
}

func (c *Controller) SignIn() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Декодируем JSON в структуру creds.
		var creds dto.SignInRequest
		if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		// Получаем данные из заголовка X-Device-Info и получаем IP address.
		deviceInfo := r.Header.Get("X-Device-Info")
		ipAddress := utils.GetIPAddress(r)

		// Аутентифицируем пользователя - проверяем пароль и логин, создаём токены.
		tokens, err := c.IAuth.Authentication(r.Context(), &creds, deviceInfo, ipAddress)
		if err != nil {
			if errors.Is(err, errors2.PasswordWrong) || errors.Is(err, errors2.UserNotExist) {
				http.Error(w, "Неверно введен Email или пароль.", http.StatusUnauthorized)
			}
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Устанавливаем Refresh токен в куки.
		utils.SetRefreshTokenCookie(w, tokens.RefreshToken, false)

		// Устанавливаем заголовок и возвращаем в теле токены.
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(tokens); err != nil {
			http.Error(w, "Не удалось серелизовать объект в JSON", http.StatusInternalServerError)
			return
		}
	}
}

func (c *Controller) SignUp() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Декодируем JSON в структуру creds.
		var creds dto.SignUpRequest
		if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		// Получаем данные из заголовка X-Device-Info и получаем IP address.
		deviceInfo := r.Header.Get("X-Device-Info")
		ipAddress := utils.GetIPAddress(r)

		// Не забыть отозвать старый токен...

		// Регистрируем пользователя - добавляем в бд, создаём токены.
		tokens, err := c.IAuth.Registration(r.Context(), &creds, deviceInfo, ipAddress)
		if err != nil {
			if errors.Is(err, errors2.PasswordWrong) || errors.Is(err, errors2.UserNotExist) {
				http.Error(w, "Неверно введен Email или пароль.", http.StatusUnauthorized)
			}
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Устанавливаем Refresh токен в куки.
		utils.SetRefreshTokenCookie(w, tokens.RefreshToken, false)

		// Устанавливаем заголовок и возвращаем в теле токены.
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]string{
			"access_token": tokens.AccessToken,
			// Refresh token не возвращаем, так как он в куках
		}); err != nil {
			slog.Error("JSON encode error: ", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
	}
}

func (c *Controller) LogOut() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Получаем данные из заголовка Авторизации.
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "authorization header required", http.StatusUnauthorized)
			return
		}

		// Убираем байты слова "Bearer ", что получить чистую строку с Access Token
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			http.Error(w, "invalid authorization header format", http.StatusUnauthorized)
			return
		}

		// Валидируем Access Token, расшифровываем его секретным ключом и сравниваем данные.
		claims, err := utils.ValidateAccessToken(tokenString, config.TokenConfig().AccessSecret)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		if err := c.IAuth.LogOut(r.Context(), claims); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("Successfully logged out"))
	}
}

func (c *Controller) RefreshTokens() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("refresh_token")
		if err != nil {
			http.Error(w, "Refresh token not found in cookies", http.StatusBadRequest)
			return
		}

		refreshToken := cookie.Value

		tokens, err := c.IAuth.RefreshTokens(r.Context(), refreshToken)
		if err != nil {
			if errors.Is(err, errors2.ErrRefreshTokenExpired) || errors.Is(err, errors2.ErrRefreshTokenRevoked) {
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			slog.Error("RefreshTokens error: ", err)
			return
		}

		// Устанавливаем Refresh токен в куки.
		utils.SetRefreshTokenCookie(w, tokens.RefreshToken, false)

		// Устанавливаем заголовок и возвращаем в теле токены.
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]string{
			"access_token": tokens.AccessToken,
			// Refresh token не возвращаем, так как он в куках
		}); err != nil {
			slog.Error("JSON encode error: ", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
	}
}

func (c *Controller) Authorization(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_ = next
		// Получаем данные из заголовка Авторизации.
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "authorization header required", http.StatusUnauthorized)
			return
		}

		// Убираем байты слова "Bearer ", что получить чистую строку с Access Token
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			http.Error(w, "invalid authorization header format", http.StatusUnauthorized)
			return
		}

		// Валидируем Access Token, расшифровываем его секретным ключом и сравниваем данные.
		claims, err := utils.ValidateAccessToken(tokenString, config.TokenConfig().AccessSecret)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		// Устанавливаем заголовки.
		w.Header().Set("Content-Type", "application/json")

		// Серелизуем карту с данными в JSON.
		if err := json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "protected content",
			"user_id": claims.UserID,
			"jti":     claims.Jti,
		}); err != nil {
			http.Error(w, "Не удалось серелизовать объект в JSON", http.StatusInternalServerError)
			return
		}
	}
}
