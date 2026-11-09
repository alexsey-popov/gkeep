// Репозитоирий для данных
package repository

import (
	"context"
	"errors"
	"log/slog"

	"github.com/alexsey-popov/gkeep/internal/crypto"
	"github.com/alexsey-popov/gkeep/internal/models"
	"github.com/jmoiron/sqlx"
)

// Repository репозиторий
type Repository interface {
	CreateUser(ctx context.Context, user *models.User) error
	GetUser(ctx context.Context, login string) (*models.User, error)

	PasswordList(ctx context.Context, userID, name, page string) ([]models.Password, error)
	PasswordCreate(ctx context.Context, password *models.Password) error
	PasswordUpdate(ctx context.Context, password *models.Password) error
	PasswordDelete(ctx context.Context, userID, id string) (*models.Password, error)

	CardList(ctx context.Context, userID, name, page string) ([]models.Card, error)
	CardCreate(ctx context.Context, userID, name, number, expired, code string) (*models.Card, error)
	CardUpdate(ctx context.Context, userID, id, name, number, expired, code string) (*models.Card, error)
	CardDelete(ctx context.Context, userID, id string) (*models.Card, error)

	TextList(ctx context.Context, userID, name, page string) ([]models.Text, error)
	TextCreate(ctx context.Context, userID, name, text string) (*models.Text, error)
	TextUpdate(ctx context.Context, userID, id, name, text string) (*models.Text, error)
	TextDelete(ctx context.Context, userID, id string) (*models.Text, error)

	FileList(ctx context.Context, userID, name, page string) ([]models.File, error)
	FileShow(ctx context.Context, userID, id string) (*models.File, error)
	FileCreate(ctx context.Context, userID, name, filename, storagePath string, size int64, checksum string) (*models.File, error)
	FileDelete(ctx context.Context, userID, id string) (*models.File, error)
}

// ErrConflictUnique ошибка конфликта уникальности при выполнении команды
var ErrConflictUnique = errors.New("ошибка уникальности по полю в БД")

// pageSize размер пачки данных (для операции получение списка)
const pageSize = 20

type Cipherer interface {
	Encode()
	Decode()
}

// DB Репозиторий БД
type DB struct {
	db     *sqlx.DB
	log    *slog.Logger
	cipher *crypto.Service
}

// Проверка на то, что структура DB удовлетворяет интерфейсу Repository
var _ Repository = (*DB)(nil)

// New Создание нового репозитория БД
func NewDB(db *sqlx.DB, log *slog.Logger, cipher *crypto.Service) *DB {
	return &DB{
		db:     db,
		log:    log,
		cipher: cipher,
	}
}

func (db *DB) List[T Cipherer](ctx context.Context, query string, args ...any) (items []T, err error) {
	err = db.db.SelectContext(ctx, &items, query, args...)
	if err != nil {
		db.log.Error("ошибка при получении списка записей", slog.Any("error", err))
	}
	return items, err
}

func (db *DB) Get[T Cipherer](ctx context.Context, query string, args ...any) (item T, err error) {
	err = db.db.GetContext(ctx, &item, query, args...)
	if err != nil {
		db.log.Error("ошибка при получении записи", slog.Any("error", err))
	}
	return item, err
}

func (db *DB) Store[T Cipherer](ctx context.Context, query string, args ...any) (item T, err error) {
	err = db.db.SelectContext(ctx, &item, query, args...)
	if err != nil {
		db.log.Error("ошибка при получении списка записей", slog.Any("error", err))
	}
	return item, err
}

func (db *DB) Update[T Cipherer](ctx context.Context, query string, args ...any) (item T, err error) {
	err = db.db.SelectContext(ctx, &item, query, args...)
	if err != nil {
		db.log.Error("ошибка при получении списка записей", slog.Any("error", err))
	}
	return item, err
}

func (db *DB) Destroy[T Cipherer](ctx context.Context, query string, args ...any) (item T, err error) {
	err = db.db.SelectContext(ctx, &item, query, args...)
	if err != nil {
		db.log.Error("ошибка при получении списка записей", slog.Any("error", err))
	}
	return item, err
}
