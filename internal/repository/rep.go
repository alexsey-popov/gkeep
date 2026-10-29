// Репозитоирий для данных
package repository

import (
	"context"
	"errors"
	"log/slog"

	"github.com/alexsey-popov/gkeep/internal/models"
	"github.com/jmoiron/sqlx"
)

// Repository репозиторий
type Repository interface {
	CreateUser(ctx context.Context, login, password string) (*models.User, error)
	GetUser(ctx context.Context, login string) (*models.User, error)

	PasswordList(ctx context.Context, userID, name, page string) ([]models.Password, error)
	PasswordCreate(ctx context.Context, userID, name, login, password string) (*models.Password, error)
	PasswordUpdate(ctx context.Context, userID, login, password string) (*models.Password, error)
	PasswordDelete(ctx context.Context, userID, id string) (*models.Password, error)
}

// ErrConflictUnique ошибка конфликта уникальности при выполнении команды
var ErrConflictUnique = errors.New("ошибка уникальности по полю в БД")

// DB Репозиторий БД
type DB struct {
	Repository

	db  *sqlx.DB
	log *slog.Logger
}

// New Создание нового репозитория БД
func NewDB(db *sqlx.DB, log *slog.Logger) *DB {
	return &DB{
		db:  db,
		log: log,
	}
}

// GenericRepository Общий generic-репозиторий для CRUD-операций
type GenericRepository[T any] struct {
	db  *sqlx.DB
	log *slog.Logger
}

// NewGenericRepository Создание нового generic-репозитория
func NewGenericRepository[T any](db *sqlx.DB, log *slog.Logger) *GenericRepository[T] {
	return &GenericRepository[T]{
		db:  db,
		log: log,
	}
}

// FindAll Получение списка записей (SELECT)
func (r GenericRepository[T]) FindAll(ctx context.Context, query string, args ...any) (items []T, err error) {
	err = r.db.SelectContext(ctx, &items, query, args...)
	if err != nil {
		r.log.Error("ошибка при получении списка записей", slog.Any("error", err))
	}
	return items, err
}

// FindOne Получение одной записи (GET)
func (r GenericRepository[T]) FindOne(ctx context.Context, query string, args ...any) (item T, err error) {
	err = r.db.GetContext(ctx, &item, query, args...)
	if err != nil {
		r.log.Error("ошибка при получении записи", slog.Any("error", err))
	}
	return item, err
}

// Insert Вставка или запрос с возвратом модели
func (r GenericRepository[T]) Insert(ctx context.Context, query string, args ...any) (item T, err error) {
	err = r.db.GetContext(ctx, &item, query, args...)
	if err != nil {
		r.log.Error("ошибка при выполнении запроса", slog.Any("error", err))
	}
	return item, err
}
