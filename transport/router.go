package transport

import (
	"DBManager/internal/shared/config"
	"log"
	"log/slog"
	"net/http"
)

// GoRouter - запускает сервер на указанном порту в .env файле.
func GoRouter(c *Controller) {
	// Создаем главный роутер
	mainRouter := http.NewServeMux()

	// Роутер для авторизации (с middleware)
	authRouter := http.NewServeMux()
	authRouter.HandleFunc("/SignIn", c.SignIn())
	authRouter.HandleFunc("/SignUp", c.SignUp())

	// Роутер для общих запросов (с middleware авторизации)
	generalRouter := http.NewServeMux()
	generalRouter.HandleFunc("/LogOut", c.LogOut())
	generalRouter.HandleFunc("/Refresh", c.RefreshTokens())

	// Подключаем роутеры с соответствующими middleware
	mainRouter.Handle("/", authRouter)                       // Без middleware авторизации
	mainRouter.Handle("/a/", c.Authorization(generalRouter)) // С middleware

	addr := config.HTTPConfig().Addr

	slog.Info("Сервер успешно запущен, на порту ", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Не удалось запустить сервер на порту %s. Ошибка: %s", addr, err)
	}
}
