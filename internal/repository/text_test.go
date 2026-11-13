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

func TestRepository_TextList(t *testing.T) {
	t.Run("success without search", func(t *testing.T) {
		rep, mock, cipher := setupTestRepo(t)

		encText, _ := cipher.EncryptStr("Important notes here")

		rows := sqlmock.NewRows([]string{"id", "user_id", "name", "text"}).
			AddRow("text-1", "user-1", "My Note", encText)

		mock.ExpectQuery(`^SELECT id, user_id, name, text FROM texts WHERE user_id = \$1 ORDER BY created_at DESC$`).
			WithArgs("user-1").
			WillReturnRows(rows)

		texts, err := rep.TextList(context.Background(), "user-1", "", "")
		require.NoError(t, err)
		require.Len(t, texts, 1)
		assert.Equal(t, "text-1", texts[0].Id)
		assert.Equal(t, "user-1", texts[0].UserId)
		assert.Equal(t, "My Note", texts[0].Name)
		assert.Equal(t, "Important notes here", texts[0].Text)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success with search and page", func(t *testing.T) {
		rep, mock, _ := setupTestRepo(t)

		rows := sqlmock.NewRows([]string{"id", "user_id", "name", "text"})

		mock.ExpectQuery(`^SELECT id, user_id, name, text FROM texts WHERE user_id = \$1 AND name ILIKE \$2 ORDER BY created_at DESC LIMIT 20 OFFSET 0$`).
			WithArgs("user-1", "%Note%").
			WillReturnRows(rows)

		texts, err := rep.TextList(context.Background(), "user-1", "Note", "1")
		require.NoError(t, err)
		assert.Empty(t, texts)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("db error", func(t *testing.T) {
		rep, mock, _ := setupTestRepo(t)

		mock.ExpectQuery(`^SELECT id, user_id, name, text FROM texts`).
			WithArgs("user-1").
			WillReturnError(sql.ErrConnDone)

		texts, err := rep.TextList(context.Background(), "user-1", "", "")
		require.Error(t, err)
		assert.True(t, errors.Is(err, ErrExec))
		assert.Nil(t, texts)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestRepository_TextCreate(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		rep, mock, cipher := setupTestRepo(t)

		item := &models.Text{
			UserId: "user-1",
			Name:   "Meeting Notes",
			Text:   "Discussed architecture roadmap",
		}

		encText, _ := cipher.EncryptStr(item.Text)

		rows := sqlmock.NewRows([]string{"id", "user_id", "name", "text", "created_at"}).
			AddRow("text-uuid-1", "user-1", "Meeting Notes", encText, time.Now())

		mock.ExpectPrepare(`^INSERT INTO texts \(user_id, name, text\) VALUES \(\$1, \$2, \$3\) RETURNING id, user_id, name, text, created_at$`).
			ExpectQuery().
			WithArgs("user-1", "Meeting Notes", sqlmock.AnyArg()).
			WillReturnRows(rows)

		err := rep.TextCreate(context.Background(), item)
		require.NoError(t, err)
		assert.Equal(t, "text-uuid-1", item.Id)
		assert.Equal(t, "Discussed architecture roadmap", item.Text)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("duplicate name error - ErrConflictUnique", func(t *testing.T) {
		rep, mock, _ := setupTestRepo(t)

		item := &models.Text{
			UserId: "user-1",
			Name:   "Meeting Notes",
			Text:   "Discussed architecture roadmap",
		}

		pgErr := &pgconn.PgError{
			Code:           pgerrcode.UniqueViolation,
			ConstraintName: "idx_texts_name_user_id_unique",
		}

		mock.ExpectPrepare(`^INSERT INTO texts`).
			ExpectQuery().
			WithArgs("user-1", "Meeting Notes", sqlmock.AnyArg()).
			WillReturnError(pgErr)

		err := rep.TextCreate(context.Background(), item)
		require.Error(t, err)
		assert.True(t, errors.Is(err, ErrConflictUnique))
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("other db error - ErrExec", func(t *testing.T) {
		rep, mock, _ := setupTestRepo(t)

		item := &models.Text{
			UserId: "user-1",
			Name:   "Meeting Notes",
			Text:   "Discussed architecture roadmap",
		}

		mock.ExpectPrepare(`^INSERT INTO texts`).
			ExpectQuery().
			WithArgs("user-1", "Meeting Notes", sqlmock.AnyArg()).
			WillReturnError(sql.ErrConnDone)

		err := rep.TextCreate(context.Background(), item)
		require.Error(t, err)
		assert.True(t, errors.Is(err, ErrExec))
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestRepository_TextUpdate(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		rep, mock, cipher := setupTestRepo(t)

		item := &models.Text{
			Id:     "text-1",
			UserId: "user-1",
			Name:   "Updated Notes",
			Text:   "Updated text content",
		}

		encText, _ := cipher.EncryptStr(item.Text)

		rows := sqlmock.NewRows([]string{"id", "user_id", "name", "text", "created_at"}).
			AddRow("text-1", "user-1", "Updated Notes", encText, time.Now())

		mock.ExpectPrepare(`^UPDATE texts SET name = \$1, text = \$2 WHERE id = \$3 AND user_id = \$4 RETURNING id, user_id, name, text, created_at$`).
			ExpectQuery().
			WithArgs("Updated Notes", sqlmock.AnyArg(), "text-1", "user-1").
			WillReturnRows(rows)

		err := rep.TextUpdate(context.Background(), item)
		require.NoError(t, err)
		assert.Equal(t, "Updated Notes", item.Name)
		assert.Equal(t, "Updated text content", item.Text)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("duplicate name error - ErrConflictUnique", func(t *testing.T) {
		rep, mock, _ := setupTestRepo(t)

		item := &models.Text{
			Id:     "text-1",
			UserId: "user-1",
			Name:   "Updated Notes",
			Text:   "Updated text content",
		}

		pgErr := &pgconn.PgError{
			Code:           pgerrcode.UniqueViolation,
			ConstraintName: "idx_texts_name_user_id_unique",
		}

		mock.ExpectPrepare(`^UPDATE texts SET`).
			ExpectQuery().
			WithArgs("Updated Notes", sqlmock.AnyArg(), "text-1", "user-1").
			WillReturnError(pgErr)

		err := rep.TextUpdate(context.Background(), item)
		require.Error(t, err)
		assert.True(t, errors.Is(err, ErrConflictUnique))
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("other db error - ErrExec", func(t *testing.T) {
		rep, mock, _ := setupTestRepo(t)

		item := &models.Text{
			Id:     "text-1",
			UserId: "user-1",
			Name:   "Updated Notes",
			Text:   "Updated text content",
		}

		mock.ExpectPrepare(`^UPDATE texts SET`).
			ExpectQuery().
			WithArgs("Updated Notes", sqlmock.AnyArg(), "text-1", "user-1").
			WillReturnError(sql.ErrConnDone)

		err := rep.TextUpdate(context.Background(), item)
		require.Error(t, err)
		assert.True(t, errors.Is(err, ErrExec))
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestRepository_TextDelete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		rep, mock, _ := setupTestRepo(t)

		rows := sqlmock.NewRows([]string{"id", "user_id", "name", "text", "created_at"}).
			AddRow("text-1", "user-1", "Meeting Notes", "enc_text", time.Now())

		mock.ExpectQuery(`^DELETE FROM texts WHERE user_id = \$1 AND id = \$2 RETURNING id, user_id, name, text, created_at$`).
			WithArgs("user-1", "text-1").
			WillReturnRows(rows)

		item, err := rep.TextDelete(context.Background(), "user-1", "text-1")
		require.NoError(t, err)
		require.NotNil(t, item)
		assert.Equal(t, "text-1", item.Id)
		assert.Equal(t, "user-1", item.UserId)
		assert.Equal(t, "Meeting Notes", item.Name)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found - sql.ErrNoRows", func(t *testing.T) {
		rep, mock, _ := setupTestRepo(t)

		mock.ExpectQuery(`^DELETE FROM texts WHERE user_id = \$1 AND id = \$2 RETURNING`).
			WithArgs("user-1", "text-1").
			WillReturnError(sql.ErrNoRows)

		item, err := rep.TextDelete(context.Background(), "user-1", "text-1")
		require.Error(t, err)
		assert.True(t, errors.Is(err, sql.ErrNoRows))
		assert.NotNil(t, item)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("other db error", func(t *testing.T) {
		rep, mock, _ := setupTestRepo(t)

		mock.ExpectQuery(`^DELETE FROM texts WHERE user_id = \$1 AND id = \$2 RETURNING`).
			WithArgs("user-1", "text-1").
			WillReturnError(sql.ErrConnDone)

		item, err := rep.TextDelete(context.Background(), "user-1", "text-1")
		require.Error(t, err)
		assert.True(t, errors.Is(err, sql.ErrConnDone))
		assert.NotNil(t, item)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}
