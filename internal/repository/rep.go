// Репозитоирий для данных
package repository

import (
	"context"
	"errors"

	"github.com/alexsey-popov/gkeep/internal/models"
)

// Repository репозиторий
type Repository interface {
	CreateUser(ctx context.Context, login, password string) (*models.User, error)
	GetUser(ctx context.Context, login string) (*models.User, error)
}

// ErrConflictUnique ошибка конфликта уникальности при выполнении команды
var ErrConflictUnique = errors.New("ошибка уникальности по полю в БД")
