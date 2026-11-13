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

func TestRepository_PasswordList(t *testing.T) {
	t.Run("success without search", func(t *testing.T) {
		rep, mock, cipher := setupTestRepo(t)

		encLogin, _ := cipher.EncryptStr("john_doe")
		encPass, _ := cipher.EncryptStr("my_secret_pass")

		rows := sqlmock.NewRows([]string{"id", "user_id", "name", "login", "password", "created_at"}).
			AddRow("pass-1", "user-1", "Github", encLogin, encPass, time.Now())

		mock.ExpectQuery(`^SELECT id, user_id, name, login, password, created_at FROM passwords WHERE user_id = \$1 ORDER BY created_at DESC$`).
			WithArgs("user-1").
			WillReturnRows(rows)

		passwords, err := rep.PasswordList(context.Background(), "user-1", "", "")
		require.NoError(t, err)
		require.Len(t, passwords, 1)
		assert.Equal(t, "pass-1", passwords[0].Id)
		assert.Equal(t, "user-1", passwords[0].UserId)
		assert.Equal(t, "Github", passwords[0].Name)
		assert.Equal(t, "john_doe", passwords[0].Login)
		assert.Equal(t, "my_secret_pass", passwords[0].Password)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success with search and page", func(t *testing.T) {
		rep, mock, _ := setupTestRepo(t)

		rows := sqlmock.NewRows([]string{"id", "user_id", "name", "login", "password", "created_at"})

		mock.ExpectQuery(`^SELECT id, user_id, name, login, password, created_at FROM passwords WHERE user_id = \$1 AND name ILIKE \$2 ORDER BY created_at DESC LIMIT 20 OFFSET 20$`).
			WithArgs("user-1", "%Git%").
			WillReturnRows(rows)

		passwords, err := rep.PasswordList(context.Background(), "user-1", "Git", "2")
		require.NoError(t, err)
		assert.Empty(t, passwords)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("db error", func(t *testing.T) {
		rep, mock, _ := setupTestRepo(t)

		mock.ExpectQuery(`^SELECT id, user_id, name, login, password, created_at FROM passwords`).
			WithArgs("user-1").
			WillReturnError(sql.ErrConnDone)

		passwords, err := rep.PasswordList(context.Background(), "user-1", "", "")
		require.Error(t, err)
		assert.True(t, errors.Is(err, ErrExec))
		assert.Nil(t, passwords)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestRepository_PasswordCreate(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		rep, mock, cipher := setupTestRepo(t)

		item := &models.Password{
			UserId:   "user-1",
			Name:     "Google Account",
			Login:    "test@gmail.com",
			Password: "superpassword123",
		}

		encLogin, _ := cipher.EncryptStr(item.Login)
		encPass, _ := cipher.EncryptStr(item.Password)

		rows := sqlmock.NewRows([]string{"id", "user_id", "name", "login", "password", "created_at"}).
			AddRow("pass-uuid-1", "user-1", "Google Account", encLogin, encPass, time.Now())

		mock.ExpectPrepare(`^INSERT INTO passwords \(user_id, name, login, password\) VALUES \(\$1, \$2, \$3, \$4\) RETURNING id, user_id, name, login, password, created_at$`).
			ExpectQuery().
			WithArgs("user-1", "Google Account", sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnRows(rows)

		err := rep.PasswordCreate(context.Background(), item)
		require.NoError(t, err)
		assert.Equal(t, "pass-uuid-1", item.Id)
		assert.Equal(t, "test@gmail.com", item.Login)
		assert.Equal(t, "superpassword123", item.Password)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("duplicate name error - ErrConflictUnique", func(t *testing.T) {
		rep, mock, _ := setupTestRepo(t)

		item := &models.Password{
			UserId:   "user-1",
			Name:     "Google Account",
			Login:    "test@gmail.com",
			Password: "superpassword123",
		}

		pgErr := &pgconn.PgError{
			Code:           pgerrcode.UniqueViolation,
			ConstraintName: "idx_passwords_name_user_id_unique",
		}

		mock.ExpectPrepare(`^INSERT INTO passwords`).
			ExpectQuery().
			WithArgs("user-1", "Google Account", sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnError(pgErr)

		err := rep.PasswordCreate(context.Background(), item)
		require.Error(t, err)
		assert.True(t, errors.Is(err, ErrConflictUnique))
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("other db error - ErrExec", func(t *testing.T) {
		rep, mock, _ := setupTestRepo(t)

		item := &models.Password{
			UserId:   "user-1",
			Name:     "Google Account",
			Login:    "test@gmail.com",
			Password: "superpassword123",
		}

		mock.ExpectPrepare(`^INSERT INTO passwords`).
			ExpectQuery().
			WithArgs("user-1", "Google Account", sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnError(sql.ErrConnDone)

		err := rep.PasswordCreate(context.Background(), item)
		require.Error(t, err)
		assert.True(t, errors.Is(err, ErrExec))
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestRepository_PasswordUpdate(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		rep, mock, cipher := setupTestRepo(t)

		item := &models.Password{
			Id:       "pass-1",
			UserId:   "user-1",
			Name:     "Updated Google Account",
			Login:    "new@gmail.com",
			Password: "newpassword123",
		}

		encLogin, _ := cipher.EncryptStr(item.Login)
		encPass, _ := cipher.EncryptStr(item.Password)

		rows := sqlmock.NewRows([]string{"id", "user_id", "name", "login", "password", "created_at"}).
			AddRow("pass-1", "user-1", "Updated Google Account", encLogin, encPass, time.Now())

		mock.ExpectPrepare(`^UPDATE passwords SET name = \$1, login = \$2, password = \$3 WHERE id = \$4 AND user_id = \$5 RETURNING id, user_id, name, login, password, created_at$`).
			ExpectQuery().
			WithArgs("Updated Google Account", sqlmock.AnyArg(), sqlmock.AnyArg(), "pass-1", "user-1").
			WillReturnRows(rows)

		err := rep.PasswordUpdate(context.Background(), item)
		require.NoError(t, err)
		assert.Equal(t, "Updated Google Account", item.Name)
		assert.Equal(t, "new@gmail.com", item.Login)
		assert.Equal(t, "newpassword123", item.Password)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("duplicate name error - ErrConflictUnique", func(t *testing.T) {
		rep, mock, _ := setupTestRepo(t)

		item := &models.Password{
			Id:       "pass-1",
			UserId:   "user-1",
			Name:     "Updated Google Account",
			Login:    "new@gmail.com",
			Password: "newpassword123",
		}

		pgErr := &pgconn.PgError{
			Code:           pgerrcode.UniqueViolation,
			ConstraintName: "idx_passwords_name_user_id_unique",
		}

		mock.ExpectPrepare(`^UPDATE passwords SET`).
			ExpectQuery().
			WithArgs("Updated Google Account", sqlmock.AnyArg(), sqlmock.AnyArg(), "pass-1", "user-1").
			WillReturnError(pgErr)

		err := rep.PasswordUpdate(context.Background(), item)
		require.Error(t, err)
		assert.True(t, errors.Is(err, ErrConflictUnique))
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("other db error - ErrExec", func(t *testing.T) {
		rep, mock, _ := setupTestRepo(t)

		item := &models.Password{
			Id:       "pass-1",
			UserId:   "user-1",
			Name:     "Updated Google Account",
			Login:    "new@gmail.com",
			Password: "newpassword123",
		}

		mock.ExpectPrepare(`^UPDATE passwords SET`).
			ExpectQuery().
			WithArgs("Updated Google Account", sqlmock.AnyArg(), sqlmock.AnyArg(), "pass-1", "user-1").
			WillReturnError(sql.ErrConnDone)

		err := rep.PasswordUpdate(context.Background(), item)
		require.Error(t, err)
		assert.True(t, errors.Is(err, ErrExec))
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestRepository_PasswordDelete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		rep, mock, _ := setupTestRepo(t)

		rows := sqlmock.NewRows([]string{"id", "user_id", "name", "login", "password", "created_at"}).
			AddRow("pass-1", "user-1", "Google Account", "enc_login", "enc_pass", time.Now())

		mock.ExpectQuery(`^DELETE FROM passwords WHERE user_id = \$1 AND id = \$2 RETURNING id, user_id, name, login, password, created_at$`).
			WithArgs("user-1", "pass-1").
			WillReturnRows(rows)

		item, err := rep.PasswordDelete(context.Background(), "user-1", "pass-1")
		require.NoError(t, err)
		require.NotNil(t, item)
		assert.Equal(t, "pass-1", item.Id)
		assert.Equal(t, "user-1", item.UserId)
		assert.Equal(t, "Google Account", item.Name)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found - sql.ErrNoRows", func(t *testing.T) {
		rep, mock, _ := setupTestRepo(t)

		mock.ExpectQuery(`^DELETE FROM passwords WHERE user_id = \$1 AND id = \$2 RETURNING`).
			WithArgs("user-1", "pass-1").
			WillReturnError(sql.ErrNoRows)

		item, err := rep.PasswordDelete(context.Background(), "user-1", "pass-1")
		require.Error(t, err)
		assert.True(t, errors.Is(err, sql.ErrNoRows))
		assert.NotNil(t, item)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("other db error", func(t *testing.T) {
		rep, mock, _ := setupTestRepo(t)

		mock.ExpectQuery(`^DELETE FROM passwords WHERE user_id = \$1 AND id = \$2 RETURNING`).
			WithArgs("user-1", "pass-1").
			WillReturnError(sql.ErrConnDone)

		item, err := rep.PasswordDelete(context.Background(), "user-1", "pass-1")
		require.Error(t, err)
		assert.True(t, errors.Is(err, sql.ErrConnDone))
		assert.NotNil(t, item)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}
