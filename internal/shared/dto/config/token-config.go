package config

import "time"

type TokenConfig struct {
	AccessSecret  string        // Секрет для подписи access токенов
	RefreshSecret string        // Секрет для подписи refresh токенов
	AccessTTL     time.Duration // Время жизни access токена (15m)
	RefreshTTL    time.Duration // Время жизни refresh токена (7d)
}
