package config

import (
	"DBManager/internal/shared/dto/config"
	"log"
	"os"
	"time"
)

var (
	accessTTL  = time.Minute * 15
	refreshTTL = time.Hour * 24 * 7
)

func PgSQLConfig() (*config.PostgresConfig, error) {
	addr := os.Getenv("POSTGRES_CONNECT_STRING")

	if addr == "" {
		log.Fatal("Не удалось получить данные для подключения к Postgres из файла .env")
	}

	return &config.PostgresConfig{
		Addr: addr,
	}, nil
}

func RedisConfig() *config.RedisConfig {
	addr := os.Getenv("REDIS_ADDR")
	pass := os.Getenv("REDIS_PASS")

	if addr == "" || pass == "" {
		log.Fatal("Не удалось получить данные для подключения к Redis из файла .env")
	}

	return &config.RedisConfig{
		Addr: addr,
		Pass: pass,
	}
}

func TokenConfig() *config.TokenConfig {
	accessSecret := os.Getenv("ACCESS_SECRET")
	refreshSecret := os.Getenv("REFRESH_SECRET")

	return &config.TokenConfig{
		AccessSecret:  accessSecret,
		RefreshSecret: refreshSecret,
		AccessTTL:     accessTTL,
		RefreshTTL:    refreshTTL,
	}
}

func HTTPConfig() *config.HTTPConfig {
	httpAddr := os.Getenv("HTTP_ADDR")
	if httpAddr == "" {
		httpAddr = ":8080"
	}

	return &config.HTTPConfig{Addr: httpAddr}
}
