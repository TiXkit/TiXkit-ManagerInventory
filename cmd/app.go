package main

import (
	"DBManager/internal/repository"
	"DBManager/internal/service"
	"DBManager/internal/shared/postgres"
	"DBManager/internal/shared/redis"
	"DBManager/transport"
	"context"
	"github.com/joho/godotenv"
	"log"
	"log/slog"
	"time"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("FATAL не удалось загрузить .env")
	}
	slog.Info(".env успешно подгружен")

	db, err := postgres.InitPostgres()
	if err != nil {
		log.Fatal("FATAL Error initializing database: ", err)
	}

	slog.Info("Context timeout set to five")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rDB := redis.InitRedis(ctx)

	// Определение слоя репозитория.
	repo := repository.NewAuthRepo(db, rDB)

	// Определения сервисного слоя бизнес-логики.
	authService := service.NewAuth(repo)

	// Определение транспортного слоя.
	controller := transport.NewController(authService)

	// Запуск сервера.
	transport.GoRouter(controller)
}
