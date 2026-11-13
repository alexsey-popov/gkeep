// Обработчик для gRPC запросов
package handler

import (
	"context"
	"log/slog"

	"github.com/alexsey-popov/gkeep/internal/auth"
	"github.com/alexsey-popov/gkeep/internal/crypto"
	"github.com/alexsey-popov/gkeep/internal/models"
	"github.com/alexsey-popov/gkeep/internal/validator"
	pb "github.com/alexsey-popov/gkeep/proto"
)

// Repository интерфейс репозитория для работы с данными
type Repository interface {
	// User
	CreateUser(ctx context.Context, user *models.User) error
	GetUser(ctx context.Context, login string) (*models.User, error)

	// Card
	CardList(ctx context.Context, userID, name, page string) ([]*models.Card, error)
	CardCreate(ctx context.Context, item *models.Card) error
	CardUpdate(ctx context.Context, item *models.Card) error
	CardDelete(ctx context.Context, userID, id string) (*models.Card, error)

	// Password
	PasswordList(ctx context.Context, userID, name, page string) ([]*models.Password, error)
	PasswordCreate(ctx context.Context, item *models.Password) error
	PasswordUpdate(ctx context.Context, item *models.Password) error
	PasswordDelete(ctx context.Context, userID, id string) (*models.Password, error)

	// Text
	TextList(ctx context.Context, userID, name, page string) ([]*models.Text, error)
	TextCreate(ctx context.Context, item *models.Text) error
	TextUpdate(ctx context.Context, item *models.Text) error
	TextDelete(ctx context.Context, userID, id string) (*models.Text, error)

	// File
	FileList(ctx context.Context, userID, name, page string) ([]*models.File, error)
	FileCreate(ctx context.Context, item *models.File) error
	FileShow(ctx context.Context, userID, id string) (*models.File, error)
	FileDelete(ctx context.Context, userID, id string) (*models.File, error)
}

// Handler Обработчик
type Handler struct {
	pb.UnimplementedAPIServer

	rep       Repository
	guard     *auth.Guard
	cipher    *crypto.Service
	validator *validator.Validator
	log       *slog.Logger
}

// New создание нового обработчика
func New(
	rep Repository,
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
