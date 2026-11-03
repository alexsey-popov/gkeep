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

// TextCreate создание новой записи с картой
func (rep DB) TextCreate(ctx context.Context, userID, name, text string) (*models.Text, error) {
	var item models.Text

	textEncrypted, err := rep.cipher.Encrypt([]byte(text))
	if err != nil {
		return nil, err
	}

	query := `INSERT INTO texts (user_id, name, text) VALUES ($1, $2, $3) RETURNING id, user_id, name, text, created_at`

	err = rep.db.GetContext(ctx, &item, query, userID, name, fmt.Sprintf("%x", textEncrypted))
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation && pgErr.ConstraintName == "idx_texts_name_user_id_unique" {
			return &item, ErrConflictUnique
		}

		rep.log.Error("ошибка при создании записи текста в БД", slog.Any("error", err))
		return &item, errors.New("ошибка при создании записи текста в БД")
	}

	item.Text = text

	return &item, nil
}

// TextList получение списка записей пользователя с возможной фильтрацией по имени и пагинацией
func (rep DB) TextList(ctx context.Context, userID, name, page string) ([]models.Text, error) {
	items := make([]models.Text, 0)

	query := `SELECT id, user_id, name, text FROM texts WHERE user_id = $1`
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
		rep.log.Error("ошибка при получении списка текстов из БД", slog.Any("error", err))
		return nil, errors.New("ошибка при получении списка текстов из БД")
	}

	for i := range items {
		textCiphertext, err := hex.DecodeString(items[i].Text)
		if err != nil {
			rep.log.Error("Ошибка при декодировании hex поля", slog.Any("error", err))
			return nil, err
		}

		textDecrypt, err := rep.cipher.Decrypt(textCiphertext)
		if err != nil {
			rep.log.Error("Ошибка при декодировании поля", slog.Any("error", err))
			// TODO Переделай, не отправляй ошибку в клиент
			return nil, err
		}

		items[i].Text = string(textDecrypt)
	}

	return items, nil
}

// TextUpdate обновление пароля по логину пользователя
func (rep DB) TextUpdate(ctx context.Context, userID, id, name, text string) (*models.Text, error) {
	var item models.Text

	textEncrypted, err := rep.cipher.Encrypt([]byte(text))
	if err != nil {
		return nil, err
	}

	query := `UPDATE texts SET name = $1, text = $2 WHERE id = $3 AND user_id = $4 RETURNING id, user_id, name, text, created_at`

	err = rep.db.GetContext(ctx, &item, query, name, fmt.Sprintf("%x", textEncrypted), id, userID)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			rep.log.Error("ошибка при обновлении пароля в БД", slog.Any("error", err))
		}
		return &item, err
	}

	item.Text = text

	return &item, nil
}

// TextDelete удаление записи пароля по идентификатору
func (rep DB) TextDelete(ctx context.Context, userID, id string) (*models.Text, error) {
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
