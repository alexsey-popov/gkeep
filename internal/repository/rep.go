// Репозитоирий для данных
package repository

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/alexsey-popov/gkeep/internal/crypto"
	"github.com/jmoiron/sqlx"
)

var (
	ErrConflictUnique = errors.New("ошибка уникальности по полю в БД")
	ErrPrepare        = errors.New("ошибка при подготовке запроса")
	ErrExec           = errors.New("ошибка при выполнении запроса")
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

// GetMany получение списка записей
func (rep *Repository) GetMany[T Cipherer](ctx context.Context, query string, args ...any) (items []T, err error) {
	// Делаем запрос
	err = rep.db.SelectContext(ctx, &items, query, args...)
	if err != nil {
		//rep.log.Error(ErrExec.Error(), slog.Any("error", err))
		//return nil, ErrExec
		return items, errors.Join(ErrExec, err)
	}

	// Расшифровываем данные
	for i := range items {
		err = items[i].Decrypt(rep.cipher)
		if err != nil {
			//rep.log.Error(ErrDecrypt.Error(), slog.Any("error", err))
			//return nil, ErrDecrypt
			return items, errors.Join(ErrDecrypt, err)
		}
	}

	return items, nil
}

// PutOne создание записи
func (rep *Repository) PutOne[T Cipherer](ctx context.Context, query string, item T) (T, error) {
	// Шифруем данные
	if err := item.Encrypt(rep.cipher); err != nil {
		//rep.log.Error(ErrEncrypt.Error(), slog.Any("error", err))
		//return item, ErrEncrypt
		return item, errors.Join(ErrEncrypt, err)
	}

	// Подготавливаем запрос
	stmt, err := rep.db.PrepareNamed(query)
	if err != nil {
		return item, errors.Join(ErrPrepare, err)
	}
	defer stmt.Close()

	err = stmt.GetContext(ctx, item, item)
	if err != nil {
		return item, errors.Join(ErrExec, err)
	}

	// Расшифровываем данные
	if err = item.Decrypt(rep.cipher); err != nil {
		//rep.log.Error(ErrDecrypt.Error(), slog.Any("error", err))
		//return item, ErrDecrypt
		return item, errors.Join(ErrDecrypt, err)
	}

	return item, nil
}
