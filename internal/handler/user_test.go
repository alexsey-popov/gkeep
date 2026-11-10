package handler

import (
	"context"
	"errors"
	"testing"

	pbUser "github.com/alexsey-popov/gkeep/proto/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestHandler_RegisterUser(t *testing.T) {
	t.Run("validation error - слишком короткий логин (< 4)", func(t *testing.T) {
		h, _, _, _, _ := setupTest(t)
		req := pbUser.Request_builder{
			Login:    stringPtr("usr"),
			Password: stringPtr("strong_password"),
		}.Build()

		resp, err := h.RegisterUser(context.Background(), req)
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, st.Code())
		assert.Nil(t, resp.GetUser())
	})

	t.Run("validation error - слишком короткий пароль (< 4)", func(t *testing.T) {
		h, _, _, _, _ := setupTest(t)
		req := pbUser.Request_builder{
			Login:    stringPtr("john_doe"),
			Password: stringPtr("123"),
		}.Build()

		resp, err := h.RegisterUser(context.Background(), req)
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, st.Code())
		assert.Nil(t, resp.GetUser())
	})

	t.Run("success - успешная регистрация", func(t *testing.T) {
		h, _, _, _, _ := setupTest(t)
		req := pbUser.Request_builder{
			Login:    stringPtr("john_doe"),
			Password: stringPtr("strong_password"),
		}.Build()

		resp, err := h.RegisterUser(context.Background(), req)
		require.NoError(t, err)
		require.NotNil(t, resp.GetUser())
		assert.NotEmpty(t, resp.GetUser().GetId())
		assert.Equal(t, "john_doe", resp.GetUser().GetLogin())
	})

	t.Run("duplicate login - AlreadyExists", func(t *testing.T) {
		h, _, _, _, _ := setupTest(t)
		req := pbUser.Request_builder{
			Login:    stringPtr("john_doe"),
			Password: stringPtr("strong_password"),
		}.Build()

		_, err := h.RegisterUser(context.Background(), req)
		require.NoError(t, err)

		// Повторная регистрация под тем же логином
		resp, err := h.RegisterUser(context.Background(), req)
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.AlreadyExists, st.Code())
		assert.Nil(t, resp.GetUser())
	})

	t.Run("repository error - Internal", func(t *testing.T) {
		h, rep, _, _, _ := setupTest(t)
		rep.userCreateErr = errors.New("db error")

		req := pbUser.Request_builder{
			Login:    stringPtr("john_doe"),
			Password: stringPtr("strong_password"),
		}.Build()

		resp, err := h.RegisterUser(context.Background(), req)
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.Internal, st.Code())
		assert.Nil(t, resp.GetUser())
	})
}

func TestHandler_LoginUser(t *testing.T) {
	t.Run("not found - пользователь не существует", func(t *testing.T) {
		h, _, _, _, _ := setupTest(t)
		req := pbUser.Request_builder{
			Login:    stringPtr("unknown_user"),
			Password: stringPtr("some_password"),
		}.Build()

		resp, err := h.LoginUser(context.Background(), req)
		require.Error(t, err)
		assert.Nil(t, resp.GetUser())
	})

	t.Run("wrong password - Unauthenticated", func(t *testing.T) {
		h, _, _, _, _ := setupTest(t)
		regReq := pbUser.Request_builder{
			Login:    stringPtr("john_doe"),
			Password: stringPtr("correct_password"),
		}.Build()

		_, err := h.RegisterUser(context.Background(), regReq)
		require.NoError(t, err)

		loginReq := pbUser.Request_builder{
			Login:    stringPtr("john_doe"),
			Password: stringPtr("wrong_password"),
		}.Build()

		resp, err := h.LoginUser(context.Background(), loginReq)
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.Unauthenticated, st.Code())
		assert.Nil(t, resp.GetUser())
	})

	t.Run("success - успешный вход", func(t *testing.T) {
		h, _, _, _, _ := setupTest(t)
		regReq := pbUser.Request_builder{
			Login:    stringPtr("john_doe"),
			Password: stringPtr("correct_password"),
		}.Build()

		_, err := h.RegisterUser(context.Background(), regReq)
		require.NoError(t, err)

		loginReq := pbUser.Request_builder{
			Login:    stringPtr("john_doe"),
			Password: stringPtr("correct_password"),
		}.Build()

		resp, err := h.LoginUser(context.Background(), loginReq)
		require.NoError(t, err)
		require.NotNil(t, resp.GetUser())
		assert.Equal(t, "john_doe", resp.GetUser().GetLogin())
		assert.NotEmpty(t, resp.GetUser().GetId())
	})
}
