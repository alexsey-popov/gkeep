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

// TextList получение списка текстовых записей
func (rep *Repository) TextList(ctx context.Context, userID, name, page string) ([]*models.Text, error) {
	// Подготавливаем шаблон запроса и дополняем его поиском и пагинацией
	query := `SELECT id, user_id, name, text FROM texts WHERE user_id = $1`
	args := []any{userID}
	query, args = rep.PrepareSearchAndPaginate(query, args, name, page)

	items, err := rep.GetMany[*models.Text](ctx, query, args...)
	if err != nil {
		rep.log.Error("Ошибка при получении списка текстовых записей", slog.Any("error", err))
		return nil, ErrExec
	}

	return items, nil
}

// TextCreate создание новой записи с картой
func (rep *Repository) TextCreate(ctx context.Context, item *models.Text) error {
	// Выполняем запрос
	query := `INSERT INTO texts (user_id, name, text) VALUES (:user_id, :name, :text) RETURNING id, user_id, name, text, created_at`
	item, err := rep.PutOne[*models.Text](ctx, query, item)

	if err != nil {
		// Если ошибка связана с уникальностью связки name-user_id, то выдаём спец сообщение и не логируем
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation && pgErr.ConstraintName == "idx_texts_name_user_id_unique" {
			return ErrConflictUnique
		}

		rep.log.Error("ошибка при создании записи в БД", slog.Any("error", err))
		return ErrExec
	}

	return nil
}

// TextUpdate обновление пароля по логину пользователя
func (rep *Repository) TextUpdate(ctx context.Context, item *models.Text) error {
	// Выполняем запрос
	query := `UPDATE texts SET name = :name, text = :text WHERE id = :id AND user_id = :user_id RETURNING id, user_id, name, text, created_at`
	item, err := rep.PutOne[*models.Text](ctx, query, item)

	if err != nil {
		// Если ошибка связана с уникальностью связки name-user_id, то выдаём спец сообщение и не логируем
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation && pgErr.ConstraintName == "idx_texts_name_user_id_unique" {
			return ErrConflictUnique
		}

		rep.log.Error("ошибка при создании записи в БД", slog.Any("error", err))
		return ErrExec
	}

	return nil
}

// TextDelete удаление записи пароля по идентификатору
func (rep *Repository) TextDelete(ctx context.Context, userID, id string) (*models.Text, error) {
	var item models.Text

	query := `DELETE FROM texts WHERE user_id = $1 AND id = $2 RETURNING id, user_id, name, text, created_at`

	err := rep.db.GetContext(ctx, &item, query, userID, id)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			rep.log.Error("ошибка при удалении записи из БД", slog.Any("error", err))
		}
		return &item, err
	}

	return &item, nil
}
