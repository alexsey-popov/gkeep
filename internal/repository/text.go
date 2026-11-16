// Репозитоирий для данных
package repository

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"

	"github.com/alexsey-popov/gkeep/internal/models"
)

// TextList получение записей
func (rep *Repository) TextList(ctx context.Context, userID, name, page string) ([]*models.Text, error) {
	// Подготавливаем шаблон запроса и дополняем его поиском и пагинацией
	query := `SELECT id, user_id, name, text FROM texts WHERE user_id = $1`
	args := []any{userID}
	query, args = rep.PrepareSearchAndPaginate(query, args, name, page)

	items, err := rep.Select[*models.Text](ctx, query, args...)
	if err != nil {
		return nil, err
	}

	return items, nil
}

// TextCreate создание новой записи
func (rep *Repository) TextCreate(ctx context.Context, item *models.Text) error {
	// Выполняем запрос БД
	query := `INSERT INTO texts (user_id, name, text) VALUES (:user_id, :name, :text) RETURNING id, user_id, name, text, created_at`
	if err := rep.Exec[*models.Text](ctx, query, item); err != nil {
		return err
	}

	return nil
}

// TextUpdate обновление записи
func (rep *Repository) TextUpdate(ctx context.Context, item *models.Text) error {
	// Выполняем запрос а БД
	query := `UPDATE texts SET name = :name, text = :text WHERE id = :id AND user_id = :user_id RETURNING id, user_id, name, text, created_at`
	if err := rep.Exec[*models.Text](ctx, query, item); err != nil {
		return err
	}

	return nil
}

// TextDelete удаление записи
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
