package models

import (
	"strings"
	"testing"
	"time"

	"github.com/alexsey-popov/gkeep/internal/crypto"
	"github.com/alexsey-popov/gkeep/internal/validator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestText_Encrypt_Decrypt(t *testing.T) {
	cipher, err := crypto.NewWithTempKey()
	require.NoError(t, err)

	t.Run("success encrypt and decrypt roundtrip", func(t *testing.T) {
		item := &Text{
			Id:        "text-1",
			UserId:    "user-1",
			Name:      "Secret Note",
			Text:      "This is a confidential note with sensitive details.",
			CreatedAt: time.Now(),
		}

		origText := item.Text

		err := item.Encrypt(cipher)
		require.NoError(t, err)

		assert.NotEqual(t, origText, item.Text)

		err = item.Decrypt(cipher)
		require.NoError(t, err)

		assert.Equal(t, origText, item.Text)
	})

	t.Run("decrypt error on invalid text", func(t *testing.T) {
		item := &Text{
			Text: "invalid-hex",
		}

		err := item.Decrypt(cipher)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Ошибка при декодировании поля")
	})
}

func TestText_Validation(t *testing.T) {
	v, err := validator.New()
	require.NoError(t, err)

	validText := func() Text {
		return Text{
			Id:     "text-1",
			UserId: "user-1",
			Name:   "Secret Note",
			Text:   "This is a confidential note.",
		}
	}

	t.Run("valid text", func(t *testing.T) {
		item := validText()
		err := v.Validate(item)
		assert.NoError(t, err)
	})

	t.Run("empty name", func(t *testing.T) {
		item := validText()
		item.Name = ""
		err := v.Validate(item)
		require.Error(t, err)
		isErr, errCheck := v.ErrorIs(err, "Название", "required")
		require.NoError(t, errCheck)
		assert.True(t, isErr)
	})

	t.Run("empty text", func(t *testing.T) {
		item := validText()
		item.Text = ""
		err := v.Validate(item)
		require.Error(t, err)
		isErr, errCheck := v.ErrorIs(err, "Текст", "required")
		require.NoError(t, errCheck)
		assert.True(t, isErr)
	})

	t.Run("text exceeds maximum length", func(t *testing.T) {
		item := validText()
		item.Text = strings.Repeat("a", 10001)
		err := v.Validate(item)
		require.Error(t, err)
		isErr, errCheck := v.ErrorIs(err, "Текст", "max")
		require.NoError(t, errCheck)
		assert.True(t, isErr)
	})
}
