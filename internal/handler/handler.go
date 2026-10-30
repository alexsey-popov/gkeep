// Обработчик для gRPC запросов
package handler

import (
	"github.com/alexsey-popov/gkeep/internal/auth"
	"github.com/alexsey-popov/gkeep/internal/repository"
	pb "github.com/alexsey-popov/gkeep/proto"
)

// Handler Обработчик
type Handler struct {
	pb.UnimplementedAPIServer

	rep   repository.Repository
	guard *auth.Guard
}

// New создание нового обработчика
func New(rep repository.Repository, guard *auth.Guard) *Handler {
	h := &Handler{
		rep:   rep,
		guard: guard,
	}

	return h
}
