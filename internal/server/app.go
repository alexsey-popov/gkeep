// Сервер для обработки gRPC запросов
package server

import (
	"errors"
	"fmt"
	"log/slog"
	"net"
	"time"

	"github.com/alexsey-popov/gkeep/internal/auth"
	"github.com/alexsey-popov/gkeep/internal/config"
	pb "github.com/alexsey-popov/gkeep/proto"
	"github.com/jmoiron/sqlx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"
)

// Server структура для взаимодействия с сервером
type Server struct {
	log *slog.Logger
	cfg *config.Server
	db  *sqlx.DB
	srv *grpc.Server
}

// shutdownDuration дедлайн для Graceful shutdown
var shutdownDuration = 15 * time.Second

// New создание нового сервера
func New(log *slog.Logger, cfg *config.Server, db *sqlx.DB, guard *auth.Guard) (*Server, error) {
	// Получаем TLS сертификат
	tlsCreds, err := credentials.NewServerTLSFromFile(cfg.TLSCert, cfg.TLSKey)
	if err != nil {
		return nil, fmt.Errorf("ошибка получении TLS сертификата: %w", err)
	}

	// Создаем gRPC сервер с зарегистрированным интерцептором аутентификации
	srv := grpc.NewServer(
		grpc.Creds(tlsCreds),
		grpc.UnaryInterceptor(guard.UnaryServerInterceptor()),
	)

	return &Server{
		log: log,
		cfg: cfg,
		db:  db,
		srv: srv,
	}, nil
}

// Serve запуск сервера
func (s *Server) Serve(h pb.APIServer) error {
	// Прослушиваем порт для сервера
	listen, err := net.Listen("tcp", s.cfg.Address)
	if err != nil {
		s.log.Error("Ошибка при попытке занять порт", slog.Any("error", err))
		return err
	}

	// Регистрируем сервис
	pb.RegisterAPIServer(s.srv, h)

	// Включаем reflection, чтобы сторонние клиенты(postman) могли получить список доступных роутов
	reflection.Register(s.srv)

	s.log.Info("сервер gRPC начал работу")

	// Поднимаем сервер
	if err = s.srv.Serve(listen); err != nil {
		return fmt.Errorf("ошибка при работе сервера: %w", err)
	}

	return nil
}

// Close остановка сервера
func (s *Server) Close() error {
	stopped := make(chan struct{})
	go func() {
		s.srv.GracefulStop()
		close(stopped)
	}()

	select {
	case <-stopped:
		s.log.Info("Сервер(gRPC) успешно остановлен")
		return nil
	case <-time.After(shutdownDuration):
		s.log.Error("Сервер(gRPC) не успел обработать запросы до таймаута и завершен принудительно")
		s.srv.Stop()
		return errors.New("Сервер(gRPC) не успел обработать запросы до таймаута и завершен принудительно")
	}
}
