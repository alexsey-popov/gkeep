// Сервер менеджера паролей, обрабатывает gRPC запросы
package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/alexsey-popov/gkeep/internal/auth"
	"github.com/alexsey-popov/gkeep/internal/config"
	"github.com/alexsey-popov/gkeep/internal/handler"
	"github.com/alexsey-popov/gkeep/internal/repository"
	"github.com/alexsey-popov/gkeep/internal/server"
	"github.com/alexsey-popov/gkeep/migrations"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"golang.org/x/sync/errgroup"
)

const (
	Version   = "0.0.1"
	BuildDate = "2026-10-30"
)

func main() {
	// Создаём логгер
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))

	log.Info("gKeep", Version, BuildDate)

	// Создаём объект конфига сервера
	cfg, err := config.NewServer(os.Args[1:], os.LookupEnv)
	if err != nil {
		log.Error("Ошибка при парсинге конфига", slog.Any("error", err))
		os.Exit(1)
	}

	// Создаём объект взаимодействия с базой
	db, err := sqlx.Connect("pgx", cfg.DatabaseURI)
	if err != nil {
		log.Error("ошибка при подключении к БД",
			slog.Any("error", err),
		)
		os.Exit(1)
	}
	defer func() {
		if err = db.Close(); err != nil {
			log.Error("Ошибка при закрытии соединения с БД", slog.Any("error", err))
		}
	}()

	// Выполняем миграции БД
	if err = migrations.Up(db); err != nil {
		log.Error("Ошибка при выполнении миграций",
			slog.Any("error", err),
		)
		os.Exit(1)
	}

	// Создаём объект аутентификации
	guard := auth.New(cfg.JwtToken)

	// Создаём репозиторий
	rep := repository.NewDB(db, log)

	// Создаём обработчик запросов
	h := handler.New(rep, guard)

	srv, err := server.New(log, cfg, db, guard)
	if err != nil {
		log.Error("Ошибка при создании сервера", slog.Any("error", err))
		os.Exit(1)
	}

	// Контекст, который отменится при получении SIGINT или SIGTERM
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Создаём errgroup
	g, gCtx := errgroup.WithContext(ctx)

	// Запускаем сервер в отдельной горутине
	g.Go(func() error {
		return srv.Serve(h)
	})

	// Graceful shutdown в отдельной горутине
	g.Go(func() error {
		// Дожидаемся сигнала и пытаемся выполнить Graceful shutdown
		<-gCtx.Done()

		return srv.Close()
	})

	// Ожидаем завершения всех горутин
	if err = g.Wait(); err != nil {
		log.Info("завершение с ошибкой", slog.Any("error", err))
	}
}
