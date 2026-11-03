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

// CardCreate создание новой записи с картой
func (rep DB) CardCreate(ctx context.Context, userID, name, number, expired, code string) (*models.Card, error) {
	var item models.Card

	numberEncrypted, err := rep.cipher.Encrypt([]byte(number))
	if err != nil {
		return nil, err
	}

	expiredEncrypted, err := rep.cipher.Encrypt([]byte(expired))
	if err != nil {
		return nil, err
	}

	codeEncrypted, err := rep.cipher.Encrypt([]byte(code))
	if err != nil {
		return nil, err
	}

	query := `INSERT INTO cards (user_id, name, number, expired, code) VALUES ($1, $2, $3, $4, $5) RETURNING id, user_id, name, number, expired, code, created_at`

	err = rep.db.GetContext(ctx, &item, query, userID, name, fmt.Sprintf("%x", numberEncrypted), fmt.Sprintf("%x", expiredEncrypted), fmt.Sprintf("%x", codeEncrypted))
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation && pgErr.ConstraintName == "idx_cards_name_user_id_unique" {
			return &item, ErrConflictUnique
		}

		rep.log.Error("ошибка при создании записи карты в БД", slog.Any("error", err))
		return &item, errors.New("ошибка при создании записи карты в БД")
	}

	item.Number = number
	item.Expired = expired
	item.Code = code

	return &item, nil
}

// CardList получение списка записей пользователя с возможной фильтрацией по имени и пагинацией
func (rep DB) CardList(ctx context.Context, userID, name, page string) ([]models.Card, error) {
	items := make([]models.Card, 0)

	query := `SELECT id, user_id, name, number, expired, code FROM cards WHERE user_id = $1`
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
		rep.log.Error("ошибка при получении списка карт из БД", slog.Any("error", err))
		return nil, errors.New("ошибка при получении списка карт из БД")
	}

	for i := range items {
		numberCiphertext, err := hex.DecodeString(items[i].Number)
		if err != nil {
			rep.log.Error("Ошибка при декодировании hex поля", slog.Any("error", err))
			return nil, err
		}

		numberDecrypt, err := rep.cipher.Decrypt(numberCiphertext)
		if err != nil {
			rep.log.Error("Ошибка при декодировании поля", slog.Any("error", err))
			// TODO Переделай, не отправляй ошибку в клиент
			return nil, err
		}

		expiredCiphertext, err := hex.DecodeString(items[i].Expired)
		if err != nil {
			rep.log.Error("Ошибка при декодировании hex поля", slog.Any("error", err))
			return nil, err
		}

		expiredDecrypt, err := rep.cipher.Decrypt(expiredCiphertext)
		if err != nil {
			rep.log.Error("Ошибка при декодировании поля", slog.Any("error", err))
			// TODO Переделай, не отправляй ошибку в клиент
			return nil, err
		}

		codeCiphertext, err := hex.DecodeString(items[i].Code)
		if err != nil {
			rep.log.Error("Ошибка при декодировании hex поля", slog.Any("error", err))
			return nil, err
		}

		codeDecrypt, err := rep.cipher.Decrypt(codeCiphertext)
		if err != nil {
			rep.log.Error("Ошибка при декодировании поля", slog.Any("error", err))
			// TODO Переделай, не отправляй ошибку в клиент
			return nil, err
		}

		items[i].Number = string(numberDecrypt)
		items[i].Expired = string(expiredDecrypt)
		items[i].Code = string(codeDecrypt)
	}

	return items, nil
}

// CardUpdate обновление пароля по логину пользователя
func (rep DB) CardUpdate(ctx context.Context, userID, id, name, number, expired, code string) (*models.Card, error) {
	var item models.Card

	numberEncrypted, err := rep.cipher.Encrypt([]byte(number))
	if err != nil {
		return nil, err
	}

	expiredEncrypted, err := rep.cipher.Encrypt([]byte(expired))
	if err != nil {
		return nil, err
	}

	codeEncrypted, err := rep.cipher.Encrypt([]byte(code))
	if err != nil {
		return nil, err
	}

	query := `UPDATE cards SET name = $1, number = $2, expired = $3, code = $4 WHERE id = $5 AND user_id = $6 RETURNING id, user_id, name, number, expired, code, created_at`

	err = rep.db.GetContext(ctx, &item, query, name, fmt.Sprintf("%x", numberEncrypted), fmt.Sprintf("%x", expiredEncrypted), fmt.Sprintf("%x", codeEncrypted), id, userID)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			rep.log.Error("ошибка при обновлении пароля в БД", slog.Any("error", err))
		}
		return &item, err
	}

	item.Number = number
	item.Expired = expired
	item.Code = code

	return &item, nil
}

// CardDelete удаление записи пароля по идентификатору
func (rep DB) CardDelete(ctx context.Context, userID, id string) (*models.Card, error) {
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
