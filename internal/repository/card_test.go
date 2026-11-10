package repository

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/alexsey-popov/gkeep/internal/models"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRepository_CardList(t *testing.T) {
	t.Run("success without search", func(t *testing.T) {
		rep, mock, cipher := setupTestRepo(t)

		encNumber, _ := cipher.EncryptStr("1234567812345678")
		encExpired, _ := cipher.EncryptStr("1225")
		encCode, _ := cipher.EncryptStr("123")

		rows := sqlmock.NewRows([]string{"id", "user_id", "name", "number", "expired", "code"}).
			AddRow("card-1", "user-1", "Main Card", encNumber, encExpired, encCode)

		mock.ExpectQuery(`^SELECT id, user_id, name, number, expired, code FROM cards WHERE user_id = \$1 ORDER BY created_at DESC$`).
			WithArgs("user-1").
			WillReturnRows(rows)

		cards, err := rep.CardList(context.Background(), "user-1", "", "")
		require.NoError(t, err)
		require.Len(t, cards, 1)
		assert.Equal(t, "card-1", cards[0].Id)
		assert.Equal(t, "user-1", cards[0].UserId)
		assert.Equal(t, "Main Card", cards[0].Name)
		assert.Equal(t, "1234567812345678", cards[0].Number)
		assert.Equal(t, "1225", cards[0].Expired)
		assert.Equal(t, "123", cards[0].Code)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success with search and page", func(t *testing.T) {
		rep, mock, _ := setupTestRepo(t)

		rows := sqlmock.NewRows([]string{"id", "user_id", "name", "number", "expired", "code"})

		mock.ExpectQuery(`^SELECT id, user_id, name, number, expired, code FROM cards WHERE user_id = \$1 AND name ILIKE \$2 ORDER BY created_at DESC LIMIT 20 OFFSET 0$`).
			WithArgs("user-1", "%Visa%").
			WillReturnRows(rows)

		cards, err := rep.CardList(context.Background(), "user-1", "Visa", "1")
		require.NoError(t, err)
		assert.Empty(t, cards)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("db error", func(t *testing.T) {
		rep, mock, _ := setupTestRepo(t)

		mock.ExpectQuery(`^SELECT id, user_id, name, number, expired, code FROM cards`).
			WithArgs("user-1").
			WillReturnError(sql.ErrConnDone)

		cards, err := rep.CardList(context.Background(), "user-1", "", "")
		require.Error(t, err)
		assert.True(t, errors.Is(err, ErrExec))
		assert.Nil(t, cards)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestRepository_CardCreate(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		rep, mock, cipher := setupTestRepo(t)

		card := &models.Card{
			UserId:  "user-1",
			Name:    "Tinkoff Black",
			Number:  "2200112233445566",
			Expired: "1126",
			Code:    "777",
		}

		encNumber, _ := cipher.EncryptStr(card.Number)
		encExpired, _ := cipher.EncryptStr(card.Expired)
		encCode, _ := cipher.EncryptStr(card.Code)

		rows := sqlmock.NewRows([]string{"id", "user_id", "name", "number", "expired", "code", "created_at"}).
			AddRow("card-uuid-1", "user-1", "Tinkoff Black", encNumber, encExpired, encCode, time.Now())

		mock.ExpectPrepare(`^INSERT INTO cards \(user_id, name, number, expired, code\) VALUES \(\$1, \$2, \$3, \$4, \$5\) RETURNING id, user_id, name, number, expired, code, created_at$`).
			ExpectQuery().
			WithArgs("user-1", "Tinkoff Black", sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnRows(rows)

		err := rep.CardCreate(context.Background(), card)
		require.NoError(t, err)
		assert.Equal(t, "card-uuid-1", card.Id)
		assert.Equal(t, "2200112233445566", card.Number)
		assert.Equal(t, "1126", card.Expired)
		assert.Equal(t, "777", card.Code)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("duplicate name error - ErrConflictUnique", func(t *testing.T) {
		rep, mock, _ := setupTestRepo(t)

		card := &models.Card{
			UserId:  "user-1",
			Name:    "Tinkoff Black",
			Number:  "2200112233445566",
			Expired: "1126",
			Code:    "777",
		}

		pgErr := &pgconn.PgError{
			Code:           pgerrcode.UniqueViolation,
			ConstraintName: "idx_cards_name_user_id_unique",
		}

		mock.ExpectPrepare(`^INSERT INTO cards`).
			ExpectQuery().
			WithArgs("user-1", "Tinkoff Black", sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnError(pgErr)

		err := rep.CardCreate(context.Background(), card)
		require.Error(t, err)
		assert.True(t, errors.Is(err, ErrConflictUnique))
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("other db error - ErrExec", func(t *testing.T) {
		rep, mock, _ := setupTestRepo(t)

		card := &models.Card{
			UserId:  "user-1",
			Name:    "Tinkoff Black",
			Number:  "2200112233445566",
			Expired: "1126",
			Code:    "777",
		}

		mock.ExpectPrepare(`^INSERT INTO cards`).
			ExpectQuery().
			WithArgs("user-1", "Tinkoff Black", sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnError(sql.ErrConnDone)

		err := rep.CardCreate(context.Background(), card)
		require.Error(t, err)
		assert.True(t, errors.Is(err, ErrExec))
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestRepository_CardUpdate(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		rep, mock, cipher := setupTestRepo(t)

		card := &models.Card{
			Id:      "card-1",
			UserId:  "user-1",
			Name:    "Updated Card",
			Number:  "2200112233445566",
			Expired: "1228",
			Code:    "888",
		}

		encNumber, _ := cipher.EncryptStr(card.Number)
		encExpired, _ := cipher.EncryptStr(card.Expired)
		encCode, _ := cipher.EncryptStr(card.Code)

		rows := sqlmock.NewRows([]string{"id", "user_id", "name", "number", "expired", "code", "created_at"}).
			AddRow("card-1", "user-1", "Updated Card", encNumber, encExpired, encCode, time.Now())

		mock.ExpectPrepare(`^UPDATE cards SET name = \$1, number = \$2, expired = \$3, code = \$4 WHERE id = \$5 AND user_id = \$6 RETURNING id, user_id, name, number, expired, code, created_at$`).
			ExpectQuery().
			WithArgs("Updated Card", sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), "card-1", "user-1").
			WillReturnRows(rows)

		err := rep.CardUpdate(context.Background(), card)
		require.NoError(t, err)
		assert.Equal(t, "Updated Card", card.Name)
		assert.Equal(t, "2200112233445566", card.Number)
		assert.Equal(t, "1228", card.Expired)
		assert.Equal(t, "888", card.Code)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("duplicate name error - ErrConflictUnique", func(t *testing.T) {
		rep, mock, _ := setupTestRepo(t)

		card := &models.Card{
			Id:      "card-1",
			UserId:  "user-1",
			Name:    "Updated Card",
			Number:  "2200112233445566",
			Expired: "1228",
			Code:    "888",
		}

		pgErr := &pgconn.PgError{
			Code:           pgerrcode.UniqueViolation,
			ConstraintName: "idx_cards_name_user_id_unique",
		}

		mock.ExpectPrepare(`^UPDATE cards SET`).
			ExpectQuery().
			WithArgs("Updated Card", sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), "card-1", "user-1").
			WillReturnError(pgErr)

		err := rep.CardUpdate(context.Background(), card)
		require.Error(t, err)
		assert.True(t, errors.Is(err, ErrConflictUnique))
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("other db error - ErrExec", func(t *testing.T) {
		rep, mock, _ := setupTestRepo(t)

		card := &models.Card{
			Id:      "card-1",
			UserId:  "user-1",
			Name:    "Updated Card",
			Number:  "2200112233445566",
			Expired: "1228",
			Code:    "888",
		}

		mock.ExpectPrepare(`^UPDATE cards SET`).
			ExpectQuery().
			WithArgs("Updated Card", sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), "card-1", "user-1").
			WillReturnError(sql.ErrConnDone)

		err := rep.CardUpdate(context.Background(), card)
		require.Error(t, err)
		assert.True(t, errors.Is(err, ErrExec))
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestRepository_CardDelete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		rep, mock, _ := setupTestRepo(t)

		rows := sqlmock.NewRows([]string{"id", "user_id", "name", "number", "expired", "code", "created_at"}).
			AddRow("card-1", "user-1", "Card to delete", "enc_num", "enc_exp", "enc_code", time.Now())

		mock.ExpectQuery(`^DELETE FROM cards WHERE user_id = \$1 AND id = \$2 RETURNING id, user_id, name, number, expired, code, created_at$`).
			WithArgs("user-1", "card-1").
			WillReturnRows(rows)

		item, err := rep.CardDelete(context.Background(), "user-1", "card-1")
		require.NoError(t, err)
		require.NotNil(t, item)
		assert.Equal(t, "card-1", item.Id)
		assert.Equal(t, "user-1", item.UserId)
		assert.Equal(t, "Card to delete", item.Name)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found - sql.ErrNoRows", func(t *testing.T) {
		rep, mock, _ := setupTestRepo(t)

		mock.ExpectQuery(`^DELETE FROM cards WHERE user_id = \$1 AND id = \$2 RETURNING`).
			WithArgs("user-1", "card-1").
			WillReturnError(sql.ErrNoRows)

		item, err := rep.CardDelete(context.Background(), "user-1", "card-1")
		require.Error(t, err)
		assert.True(t, errors.Is(err, sql.ErrNoRows))
		assert.NotNil(t, item)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("other db error", func(t *testing.T) {
		rep, mock, _ := setupTestRepo(t)

		mock.ExpectQuery(`^DELETE FROM cards WHERE user_id = \$1 AND id = \$2 RETURNING`).
			WithArgs("user-1", "card-1").
			WillReturnError(sql.ErrConnDone)

		item, err := rep.CardDelete(context.Background(), "user-1", "card-1")
		require.Error(t, err)
		assert.True(t, errors.Is(err, sql.ErrConnDone))
		assert.NotNil(t, item)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}
