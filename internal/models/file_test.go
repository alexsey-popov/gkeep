package models

import (
	"testing"
	"time"

	"github.com/alexsey-popov/gkeep/internal/crypto"
	"github.com/alexsey-popov/gkeep/internal/validator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFile_Encrypt_Decrypt(t *testing.T) {
	cipher, err := crypto.NewWithTempKey()
	require.NoError(t, err)

	t.Run("success encrypt and decrypt roundtrip", func(t *testing.T) {
		f := &File{
			Id:          "file-1",
			UserId:      "user-1",
			Name:        "Secret Document",
			Filename:    "passwords.docx",
			StoragePath: "/attachments/uuid-file.bin",
			Size:        1024,
			Checksum:    "d41d8cd98f00b204e9800998ecf8427e",
			CreatedAt:   time.Now(),
		}

		origFilename := f.Filename
		origStoragePath := f.StoragePath

		err := f.Encrypt(cipher)
		require.NoError(t, err)

		assert.NotEqual(t, origFilename, f.Filename)
		assert.NotEqual(t, origStoragePath, f.StoragePath)

		err = f.Decrypt(cipher)
		require.NoError(t, err)

		assert.Equal(t, origFilename, f.Filename)
		assert.Equal(t, origStoragePath, f.StoragePath)
	})

	t.Run("decrypt error on invalid filename", func(t *testing.T) {
		f := &File{
			Filename:    "invalid-hex",
			StoragePath: "/valid/path",
		}

		err := f.Decrypt(cipher)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Ошибка при декодировании поля")
	})

	t.Run("decrypt error on invalid storage path", func(t *testing.T) {
		encFilename, err := cipher.EncryptStr("document.pdf")
		require.NoError(t, err)

		f := &File{
			Filename:    encFilename,
			StoragePath: "invalid-hex",
		}

		err = f.Decrypt(cipher)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Ошибка при декодировании поля")
	})
}

func TestFile_Validation(t *testing.T) {
	v, err := validator.New()
	require.NoError(t, err)

	validFile := func() File {
		return File{
			Id:          "file-1",
			UserId:      "user-1",
			Name:        "Secret Document",
			Filename:    "document.pdf",
			StoragePath: "/attachments/test.pdf",
			Size:        2048,
			Checksum:    "abc123checksum",
		}
	}

	t.Run("valid file", func(t *testing.T) {
		f := validFile()
		err := v.Validate(f)
		assert.NoError(t, err)
	})

	t.Run("empty name", func(t *testing.T) {
		f := validFile()
		f.Name = ""
		err := v.Validate(f)
		require.Error(t, err)
		isErr, errCheck := v.ErrorIs(err, "Название", "required")
		require.NoError(t, errCheck)
		assert.True(t, isErr)
	})

	t.Run("empty filename", func(t *testing.T) {
		f := validFile()
		f.Filename = ""
		err := v.Validate(f)
		require.Error(t, err)
		isErr, errCheck := v.ErrorIs(err, "Название файла", "required")
		require.NoError(t, errCheck)
		assert.True(t, isErr)
	})
}
