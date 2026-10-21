// Репозитоирий для данных
package repository

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"

	"github.com/alexsey-popov/gkeep/internal/models"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jmoiron/sqlx"
)

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

// CreateUser создание нового пользователя
func (rep DB) CreateUser(ctx context.Context, login, passHash string) (*models.User, error) {
	var user models.User

	// Добавляем нового пользователя в БД
	query := `INSERT INTO users (login, password) VALUES ($1, $2) RETURNING *`

	err := rep.db.GetContext(ctx, &user, query, login, passHash)
	if err != nil {
		// Если произошла ошибка уникальности по полю login - выводим соответствующую ошибку
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation && pgErr.ConstraintName == "idx_users_login_unique" {
			return &user, ErrConflictUnique
		}

		rep.log.Error("ошибка при создании нового пользователя в БД", slog.Any("error", err))
		return &user, errors.New("ошибка при создании нового пользователя в БД")
	}

	return &user, nil
}

// GetUser получение пользователя по логину
func (rep DB) GetUser(ctx context.Context, login string) (*models.User, error) {
	var user models.User

	err := rep.db.GetContext(ctx, &user, `SELECT * FROM users WHERE login = $1 LIMIT 1`, login)
	if err != nil {
		// Логируем любую ошибку, кроме sql.ErrNoRows
		if !errors.Is(err, sql.ErrNoRows) {
			rep.log.Error("ошибка при получении данных пользователя из БД", slog.Any("error", err))
		}

		return &user, errors.New("ошибка при получении данных пользователя из БД")
	}

	return &user, nil
}
