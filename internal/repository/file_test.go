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

func TestRepository_FileList(t *testing.T) {
	t.Run("success without search", func(t *testing.T) {
		rep, mock, cipher := setupTestRepo(t)

		encFilename, _ := cipher.EncryptStr("passport.pdf")
		encStoragePath, _ := cipher.EncryptStr("/storage/files/user-1/passport.pdf")

		rows := sqlmock.NewRows([]string{"id", "user_id", "name", "filename", "storage_path", "size", "checksum", "created_at"}).
			AddRow("file-1", "user-1", "Passport Scan", encFilename, encStoragePath, int64(1024), "sha256checksum", time.Now())

		mock.ExpectQuery(`^SELECT id, user_id, name, filename, storage_path, size, checksum, created_at FROM files WHERE user_id = \$1 ORDER BY created_at DESC$`).
			WithArgs("user-1").
			WillReturnRows(rows)

		files, err := rep.FileList(context.Background(), "user-1", "", "")
		require.NoError(t, err)
		require.Len(t, files, 1)
		assert.Equal(t, "file-1", files[0].Id)
		assert.Equal(t, "user-1", files[0].UserId)
		assert.Equal(t, "Passport Scan", files[0].Name)
		assert.Equal(t, "passport.pdf", files[0].Filename)
		assert.Equal(t, "/storage/files/user-1/passport.pdf", files[0].StoragePath)
		assert.Equal(t, int64(1024), files[0].Size)
		assert.Equal(t, "sha256checksum", files[0].Checksum)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success with search and page", func(t *testing.T) {
		rep, mock, _ := setupTestRepo(t)

		rows := sqlmock.NewRows([]string{"id", "user_id", "name", "filename", "storage_path", "size", "checksum", "created_at"})

		mock.ExpectQuery(`^SELECT id, user_id, name, filename, storage_path, size, checksum, created_at FROM files WHERE user_id = \$1 AND name ILIKE \$2 ORDER BY created_at DESC LIMIT 20 OFFSET 0$`).
			WithArgs("user-1", "%Doc%").
			WillReturnRows(rows)

		files, err := rep.FileList(context.Background(), "user-1", "Doc", "1")
		require.NoError(t, err)
		assert.Empty(t, files)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("db error", func(t *testing.T) {
		rep, mock, _ := setupTestRepo(t)

		mock.ExpectQuery(`^SELECT id, user_id, name, filename, storage_path, size, checksum, created_at FROM files`).
			WithArgs("user-1").
			WillReturnError(sql.ErrConnDone)

		files, err := rep.FileList(context.Background(), "user-1", "", "")
		require.Error(t, err)
		assert.True(t, errors.Is(err, ErrExec))
		assert.Nil(t, files)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestRepository_FileCreate(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		rep, mock, cipher := setupTestRepo(t)

		item := &models.File{
			UserId:      "user-1",
			Name:        "ID Card Scan",
			Filename:    "id_card.png",
			StoragePath: "/storage/files/id_card.png",
			Size:        2048,
			Checksum:    "checksum123",
		}

		encFilename, _ := cipher.EncryptStr(item.Filename)
		encStoragePath, _ := cipher.EncryptStr(item.StoragePath)

		rows := sqlmock.NewRows([]string{"id", "user_id", "name", "filename", "storage_path", "size", "checksum", "created_at"}).
			AddRow("file-uuid-1", "user-1", "ID Card Scan", encFilename, encStoragePath, int64(2048), "checksum123", time.Now())

		mock.ExpectPrepare(`^INSERT INTO files \(user_id, name, filename, storage_path, size, checksum\) VALUES \(\$1, \$2, \$3, \$4, \$5, \$6\) RETURNING id, user_id, name, filename, storage_path, size, checksum, created_at$`).
			ExpectQuery().
			WithArgs("user-1", "ID Card Scan", sqlmock.AnyArg(), sqlmock.AnyArg(), int64(2048), "checksum123").
			WillReturnRows(rows)

		err := rep.FileCreate(context.Background(), item)
		require.NoError(t, err)
		assert.Equal(t, "file-uuid-1", item.Id)
		assert.Equal(t, "id_card.png", item.Filename)
		assert.Equal(t, "/storage/files/id_card.png", item.StoragePath)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("duplicate name error - ErrConflictUnique", func(t *testing.T) {
		rep, mock, _ := setupTestRepo(t)

		item := &models.File{
			UserId:      "user-1",
			Name:        "ID Card Scan",
			Filename:    "id_card.png",
			StoragePath: "/storage/files/id_card.png",
			Size:        2048,
			Checksum:    "checksum123",
		}

		pgErr := &pgconn.PgError{
			Code:           pgerrcode.UniqueViolation,
			ConstraintName: "idx_files_name_user_id_unique",
		}

		mock.ExpectPrepare(`^INSERT INTO files`).
			ExpectQuery().
			WithArgs("user-1", "ID Card Scan", sqlmock.AnyArg(), sqlmock.AnyArg(), int64(2048), "checksum123").
			WillReturnError(pgErr)

		err := rep.FileCreate(context.Background(), item)
		require.Error(t, err)
		assert.True(t, errors.Is(err, ErrConflictUnique))
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("other db error - ErrExec", func(t *testing.T) {
		rep, mock, _ := setupTestRepo(t)

		item := &models.File{
			UserId:      "user-1",
			Name:        "ID Card Scan",
			Filename:    "id_card.png",
			StoragePath: "/storage/files/id_card.png",
			Size:        2048,
			Checksum:    "checksum123",
		}

		mock.ExpectPrepare(`^INSERT INTO files`).
			ExpectQuery().
			WithArgs("user-1", "ID Card Scan", sqlmock.AnyArg(), sqlmock.AnyArg(), int64(2048), "checksum123").
			WillReturnError(sql.ErrConnDone)

		err := rep.FileCreate(context.Background(), item)
		require.Error(t, err)
		assert.True(t, errors.Is(err, ErrExec))
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestRepository_FileShow(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		rep, mock, cipher := setupTestRepo(t)

		encFilename, _ := cipher.EncryptStr("passport.pdf")
		encStoragePath, _ := cipher.EncryptStr("/storage/files/passport.pdf")

		rows := sqlmock.NewRows([]string{"id", "user_id", "name", "filename", "storage_path", "size", "checksum", "created_at"}).
			AddRow("file-1", "user-1", "Passport Scan", encFilename, encStoragePath, int64(5000), "checksum", time.Now())

		mock.ExpectPrepare(`^SELECT id, user_id, name, filename, storage_path, size, checksum, created_at FROM files WHERE user_id = \$1 AND id = \$2$`).
			ExpectQuery().
			WithArgs("user-1", "file-1").
			WillReturnRows(rows)

		file, err := rep.FileShow(context.Background(), "user-1", "file-1")
		require.NoError(t, err)
		require.NotNil(t, file)
		assert.Equal(t, "file-1", file.Id)
		assert.Equal(t, "user-1", file.UserId)
		assert.Equal(t, "Passport Scan", file.Name)
		assert.Equal(t, "passport.pdf", file.Filename)
		assert.Equal(t, "/storage/files/passport.pdf", file.StoragePath)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found - sql.ErrNoRows", func(t *testing.T) {
		rep, mock, _ := setupTestRepo(t)

		mock.ExpectPrepare(`^SELECT id, user_id, name, filename, storage_path, size, checksum, created_at FROM files WHERE user_id = \$1 AND id = \$2$`).
			ExpectQuery().
			WithArgs("user-1", "file-1").
			WillReturnError(sql.ErrNoRows)

		file, err := rep.FileShow(context.Background(), "user-1", "file-1")
		require.Error(t, err)
		assert.True(t, errors.Is(err, ErrNoRows))
		assert.NotNil(t, file)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("other db error", func(t *testing.T) {
		rep, mock, _ := setupTestRepo(t)

		mock.ExpectPrepare(`^SELECT id, user_id, name, filename, storage_path, size, checksum, created_at FROM files WHERE user_id = \$1 AND id = \$2$`).
			ExpectQuery().
			WithArgs("user-1", "file-1").
			WillReturnError(sql.ErrConnDone)

		file, err := rep.FileShow(context.Background(), "user-1", "file-1")
		require.Error(t, err)
		assert.True(t, errors.Is(err, ErrExec))
		assert.NotNil(t, file)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestRepository_FileDelete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		rep, mock, cipher := setupTestRepo(t)

		encFilename, _ := cipher.EncryptStr("filename.txt")
		encStoragePath, _ := cipher.EncryptStr("/storage/path")

		rows := sqlmock.NewRows([]string{"id", "user_id", "name", "filename", "storage_path", "size", "checksum", "created_at"}).
			AddRow("file-1", "user-1", "File to delete", encFilename, encStoragePath, int64(1024), "chk", time.Now())

		mock.ExpectQuery(`^DELETE FROM files WHERE user_id = \$1 AND id = \$2 RETURNING id, user_id, name, filename, storage_path, size, checksum, created_at$`).
			WithArgs("user-1", "file-1").
			WillReturnRows(rows)

		item, err := rep.FileDelete(context.Background(), "user-1", "file-1")
		require.NoError(t, err)
		require.NotNil(t, item)
		assert.Equal(t, "file-1", item.Id)
		assert.Equal(t, "user-1", item.UserId)
		assert.Equal(t, "File to delete", item.Name)
		assert.Equal(t, "filename.txt", item.Filename)
		assert.Equal(t, "/storage/path", item.StoragePath)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found - sql.ErrNoRows", func(t *testing.T) {
		rep, mock, _ := setupTestRepo(t)

		mock.ExpectQuery(`^DELETE FROM files WHERE user_id = \$1 AND id = \$2 RETURNING`).
			WithArgs("user-1", "file-1").
			WillReturnError(sql.ErrNoRows)

		item, err := rep.FileDelete(context.Background(), "user-1", "file-1")
		require.Error(t, err)
		assert.True(t, errors.Is(err, sql.ErrNoRows))
		assert.NotNil(t, item)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("other db error", func(t *testing.T) {
		rep, mock, _ := setupTestRepo(t)

		mock.ExpectQuery(`^DELETE FROM files WHERE user_id = \$1 AND id = \$2 RETURNING`).
			WithArgs("user-1", "file-1").
			WillReturnError(sql.ErrConnDone)

		item, err := rep.FileDelete(context.Background(), "user-1", "file-1")
		require.Error(t, err)
		assert.True(t, errors.Is(err, sql.ErrConnDone))
		assert.NotNil(t, item)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}
