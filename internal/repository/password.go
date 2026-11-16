// Репозитоирий для данных
package repository

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"

	"github.com/alexsey-popov/gkeep/internal/models"
)

// PasswordList получение списка записей
func (rep *Repository) PasswordList(ctx context.Context, userID, name, page string) ([]*models.Password, error) {
	// Подготавливаем шаблон запроса и дополняем его поиском и пагинацией
	query := `SELECT id, user_id, name, login, password, created_at FROM passwords WHERE user_id = $1`
	args := []any{userID}
	query, args = rep.PrepareSearchAndPaginate(query, args, name, page)

	items, err := rep.Select[*models.Password](ctx, query, args...)
	if err != nil {
		rep.log.Error("Ошибка при получении списка паролей", slog.Any("error", err))
		return nil, ErrExec
	}

	return items, nil
}

// PasswordCreate создание новой записи
func (rep *Repository) PasswordCreate(ctx context.Context, item *models.Password) error {
	// Выполняем запрос
	query := `INSERT INTO passwords (user_id, name, login, password) VALUES (:user_id, :name, :login, :password) RETURNING id, user_id, name, login, password, created_at`
	if err := rep.Exec[*models.Password](ctx, query, item); err != nil {
		return err
	}

	return nil
}

// PasswordCreate обновление записи
func (rep *Repository) PasswordUpdate(ctx context.Context, item *models.Password) error {
	// Выполняем запрос
	query := `UPDATE passwords SET name = :name, login = :login, password = :password WHERE id = :id AND user_id = :user_id RETURNING id, user_id, name, login, password, created_at`
	if err := rep.Exec[*models.Password](ctx, query, item); err != nil {
		return err
	}

	return nil
}

// PasswordDelete удаление записи
func (rep *Repository) PasswordDelete(ctx context.Context, userID, id string) (*models.Password, error) {
	var item models.Password

	query := `DELETE FROM passwords WHERE user_id = $1 AND id = $2 RETURNING id, user_id, name, login, password, created_at`

	err := rep.db.GetContext(ctx, &item, query, userID, id)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			rep.log.Error("ошибка при удалении пароля из БД", slog.Any("error", err))
		}
		return &item, err
	}

	return &item, nil
}
