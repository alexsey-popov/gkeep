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
)

// FileList получение списка файлов
func (rep *Repository) FileList(ctx context.Context, userID, name, page string) ([]*models.File, error) {
	// Подготавливаем шаблон запроса и дополняем его поиском и пагинацией
	query := `SELECT id, user_id, name, filename, storage_path, size, checksum, created_at FROM files WHERE user_id = $1`
	args := []any{userID}
	query, args = rep.PrepareSearchAndPaginate(query, args, name, page)

	items, err := rep.GetMany[*models.File](ctx, query, args...)
	if err != nil {
		rep.log.Error("Ошибка при получении списка паролей", slog.Any("error", err))
		return nil, ErrExec
	}

	return items, nil
}

// FileCreate создание новой записи с файлом
func (rep *Repository) FileCreate(ctx context.Context, item *models.File) error {
	query := `INSERT INTO files (user_id, name, filename, storage_path, size, checksum) VALUES (:user_id, :name, :filename, :storage_path, :size, :checksum) RETURNING id, user_id, name, filename, storage_path, size, checksum, created_at`
	err := rep.PutOne[*models.File](ctx, query, item)

	if err != nil {
		// Если ошибка связана с уникальностью связки name-user_id, то выдаём спец сообщение и не логируем
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation && pgErr.ConstraintName == "idx_files_name_user_id_unique" {
			return ErrConflictUnique
		}

		rep.log.Error("ошибка при создании записи в БД", slog.Any("error", err))
		return ErrExec
	}

	return nil
}

// FileShow получение информации о файле по идентификатору
func (rep *Repository) FileShow(ctx context.Context, userID, id string) (*models.File, error) {
	query := `SELECT id, user_id, name, filename, storage_path, size, checksum, created_at FROM files WHERE user_id = :user_id AND id = :id`

	item := &models.File{
		Id:     id,
		UserId: userID,
	}

	err := rep.PutOne[*models.File](ctx, query, item)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			rep.log.Error("ошибка при получении файла из БД", slog.Any("error", err))
		}
		return item, err
	}

	return item, nil
}

// FileDelete удаление записи файла по идентификатору
func (rep *Repository) FileDelete(ctx context.Context, userID, id string) (*models.File, error) {
	var item models.File

	query := `DELETE FROM files WHERE user_id = $1 AND id = $2 RETURNING id, user_id, name, filename, storage_path, size, checksum, created_at`

	err := rep.db.GetContext(ctx, &item, query, userID, id)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			rep.log.Error("ошибка при удалении записи файла из БД", slog.Any("error", err))
		}
		return &item, err
	}

	return &item, nil
}
