package models

import (
	"testing"
	"time"

	"github.com/alexsey-popov/gkeep/internal/crypto"
	"github.com/alexsey-popov/gkeep/internal/validator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPassword_Encrypt_Decrypt(t *testing.T) {
	cipher, err := crypto.NewWithTempKey()
	require.NoError(t, err)

	t.Run("success encrypt and decrypt roundtrip", func(t *testing.T) {
		pass := &Password{
			Id:        "pass-1",
			UserId:    "9b2e8f96-3c48-4394-bb9e-3d8495033c46",
			Name:      "Google Account",
			Login:     "user@gmail.com",
			Password:  "superSecret123",
			CreatedAt: time.Now(),
		}

		origLogin := pass.Login
		origPassword := pass.Password

		err := pass.Encrypt(cipher)
		require.NoError(t, err)

		assert.NotEqual(t, origLogin, pass.Login)
		assert.NotEqual(t, origPassword, pass.Password)

		err = pass.Decrypt(cipher)
		require.NoError(t, err)

		assert.Equal(t, origLogin, pass.Login)
		assert.Equal(t, origPassword, pass.Password)
	})

	t.Run("decrypt error on invalid login", func(t *testing.T) {
		pass := &Password{
			Login:    "invalid-hex",
			Password: "valid",
		}

		err := pass.Decrypt(cipher)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "ошибка при декодировании поля login")
	})

	t.Run("decrypt error on invalid password", func(t *testing.T) {
		encLogin, err := cipher.EncryptStr("user@gmail.com")
		require.NoError(t, err)

		pass := &Password{
			Login:    encLogin,
			Password: "invalid-hex",
		}

		err = pass.Decrypt(cipher)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "ошибка при декодировании поля password")
	})
}

func TestPassword_Validation(t *testing.T) {
	v, err := validator.New()
	require.NoError(t, err)

	validPassword := func() Password {
		return Password{
			Id:       "pass-1",
			UserId:   "9b2e8f96-3c48-4394-bb9e-3d8495033c46",
			Name:     "Google Account",
			Login:    "my_login",
			Password: "my_password",
		}
	}

	t.Run("valid password", func(t *testing.T) {
		pass := validPassword()
		err := v.Validate(pass)
		assert.NoError(t, err)
	})

	t.Run("invalid user_id - not uuid", func(t *testing.T) {
		pass := validPassword()
		pass.UserId = "not-a-uuid"
		err := v.Validate(pass)
		require.Error(t, err)
		isErr, errCheck := v.ErrorIs(err, "ID пользователя", "uuid")
		require.NoError(t, errCheck)
		assert.True(t, isErr)
	})

	t.Run("empty user_id", func(t *testing.T) {
		pass := validPassword()
		pass.UserId = ""
		err := v.Validate(pass)
		require.Error(t, err)
		isErr, errCheck := v.ErrorIs(err, "ID пользователя", "required")
		require.NoError(t, errCheck)
		assert.True(t, isErr)
	})

	t.Run("empty name", func(t *testing.T) {
		pass := validPassword()
		pass.Name = ""
		err := v.Validate(pass)
		require.Error(t, err)
		isErr, errCheck := v.ErrorIs(err, "Название", "required")
		require.NoError(t, errCheck)
		assert.True(t, isErr)
	})

	t.Run("name too short", func(t *testing.T) {
		pass := validPassword()
		pass.Name = "a"
		err := v.Validate(pass)
		require.Error(t, err)
		isErr, errCheck := v.ErrorIs(err, "Название", "min")
		require.NoError(t, errCheck)
		assert.True(t, isErr)
	})

	t.Run("login too short", func(t *testing.T) {
		pass := validPassword()
		pass.Login = "a"
		err := v.Validate(pass)
		require.Error(t, err)
		isErr, errCheck := v.ErrorIs(err, "Логин", "min")
		require.NoError(t, errCheck)
		assert.True(t, isErr)
	})

	t.Run("empty login", func(t *testing.T) {
		pass := validPassword()
		pass.Login = ""
		err := v.Validate(pass)
		require.Error(t, err)
		isErr, errCheck := v.ErrorIs(err, "Логин", "required")
		require.NoError(t, errCheck)
		assert.True(t, isErr)
	})

	t.Run("password too short", func(t *testing.T) {
		pass := validPassword()
		pass.Password = "a"
		err := v.Validate(pass)
		require.Error(t, err)
		isErr, errCheck := v.ErrorIs(err, "Пароль", "min")
		require.NoError(t, errCheck)
		assert.True(t, isErr)
	})

	t.Run("empty password", func(t *testing.T) {
		pass := validPassword()
		pass.Password = ""
		err := v.Validate(pass)
		require.Error(t, err)
		isErr, errCheck := v.ErrorIs(err, "Пароль", "required")
		require.NoError(t, errCheck)
		assert.True(t, isErr)
	})
}
