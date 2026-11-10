package repository

import (
	"context"
	"database/sql"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/alexsey-popov/gkeep/internal/crypto"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestRepo(t *testing.T) (*Repository, sqlmock.Sqlmock, *crypto.Service) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	sqlxDB := sqlx.NewDb(db, "postgres")
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	cipher, err := crypto.New([]byte("12345678901234567890123456789012"))
	require.NoError(t, err)

	rep := New(sqlxDB, logger, cipher)
	return rep, mock, cipher
}

type mockCipherItem struct {
	Id         string    `db:"id"`
	Data       string    `db:"data"`
	CreatedAt  time.Time `db:"created_at"`
	encryptErr error
	decryptErr error
}

func (m *mockCipherItem) Encrypt(cipher *crypto.Service) error {
	if m.encryptErr != nil {
		return m.encryptErr
	}
	enc, err := cipher.EncryptStr(m.Data)
	if err != nil {
		return err
	}
	m.Data = enc
	return nil
}

func (m *mockCipherItem) Decrypt(cipher *crypto.Service) error {
	if m.decryptErr != nil {
		return m.decryptErr
	}
	dec, err := cipher.DecryptStr(m.Data)
	if err != nil {
		return err
	}
	m.Data = dec
	return nil
}

func TestNew(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	sqlxDB := sqlx.NewDb(db, "postgres")
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	cipher, err := crypto.New([]byte("12345678901234567890123456789012"))
	require.NoError(t, err)

	rep := New(sqlxDB, logger, cipher)
	require.NotNil(t, rep)
	assert.Equal(t, sqlxDB, rep.db)
	assert.Equal(t, logger, rep.log)
	assert.Equal(t, cipher, rep.cipher)
}

func TestRepository_PrepareSearchAndPaginate(t *testing.T) {
	rep, _, _ := setupTestRepo(t)

	t.Run("without search and without page", func(t *testing.T) {
		baseQuery := "SELECT * FROM texts WHERE user_id = $1"
		baseArgs := []any{"user-1"}

		query, args := rep.PrepareSearchAndPaginate(baseQuery, baseArgs, "", "")
		assert.Equal(t, "SELECT * FROM texts WHERE user_id = $1 ORDER BY created_at DESC", query)
		assert.Equal(t, []any{"user-1"}, args)
	})

	t.Run("with search and valid page", func(t *testing.T) {
		baseQuery := "SELECT * FROM texts WHERE user_id = $1"
		baseArgs := []any{"user-1"}

		query, args := rep.PrepareSearchAndPaginate(baseQuery, baseArgs, "test", "2")
		assert.Equal(t, "SELECT * FROM texts WHERE user_id = $1 AND name ILIKE $2 ORDER BY created_at DESC LIMIT 20 OFFSET 20", query)
		assert.Equal(t, []any{"user-1", "%test%"}, args)
	})

	t.Run("with invalid page (non-number and negative)", func(t *testing.T) {
		baseQuery := "SELECT * FROM texts WHERE user_id = $1"
		baseArgs := []any{"user-1"}

		query, args := rep.PrepareSearchAndPaginate(baseQuery, baseArgs, "", "invalid")
		assert.Equal(t, "SELECT * FROM texts WHERE user_id = $1 ORDER BY created_at DESC", query)
		assert.Equal(t, []any{"user-1"}, args)

		query, args = rep.PrepareSearchAndPaginate(baseQuery, baseArgs, "", "-1")
		assert.Equal(t, "SELECT * FROM texts WHERE user_id = $1 ORDER BY created_at DESC", query)
		assert.Equal(t, []any{"user-1"}, args)
	})
}

