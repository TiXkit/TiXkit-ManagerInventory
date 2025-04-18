package utils

import (
	"net/http"
	"time"
)

// SetRefreshTokenCookie - Вспомогательная функция для установки cookie
func SetRefreshTokenCookie(w http.ResponseWriter, token string, isProduction bool) {
	cookie := &http.Cookie{
		Name:     "refresh_token",
		Value:    token,
		Path:     "/",
		Expires:  time.Now().Add(7 * 24 * time.Hour),
		MaxAge:   604800,
		HttpOnly: true,
		Secure:   isProduction, // true в production, false в development
		SameSite: http.SameSiteLaxMode,
	}

	http.SetCookie(w, cookie)
}
