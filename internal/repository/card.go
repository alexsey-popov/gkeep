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

// CardList получение списка банковских карт
func (rep *Repository) CardList(ctx context.Context, userID, name, page string) ([]*models.Card, error) {
	// Подготавливаем шаблон запроса и дополняем его поиском и пагинацией
	query := `SELECT id, user_id, name, number, expired, code FROM cards WHERE user_id = $1`
	args := []any{userID}
	query, args = rep.PrepareSearchAndPaginate(query, args, name, page)

	items, err := rep.Select[*models.Card](ctx, query, args...)
	if err != nil {
		rep.log.Error("Ошибка при получении списка банковских карт", slog.Any("error", err))
		return nil, ErrExec
	}

	return items, nil
}

// CardCreate создание новой записи с банковской картой
func (rep *Repository) CardCreate(ctx context.Context, item *models.Card) error {
	// Выполняем запрос
	query := `INSERT INTO cards (user_id, name, number, expired, code) VALUES (:user_id, :name, :number, :expired, :code) RETURNING id, user_id, name, number, expired, code, created_at`
	err := rep.Exec[*models.Card](ctx, query, item)

	if err != nil {
		// Если ошибка связана с уникальностью связки name-user_id, то выдаём спец сообщение и не логируем
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation && pgErr.ConstraintName == "idx_cards_name_user_id_unique" {
			return ErrConflictUnique
		}

		rep.log.Error("ошибка при создании записи в БД", slog.Any("error", err))
		return ErrExec
	}

	return nil
}

// CardUpdate обновление пароля по логину пользователя
func (rep *Repository) CardUpdate(ctx context.Context, item *models.Card) error {
	query := `UPDATE cards SET name = :name, number = :number, expired = :expired, code = :code WHERE id = :id AND user_id = :user_id RETURNING id, user_id, name, number, expired, code, created_at`
	err := rep.Exec[*models.Card](ctx, query, item)

	if err != nil {
		// Если ошибка связана с уникальностью связки name-user_id, то выдаём спец сообщение и не логируем
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation && pgErr.ConstraintName == "idx_cards_name_user_id_unique" {
			return ErrConflictUnique
		}

		rep.log.Error("ошибка при создании записи в БД", slog.Any("error", err))
		return ErrExec
	}

	return nil
}

// CardDelete удаление записи пароля по идентификатору
func (rep *Repository) CardDelete(ctx context.Context, userID, id string) (*models.Card, error) {
	var item models.Card

	query := `DELETE FROM cards WHERE user_id = $1 AND id = $2 RETURNING id, user_id, name, number, expired, code, created_at`

	err := rep.db.GetContext(ctx, &item, query, userID, id)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			rep.log.Error("ошибка при удалении записи из БД", slog.Any("error", err))
		}
		return &item, err
	}

	return &item, nil
}