func TestRepository_GetMany(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		rep, mock, cipher := setupTestRepo(t)

		encData1, err := cipher.EncryptStr("secret1")
		require.NoError(t, err)
		encData2, err := cipher.EncryptStr("secret2")
		require.NoError(t, err)

		rows := sqlmock.NewRows([]string{"id", "data", "created_at"}).
			AddRow("id-1", encData1, time.Now()).
			AddRow("id-2", encData2, time.Now())

		mock.ExpectQuery(`^SELECT id, data, created_at FROM texts WHERE user_id = \$1$`).
			WithArgs("user-1").
			WillReturnRows(rows)

		items, err := rep.GetMany[*mockCipherItem](context.Background(), "SELECT id, data, created_at FROM texts WHERE user_id = $1", "user-1")
		require.NoError(t, err)
		require.Len(t, items, 2)
		assert.Equal(t, "secret1", items[0].Data)
		assert.Equal(t, "secret2", items[1].Data)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("select error", func(t *testing.T) {
		rep, mock, _ := setupTestRepo(t)

		mock.ExpectQuery(`^SELECT id FROM texts$`).
			WillReturnError(errors.New("db error"))

		items, err := rep.GetMany[*mockCipherItem](context.Background(), "SELECT id FROM texts")
		require.Error(t, err)
		assert.True(t, errors.Is(err, ErrExec))
		assert.Nil(t, items)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("decrypt error", func(t *testing.T) {
		rep, mock, _ := setupTestRepo(t)

		// Данные не зашифрованы корректно в hex/GCM формате
		rows := sqlmock.NewRows([]string{"id", "data", "created_at"}).
			AddRow("id-1", "invalid_encrypted_data", time.Now())

		mock.ExpectQuery(`^SELECT id, data, created_at FROM texts$`).
			WillReturnRows(rows)

		items, err := rep.GetMany[*mockCipherItem](context.Background(), "SELECT id, data, created_at FROM texts")
		require.Error(t, err)
		assert.True(t, errors.Is(err, ErrDecrypt))
		assert.Len(t, items, 1)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestRepository_PutOne(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		rep, mock, cipher := setupTestRepo(t)

		item := &mockCipherItem{
			Id:   "id-1",
			Data: "plain-secret",
		}

		encData, err := cipher.EncryptStr("plain-secret")
		require.NoError(t, err)

		mock.ExpectPrepare(`^INSERT INTO texts \(id, data\) VALUES \(\$1, \$2\) RETURNING id, data, created_at$`).
			ExpectQuery().
			WithArgs("id-1", sqlmock.AnyArg()).
			WillReturnRows(sqlmock.NewRows([]string{"id", "data", "created_at"}).AddRow("id-1", encData, time.Now()))

		query := `INSERT INTO texts (id, data) VALUES (:id, :data) RETURNING id, data, created_at`
		err = rep.PutOne[*mockCipherItem](context.Background(), query, item)
		require.NoError(t, err)
		assert.Equal(t, "plain-secret", item.Data)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("encrypt error", func(t *testing.T) {
		rep, _, _ := setupTestRepo(t)

		item := &mockCipherItem{
			Id:         "id-1",
			Data:       "plain",
			encryptErr: errors.New("custom encrypt error"),
		}

		query := `INSERT INTO texts (id, data) VALUES (:id, :data) RETURNING id, data, created_at`
		err := rep.PutOne[*mockCipherItem](context.Background(), query, item)
		require.Error(t, err)
		assert.True(t, errors.Is(err, ErrEncrypt))
	})

	t.Run("prepare error", func(t *testing.T) {
		rep, mock, _ := setupTestRepo(t)

		item := &mockCipherItem{
			Id:   "id-1",
			Data: "plain",
		}

		mock.ExpectPrepare(`^INSERT INTO texts`).
			WillReturnError(errors.New("prepare failed"))

		query := `INSERT INTO texts (id, data) VALUES (:id, :data) RETURNING id, data, created_at`
		err := rep.PutOne[*mockCipherItem](context.Background(), query, item)
		require.Error(t, err)
		assert.True(t, errors.Is(err, ErrPrepare))
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("exec error", func(t *testing.T) {
		rep, mock, _ := setupTestRepo(t)

		item := &mockCipherItem{
			Id:   "id-1",
			Data: "plain",
		}

		mock.ExpectPrepare(`^INSERT INTO texts`).
			ExpectQuery().
			WithArgs("id-1", sqlmock.AnyArg()).
			WillReturnError(sql.ErrConnDone)

		query := `INSERT INTO texts (id, data) VALUES (:id, :data) RETURNING id, data, created_at`
		err := rep.PutOne[*mockCipherItem](context.Background(), query, item)
		require.Error(t, err)
		assert.True(t, errors.Is(err, ErrExec))
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("decrypt error", func(t *testing.T) {
		rep, mock, _ := setupTestRepo(t)

		item := &mockCipherItem{
			Id:   "id-1",
			Data: "plain",
		}

		mock.ExpectPrepare(`^INSERT INTO texts`).
			ExpectQuery().
			WithArgs("id-1", sqlmock.AnyArg()).
			WillReturnRows(sqlmock.NewRows([]string{"id", "data", "created_at"}).AddRow("id-1", "corrupted_cipher_data", time.Now()))

		query := `INSERT INTO texts (id, data) VALUES (:id, :data) RETURNING id, data, created_at`
		err := rep.PutOne[*mockCipherItem](context.Background(), query, item)
		require.Error(t, err)
		assert.True(t, errors.Is(err, ErrDecrypt))
		require.NoError(t, mock.ExpectationsWereMet())
	})
}
