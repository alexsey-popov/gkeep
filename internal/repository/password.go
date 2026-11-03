// Репозитоирий для данных
package repository

import (
	"context"
	"database/sql"
	"encoding/hex"
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

	loginEncrypted, err := rep.cipher.Encrypt([]byte(login))
	if err != nil {
		return nil, err
	}

	passwordEncrypted, err := rep.cipher.Encrypt([]byte(password))
	if err != nil {
		return nil, err
	}

	query := `INSERT INTO passwords (user_id, name, login, password) VALUES ($1, $2, $3, $4) RETURNING id, user_id, name, login, password, created_at`

	err = rep.db.GetContext(ctx, &item, query, userID, name, fmt.Sprintf("%x", loginEncrypted), fmt.Sprintf("%x", passwordEncrypted))
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation && pgErr.ConstraintName == "idx_passwords_name_user_id_unique" {
			return &item, ErrConflictUnique
		}

		rep.log.Error("ошибка при создании записи пароля в БД", slog.Any("error", err))
		return &item, errors.New("ошибка при создании записи пароля в БД")
	}

	item.Login = login
	item.Password = password

	return &item, nil
}

// PasswordList получение списка записей пользователя с возможной фильтрацией по имени и пагинацией
func (rep DB) PasswordList(ctx context.Context, userID, name, page string) ([]models.Password, error) {
	items := make([]models.Password, 0)

	query := `SELECT id, user_id, name, login, password, created_at FROM passwords WHERE user_id = $1`
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
		rep.log.Error("ошибка при получении списка паролей из БД", slog.Any("error", err))
		return nil, errors.New("ошибка при получении списка паролей из БД")
	}

	for i := range items {
		loginCiphertext, err := hex.DecodeString(items[i].Login)
		if err != nil {
			rep.log.Error("Ошибка при декодировании hex поля", slog.Any("error", err))
			return nil, err
		}

		loginDecrypt, err := rep.cipher.Decrypt(loginCiphertext)
		if err != nil {
			rep.log.Error("Ошибка при декодировании поля", slog.Any("error", err))
			// TODO Переделай, не отправляй ошибку в клиент
			return nil, err
		}

		items[i].Login = string(loginDecrypt)

		passwordCiphertext, err := hex.DecodeString(items[i].Password)
		if err != nil {
			rep.log.Error("Ошибка при декодировании hex поля", slog.Any("error", err))
			return nil, err
		}

		passwordDecrypt, err := rep.cipher.Decrypt(passwordCiphertext)
		if err != nil {
			rep.log.Error("Ошибка при декодировании поля", slog.Any("error", err))
			// TODO Переделай, не отправляй ошибку в клиент
			return nil, err
		}

		items[i].Password = string(passwordDecrypt)
	}

	return items, nil
}

// PasswordUpdate обновление пароля по логину пользователя
func (rep DB) PasswordUpdate(ctx context.Context, userID, id, name, login, password string) (*models.Password, error) {
	var item models.Password

	loginEncrypted, err := rep.cipher.Encrypt([]byte(login))
	if err != nil {
		return nil, err
	}

	passwordEncrypted, err := rep.cipher.Encrypt([]byte(password))
	if err != nil {
		return nil, err
	}

	query := `UPDATE passwords SET name = $1, login = $2, password = $3 WHERE id = $4 AND user_id = $5 RETURNING id, user_id, name, login, password, created_at`

	err = rep.db.GetContext(ctx, &item, query, name, fmt.Sprintf("%x", loginEncrypted), fmt.Sprintf("%x", passwordEncrypted), id, userID)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			rep.log.Error("ошибка при обновлении пароля в БД", slog.Any("error", err))
		}
		return &item, err
	}

	item.Login = login
	item.Password = password

	return &item, nil
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
