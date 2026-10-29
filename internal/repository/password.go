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

// PasswordCreate создание новой записи с паролем
func (rep DB) PasswordCreate(ctx context.Context, userID, name, login, password string) (*models.Password, error) {
	var item models.Password

	query := `INSERT INTO passwords (user_id, name, login, password) VALUES ($1, $2, $3, $4) RETURNING *`

	err := rep.db.GetContext(ctx, &item, query, userID, name, login, password)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation && pgErr.ConstraintName == "idx_passwords_name_user_id_unique" {
			return &item, ErrConflictUnique
		}

		rep.log.Error("ошибка при создании записи пароля в БД", slog.Any("error", err))
		return &item, errors.New("ошибка при создании записи пароля в БД")
	}

	return &item, nil
}

// PasswordList получение списка записей пользователя с возможной фильтрацией по имени и пагинацией
func (rep DB) PasswordList(ctx context.Context, userID, name, page string) ([]models.Password, error) {
	items := make([]models.Password, 0)

	query := `SELECT * FROM passwords WHERE user_id = $1`
	args := []any{userID}

	if name != "" {
		query += fmt.Sprintf(" AND name ILIKE $%d", len(args)+1)
		args = append(args, "%"+name+"%")
	}

	query += " ORDER BY created_at DESC"

	if page != "" {
		if pageNum, err := strconv.Atoi(page); err == nil && pageNum > 0 {
			pageSize := 20
			offset := (pageNum - 1) * pageSize
			query += fmt.Sprintf(" LIMIT %d OFFSET %d", pageSize, offset)
		}
	}

	err := rep.db.SelectContext(ctx, &items, query, args...)
	if err != nil {
		rep.log.Error("ошибка при получении списка паролей из БД", slog.Any("error", err))
		return nil, errors.New("ошибка при получении списка паролей из БД")
	}

	return items, nil
}

// PasswordUpdate обновление пароля по логину пользователя
func (rep DB) PasswordUpdate(ctx context.Context, userID, login, password string) (*models.Password, error) {
	var item models.Password

	query := `UPDATE passwords SET password = $1 WHERE user_id = $2 AND login = $3 RETURNING *`

	err := rep.db.GetContext(ctx, &item, query, password, userID, login)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			rep.log.Error("ошибка при обновлении пароля в БД", slog.Any("error", err))
		}
		return &item, err
	}

	return &item, nil
}

// PasswordDelete удаление записи пароля по идентификатору
func (rep DB) PasswordDelete(ctx context.Context, userID, id string) (*models.Password, error) {
	var item models.Password

	query := `DELETE FROM passwords WHERE user_id = $1 AND id = $2 RETURNING *`

	err := rep.db.GetContext(ctx, &item, query, userID, id)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			rep.log.Error("ошибка при удалении пароля из БД", slog.Any("error", err))
		}
		return &item, err
	}

	return &item, nil
}
