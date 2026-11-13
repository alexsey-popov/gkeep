package models

import (
	"strings"
	"testing"

	"github.com/alexsey-popov/gkeep/internal/validator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUser_Validation(t *testing.T) {
	v, err := validator.New()
	require.NoError(t, err)

	validUser := func() User {
		return User{
			Id:       "user-1",
			Login:    "valid_login",
			Password: "secure_password",
		}
	}

	t.Run("valid user", func(t *testing.T) {
		u := validUser()
		err := v.Validate(u)
		assert.NoError(t, err)
	})

	t.Run("empty login", func(t *testing.T) {
		u := validUser()
		u.Login = ""
		err := v.Validate(u)
		require.Error(t, err)
		isErr, errCheck := v.ErrorIs(err, "Логин", "required")
		require.NoError(t, errCheck)
		assert.True(t, isErr)
	})

	t.Run("login too short", func(t *testing.T) {
		u := validUser()
		u.Login = "abc"
		err := v.Validate(u)
		require.Error(t, err)
		isErr, errCheck := v.ErrorIs(err, "Логин", "min")
		require.NoError(t, errCheck)
		assert.True(t, isErr)
	})

	t.Run("login too long", func(t *testing.T) {
		u := validUser()
		u.Login = strings.Repeat("a", 256)
		err := v.Validate(u)
		require.Error(t, err)
		isErr, errCheck := v.ErrorIs(err, "Логин", "max")
		require.NoError(t, errCheck)
		assert.True(t, isErr)
	})

	t.Run("empty password", func(t *testing.T) {
		u := validUser()
		u.Password = ""
		err := v.Validate(u)
		require.Error(t, err)
		isErr, errCheck := v.ErrorIs(err, "Пароль", "required")
		require.NoError(t, errCheck)
		assert.True(t, isErr)
	})

	t.Run("password too short", func(t *testing.T) {
		u := validUser()
		u.Password = "abc"
		err := v.Validate(u)
		require.Error(t, err)
		isErr, errCheck := v.ErrorIs(err, "Пароль", "min")
		require.NoError(t, errCheck)
		assert.True(t, isErr)
	})

	t.Run("password too long", func(t *testing.T) {
		u := validUser()
		u.Password = strings.Repeat("a", 256)
		err := v.Validate(u)
		require.Error(t, err)
		isErr, errCheck := v.ErrorIs(err, "Пароль", "max")
		require.NoError(t, errCheck)
		assert.True(t, isErr)
	})
}
