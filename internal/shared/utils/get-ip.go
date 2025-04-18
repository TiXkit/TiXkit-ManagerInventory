package utils

import (
	"net/http"
	"strings"
)

// GetIPAddress - Вспомогательная функция для получения IP.
func GetIPAddress(r *http.Request) string {
	// Пробуем получить IP из заголовков (если за прокси)
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		return strings.Split(ip, ",")[0]
	}
	return strings.Split(r.RemoteAddr, ":")[0]
}
