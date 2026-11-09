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

// PasswordList получение списка записей пользователя с возможной фильтрацией по имени и пагинацией
func (rep DB) PasswordList(ctx context.Context, userID, name, page string) ([]models.Password, error) {
	items := make([]models.Password, 0)

	query := `SELECT id, user_id, name, login, password, created_at FROM passwords WHERE user_id = $1`
	args := []any{userID}

	// Поиск по названию (заложил на бэке, но не реализовал в клиенте)
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

	// Делаем запрос в БД
	err := rep.db.SelectContext(ctx, &items, query, args...)
	if err != nil {
		rep.log.Error("ошибка при получении списка паролей из БД", slog.Any("error", err))
		return nil, errors.New("ошибка при получении списка паролей из БД")
	}

	// Расшифровываем данные
	for i := range items {
		err = items[i].Decrypt(rep.cipher)
		if err != nil {
			rep.log.Error("ошибка при шифровании данных", slog.Any("error", err))
			return nil, errors.New("ошибка при расшифровывании данных, :%w")
		}
	}

	return items, nil
}

// PasswordCreate создание новой записи с паролем
func (rep DB) PasswordCreate(ctx context.Context, item *models.Password) error {
	// Шифруем чувствительные данные
	if err := item.Encrypt(rep.cipher); err != nil {
		rep.log.Error("ошибка при шифровании данных", slog.Any("error", err))
		return fmt.Errorf("ошибка при шифровании данных: %w", err)
	}

	// Делаем запрос к БД
	query := `INSERT INTO passwords (user_id, name, login, password) VALUES ($1, $2, $3, $4) RETURNING id, user_id, name, login, password, created_at`
	if err := rep.db.GetContext(ctx, item, query, item.UserId, item.Name, item.Login, item.Password); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation && pgErr.ConstraintName == "idx_passwords_name_user_id_unique" {
			return ErrConflictUnique
		}

		rep.log.Error("ошибка при создании записи пароля в БД", slog.Any("error", err))
		return errors.New("ошибка при создании записи пароля в БД")
	}

	// Расшифровываем чувствительные данные
	if err := item.Decrypt(rep.cipher); err != nil {
		rep.log.Error("ошибка при расшифровывании данных", slog.Any("error", err))
		return fmt.Errorf("ошибка при расшифровывании данных: %w", err)
	}

	return nil
}

// PasswordUpdate обновление пароля по логину пользователя
func (rep DB) PasswordUpdate(ctx context.Context, item *models.Password) error {
	// Шифруем чувствительные данные
	if err := item.Encrypt(rep.cipher); err != nil {
		rep.log.Error("ошибка при шифровании данных", slog.Any("error", err))
		return fmt.Errorf("ошибка при шифровании данных: %w", err)
	}

	query := `UPDATE passwords SET name = $1, login = $2, password = $3 WHERE id = $4 AND user_id = $5 RETURNING id, user_id, name, login, password, created_at`
	err := rep.db.GetContext(ctx, item, query, item.Name, item.Login, item.Password, item.Id, item.UserId)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			rep.log.Error("ошибка при обновлении пароля в БД", slog.Any("error", err))
		}
		return err
	}

	// Расшифровываем чувствительные данные
	if err = item.Decrypt(rep.cipher); err != nil {
		rep.log.Error("ошибка при расшифровывании данных", slog.Any("error", err))
		return fmt.Errorf("ошибка при расшифровывании данных: %w", err)
	}

	return nil
}

// PasswordDelete удаление записи пароля по идентификатору
func (rep DB) PasswordDelete(ctx context.Context, userID, id string) (*models.Password, error) {
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
