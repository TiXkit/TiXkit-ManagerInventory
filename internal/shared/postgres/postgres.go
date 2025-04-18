package postgres

import (
	"DBManager/internal/shared/config"
	"DBManager/internal/shared/dto"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log/slog"
)

func InitPostgres() (*gorm.DB, error) {
	// Получение строки подключения к БД из конфига
	cfg, err := config.PgSQLConfig()
	if err != nil {
		return nil, err
	}

	// Подключение к базе данных
	db, err := gorm.Open(postgres.Open(cfg.Addr), &gorm.Config{})
	if err != nil {
		slog.Error("Не удалось подключиться к Postgres.", "Ошибка", err)
		return nil, err
	}

	// Применяем миграции к БД, мигрируя следующие таблицы:
	if err := db.AutoMigrate(&dto.User{}); err != nil {
		slog.Error("Не удалось подключиться к Postgres.", "Ошибка", err)
		return nil, err
	}

	slog.Info("Соединение с Postgres успешно установлено")

	return db, nil
}
