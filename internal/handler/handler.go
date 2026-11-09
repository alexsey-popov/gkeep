// Обработчик для gRPC запросов
package handler

import (
	"log/slog"

	"github.com/alexsey-popov/gkeep/internal/auth"
	"github.com/alexsey-popov/gkeep/internal/crypto"
	"github.com/alexsey-popov/gkeep/internal/repository"
	"github.com/alexsey-popov/gkeep/internal/validator"
	pb "github.com/alexsey-popov/gkeep/proto"
)

// Handler Обработчик
type Handler struct {
	pb.UnimplementedAPIServer

	rep       *repository.Repository
	guard     *auth.Guard
	cipher    *crypto.Service
	validator *validator.Validator
	log       *slog.Logger
}

// New создание нового обработчика
func New(
	rep *repository.Repository,
	guard *auth.Guard,
	cipher *crypto.Service,
	val *validator.Validator,
	log *slog.Logger,
) *Handler {
	return &Handler{
		rep:       rep,
		guard:     guard,
		cipher:    cipher,
		validator: val,
		log:       log,
	}
}
