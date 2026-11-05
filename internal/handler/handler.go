// Обработчик для gRPC запросов
package handler

import (
	"github.com/alexsey-popov/gkeep/internal/auth"
	"github.com/alexsey-popov/gkeep/internal/crypto"
	"github.com/alexsey-popov/gkeep/internal/repository"
	pb "github.com/alexsey-popov/gkeep/proto"
)

// Handler Обработчик
type Handler struct {
	pb.UnimplementedAPIServer

	rep    repository.Repository
	guard  *auth.Guard
	cipher *crypto.Service
}

// New создание нового обработчика
func New(rep repository.Repository, guard *auth.Guard, cipher ...*crypto.Service) *Handler {
	var c *crypto.Service
	if len(cipher) > 0 {
		c = cipher[0]
	}

	h := &Handler{
		rep:    rep,
		guard:  guard,
		cipher: c,
	}

	return h
}
