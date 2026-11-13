package repository

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/alexsey-popov/gkeep/internal/models"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRepository_CreateUser(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		rep, mock, _ := setupTestRepo(t)

		user := &models.User{
			Login:    "testuser",
			Password: "hashedpassword",
		}

		rows := sqlmock.NewRows([]string{"id", "login", "password"}).
			AddRow("user-uuid-1", "testuser", "hashedpassword")

		mock.ExpectQuery(`^INSERT INTO users \(login, password\) VALUES \(\$1, \$2\) RETURNING \*$`).
			WithArgs("testuser", "hashedpassword").
			WillReturnRows(rows)

		err := rep.CreateUser(context.Background(), user)
		require.NoError(t, err)
		assert.Equal(t, "user-uuid-1", user.Id)
		assert.Equal(t, "testuser", user.Login)
		assert.Equal(t, "hashedpassword", user.Password)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("duplicate login error - ErrConflictUnique", func(t *testing.T) {
		rep, mock, _ := setupTestRepo(t)

		user := &models.User{
			Login:    "testuser",
			Password: "hashedpassword",
		}

		pgErr := &pgconn.PgError{
			Code:           pgerrcode.UniqueViolation,
			ConstraintName: "idx_users_login_unique",
		}

		mock.ExpectQuery(`^INSERT INTO users`).
			WithArgs("testuser", "hashedpassword").
			WillReturnError(pgErr)

		err := rep.CreateUser(context.Background(), user)
		require.Error(t, err)
		assert.True(t, errors.Is(err, ErrConflictUnique))
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("other database error", func(t *testing.T) {
		rep, mock, _ := setupTestRepo(t)

		user := &models.User{
			Login:    "testuser",
			Password: "hashedpassword",
		}

		mock.ExpectQuery(`^INSERT INTO users`).
			WithArgs("testuser", "hashedpassword").
			WillReturnError(sql.ErrConnDone)

		err := rep.CreateUser(context.Background(), user)
		require.Error(t, err)
		assert.Equal(t, "ошибка при создании нового пользователя в БД", err.Error())
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestRepository_GetUser(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		rep, mock, _ := setupTestRepo(t)

		rows := sqlmock.NewRows([]string{"id", "login", "password"}).
			AddRow("user-uuid-1", "testuser", "hashedpassword")

		mock.ExpectQuery(`^SELECT \* FROM users WHERE login = \$1 LIMIT 1$`).
			WithArgs("testuser").
			WillReturnRows(rows)

		user, err := rep.GetUser(context.Background(), "testuser")
		require.NoError(t, err)
		require.NotNil(t, user)
		assert.Equal(t, "user-uuid-1", user.Id)
		assert.Equal(t, "testuser", user.Login)
		assert.Equal(t, "hashedpassword", user.Password)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found - sql.ErrNoRows", func(t *testing.T) {
		rep, mock, _ := setupTestRepo(t)

		mock.ExpectQuery(`^SELECT \* FROM users WHERE login = \$1 LIMIT 1$`).
			WithArgs("nonexistent").
			WillReturnError(sql.ErrNoRows)

		user, err := rep.GetUser(context.Background(), "nonexistent")
		require.Error(t, err)
		assert.Equal(t, "ошибка при получении данных пользователя из БД", err.Error())
		assert.NotNil(t, user)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("other db error", func(t *testing.T) {
		rep, mock, _ := setupTestRepo(t)

		mock.ExpectQuery(`^SELECT \* FROM users WHERE login = \$1 LIMIT 1$`).
			WithArgs("testuser").
			WillReturnError(errors.New("connection failed"))

		user, err := rep.GetUser(context.Background(), "testuser")
		require.Error(t, err)
		assert.Equal(t, "ошибка при получении данных пользователя из БД", err.Error())
		assert.NotNil(t, user)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}
