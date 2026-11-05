// Репозитоирий для данных
package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/alexsey-popov/gkeep/internal/models"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

// FileCreate создание новой записи с файлом
func (rep DB) FileCreate(ctx context.Context, userID, name, filename, storagePath string, size int64, checksum string) (*models.File, error) {
	var item models.File

	query := `INSERT INTO files (user_id, name, filename, storage_path, size, checksum) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id, user_id, name, filename, storage_path, size, checksum, created_at`

	err := rep.db.GetContext(ctx, &item, query, userID, name, filename, storagePath, size, checksum)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation && pgErr.ConstraintName == "idx_files_name_user_id_unique" {
			return &item, ErrConflictUnique
		}

		rep.log.Error("ошибка при создании записи файла в БД", slog.Any("error", err))
		return &item, errors.New("ошибка при создании записи файла в БД")
	}

	return &item, nil
}

// FileList получение списка файлов пользователя с возможной фильтрацией по имени и пагинацией
func (rep DB) FileList(ctx context.Context, userID, name, page string) ([]models.File, error) {
	items := make([]models.File, 0)

	query := `SELECT id, user_id, name, filename, storage_path, size, checksum, created_at FROM files WHERE user_id = $1`
	args := []any{userID}

	if name != "" {
		query += fmt.Sprintf(" AND name ILIKE $%d", len(args)+1)
		args = append(args, "%"+name+"%")
	}

	query += " ORDER BY created_at DESC"

	if page != "" {
		if pageNum, err := strconv.Atoi(page); err == nil && pageNum > 0 {
			offset := (pageNum - 1) * pageSize
			query += fmt.Sprintf(" LIMIT %d OFFSET %d", pageSize, offset)
		}
	}

	err := rep.db.SelectContext(ctx, &items, query, args...)
	if err != nil {
		rep.log.Error("ошибка при получении списка файлов из БД", slog.Any("error", err))
		return nil, errors.New("ошибка при получении списка файлов из БД")
	}

	return items, nil
}

// FileShow получение информации о файле по идентификатору
func (rep DB) FileShow(ctx context.Context, userID, id string) (*models.File, error) {
	var item models.File

	query := `SELECT id, user_id, name, filename, storage_path, size, checksum, created_at FROM files WHERE user_id = $1 AND id = $2`

	err := rep.db.GetContext(ctx, &item, query, userID, id)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			rep.log.Error("ошибка при получении файла из БД", slog.Any("error", err))
		}
		return &item, err
	}

	return &item, nil
}

// FileDelete удаление записи файла по идентификатору
func (rep DB) FileDelete(ctx context.Context, userID, id string) (*models.File, error) {
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
