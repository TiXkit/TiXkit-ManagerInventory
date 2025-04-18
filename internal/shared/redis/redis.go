package redis

import (
	"DBManager/internal/shared/config"
	"context"
	"github.com/redis/go-redis/v9"
	"log"
	"log/slog"
)

func InitRedis(ctx context.Context) *redis.Client {
	cfg := config.RedisConfig()

	rDB := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Pass,
		DB:       0,
	})

	if _, err := rDB.Ping(ctx).Result(); err != nil {
		log.Fatalf("Не удалось подключиться к Redis: %v", err)
	}

	slog.Info("Соединение с Redis успешно установлено")

	return rDB
}
