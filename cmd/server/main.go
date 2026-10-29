// Сервер
package main

import (
	"fmt"
	"log/slog"
	"net"
	"os"

	"github.com/alexsey-popov/gkeep/internal/auth"
	"github.com/alexsey-popov/gkeep/internal/config"
	"github.com/alexsey-popov/gkeep/internal/handler"
	"github.com/alexsey-popov/gkeep/internal/repository"
	"github.com/alexsey-popov/gkeep/migrations"
	pb "github.com/alexsey-popov/gkeep/proto"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"
)

func main() {
	// Создаём логгер
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// Создаём объект конфига сервера
	cfg, err := config.NewServer(os.Args[1:], os.LookupEnv)
	if err != nil {
		log.Error(err.Error(),
			slog.Any("error", err),
		)
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
	defer db.Close()

	// Выполняем миграции БД
	if err = migrations.Up(db); err != nil {
		log.Error(err.Error(),
			slog.Any("error", err),
		)
		os.Exit(1)
	}

	// Нужно определить порт для сервера
	listen, err := net.Listen("tcp", cfg.Address)
	if err != nil {
		slog.Error("ошибка при инициализации listener", "error", err)
		os.Exit(1)
	}

	// Создаём объект аутентификации
	guard := auth.New(cfg.JwtToken)

	// Получаем TLS сертификат
	tlsCreds, err := generateTLSCreds()
	if err != nil {
		slog.Error("ошибка получении TLS сертификата", "error", err)
		os.Exit(1)
	}

	// Создаем gRPC сервер с зарегистрированным интерцептором аутентификации
	s := grpc.NewServer(
		grpc.Creds(tlsCreds),
		grpc.UnaryInterceptor(guard.UnaryServerInterceptor()),
	)

	rep := repository.NewDB(db, log)
	h := handler.New(rep, guard)

	//// Регистрируем сервис
	pb.RegisterAPIServer(s, h)

	// Включаем reflection — одна строка
	reflection.Register(s)

	//
	fmt.Println("сервер gRPC начал работу")
	// Получение запроса gRpc
	if err := s.Serve(listen); err != nil {
		slog.Error("ошибка при работе сервера", "error", err)
		os.Exit(1)
	}
}

// generateTLSCreds получение TLS сертификата сервера
func generateTLSCreds() (credentials.TransportCredentials, error) {
	certFile := "server.crt"
	keyFile := "server.key"

	return credentials.NewServerTLSFromFile(certFile, keyFile)
}
