// Репозитоирий для данных
package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/alexsey-popov/gkeep/internal/crypto"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jmoiron/sqlx"
)

var (
	ErrConflictUnique = errors.New("ошибка уникальности по полю в БД")
	ErrPrepare        = errors.New("ошибка при подготовке запроса")
	ErrExec           = errors.New("ошибка при выполнении запроса")
	ErrNoRows         = errors.New("данные не найдены")
	ErrEncrypt        = errors.New("ошибка при шифровании данных")
	ErrDecrypt        = errors.New("ошибка при расшифровывании данных")
)

// pageSize размер пачки данных (для операции получение списка)
const pageSize = 20

// Cipherer интерфейс для моделей, которые можно зашифровать/расшифровать
type Cipherer interface {
	Encrypt(cipher *crypto.Service) error
	Decrypt(cipher *crypto.Service) error
}

// Repository Репозиторий
type Repository struct {
	db     *sqlx.DB
	log    *slog.Logger
	cipher *crypto.Service
}

// New Создание нового репозитория
func New(db *sqlx.DB, log *slog.Logger, cipher *crypto.Service) *Repository {
	return &Repository{
		db:     db,
		log:    log,
		cipher: cipher,
	}
}

// PrepareSearchAndPaginate подготавливает к запросу поиск и пагинацию
func (rep *Repository) PrepareSearchAndPaginate(query string, args []any, search, page string) (string, []any) {
	// Поиск по названию (заложил на бэке, но не реализовал в клиенте)
	if search != "" {
		query += fmt.Sprintf(" AND name ILIKE $%d", len(args)+1)
		args = append(args, "%"+search+"%")
	}
	query += " ORDER BY created_at DESC"

	if page != "" {
		if pageNum, err := strconv.Atoi(page); err == nil && pageNum > 0 {
			offset := (pageNum - 1) * pageSize
			query += fmt.Sprintf(" LIMIT %d OFFSET %d", pageSize, offset)
		}
	}

	return query, args
}

// Select получение списка записей
func (rep *Repository) Select[T Cipherer](ctx context.Context, query string, args ...any) (items []T, err error) {
	// Делаем запрос
	err = rep.db.SelectContext(ctx, &items, query, args...)
	if err != nil {
		rep.log.Error("Ошибка при совершении запроса",
			slog.Any("error", err),
			slog.String("query", query),
			slog.Any("args", args),
		)

		return items, ErrExec
	}

	// Расшифровываем данные
	for i := range items {
		err = items[i].Decrypt(rep.cipher)
		if err != nil {
			rep.log.Error("Ошибка при расшифровывании данных", slog.Any("error", err))

			return items, ErrDecrypt
		}
	}

	return items, nil
}

// Exec создание записи
func (rep *Repository) Exec[T Cipherer](ctx context.Context, query string, item T) error {
	// Шифруем данные
	if err := item.Encrypt(rep.cipher); err != nil {
		rep.log.Error("Ошибка при шифровании данных", slog.Any("error", err))

		return ErrEncrypt
	}

	// Подготавливаем запрос
	stmt, err := rep.db.PrepareNamed(query)
	if err != nil {
		rep.log.Error("Ошибка при формировании sql запроса", slog.Any("error", err))

		return ErrPrepare
	}
	defer func() {
		if err := stmt.Close(); err != nil {
			rep.log.Error("Ошибка при закрытии подготовленного запроса", slog.Any("error", err))
		}
	}()

	err = stmt.GetContext(ctx, item, item)
	if err != nil {
		// Если ошибка связана с sql.ErrNoRows, то выдаём спец сообщение и не логируем
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNoRows
		}

		// Если ошибка связана с уникальностью, то выдаём спец сообщение и не логируем
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return ErrConflictUnique
		}

		// Иначе логируем ошибку и выводим общий ответ
		rep.log.Error("Ошибка при совершении запроса",
			slog.Any("error", err),
			slog.String("query", query),
			slog.Any("item", item),
		)

		return ErrExec
	}

	// Расшифровываем данные
	if err = item.Decrypt(rep.cipher); err != nil {
		rep.log.Error("Ошибка при расшифровывании данных", slog.Any("error", err))

		return ErrDecrypt
	}

	return nil
}
