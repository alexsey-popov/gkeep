package handler

import (
	"context"
	"errors"
	"testing"

	"github.com/alexsey-popov/gkeep/internal/models"
	api "github.com/alexsey-popov/gkeep/proto"
	pbPassword "github.com/alexsey-popov/gkeep/proto/password"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestHandler_PasswordCreate(t *testing.T) {
	t.Run("unauthenticated - пустой контекст", func(t *testing.T) {
		h, _, _, _, _ := setupTest(t)
		req := pbPassword.CreateRequest_builder{
			Name:     stringPtr("Google"),
			Login:    stringPtr("john@gmail.com"),
			Password: stringPtr("mysecret123"),
		}.Build()

		resp, err := h.PasswordCreate(context.Background(), req)
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.Unauthenticated, st.Code())
		assert.Nil(t, resp.GetPassword())
	})

	t.Run("validation error - слишком короткое имя (< 2)", func(t *testing.T) {
		h, _, _, authCtx, _ := setupTest(t)
		req := pbPassword.CreateRequest_builder{
			Name:     stringPtr("a"),
			Login:    stringPtr("john@gmail.com"),
			Password: stringPtr("mysecret123"),
		}.Build()

		resp, err := h.PasswordCreate(authCtx, req)
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, st.Code())
		assert.Nil(t, resp.GetPassword())
	})

	t.Run("validation error - слишком короткий логин (< 2)", func(t *testing.T) {
		h, _, _, authCtx, _ := setupTest(t)
		req := pbPassword.CreateRequest_builder{
			Name:     stringPtr("Google"),
			Login:    stringPtr("j"),
			Password: stringPtr("mysecret123"),
		}.Build()

		resp, err := h.PasswordCreate(authCtx, req)
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, st.Code())
		assert.Nil(t, resp.GetPassword())
	})

	t.Run("validation error - слишком короткий пароль (< 2)", func(t *testing.T) {
		h, _, _, authCtx, _ := setupTest(t)
		req := pbPassword.CreateRequest_builder{
			Name:     stringPtr("Google"),
			Login:    stringPtr("john@gmail.com"),
			Password: stringPtr("p"),
		}.Build()

		resp, err := h.PasswordCreate(authCtx, req)
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, st.Code())
		assert.Nil(t, resp.GetPassword())
	})

	t.Run("success - успешное создание пароля", func(t *testing.T) {
		h, _, _, authCtx, _ := setupTest(t)
		req := pbPassword.CreateRequest_builder{
			Name:     stringPtr("Google"),
			Login:    stringPtr("john@gmail.com"),
			Password: stringPtr("mysecret123"),
		}.Build()

		resp, err := h.PasswordCreate(authCtx, req)
		require.NoError(t, err)
		require.NotNil(t, resp.GetPassword())
		assert.NotEmpty(t, resp.GetPassword().GetId())
		assert.Equal(t, "Google", resp.GetPassword().GetName())
		assert.Equal(t, "john@gmail.com", resp.GetPassword().GetLogin())
		assert.Equal(t, "mysecret123", resp.GetPassword().GetPassword())
	})

	t.Run("duplicate name - AlreadyExists", func(t *testing.T) {
		h, _, _, authCtx, _ := setupTest(t)
		req := pbPassword.CreateRequest_builder{
			Name:     stringPtr("Google"),
			Login:    stringPtr("john@gmail.com"),
			Password: stringPtr("mysecret123"),
		}.Build()

		_, err := h.PasswordCreate(authCtx, req)
		require.NoError(t, err)

		// Повторный запрос с тем же именем
		resp, err := h.PasswordCreate(authCtx, req)
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.AlreadyExists, st.Code())
		assert.Nil(t, resp.GetPassword())
	})

	t.Run("repository error - Internal", func(t *testing.T) {
		h, rep, _, authCtx, _ := setupTest(t)
		rep.passwordCreateErr = errors.New("db error")

		req := pbPassword.CreateRequest_builder{
			Name:     stringPtr("Google"),
			Login:    stringPtr("john@gmail.com"),
			Password: stringPtr("mysecret123"),
		}.Build()

		resp, err := h.PasswordCreate(authCtx, req)
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.Internal, st.Code())
		assert.Nil(t, resp.GetPassword())
	})
}

func TestHandler_PasswordList(t *testing.T) {
	t.Run("unauthenticated - пустой контекст", func(t *testing.T) {
		h, _, _, _, _ := setupTest(t)
		req := &api.ListRequest{}

		resp, err := h.PasswordList(context.Background(), req)
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.Unauthenticated, st.Code())
		assert.Nil(t, resp.GetPassword())
	})

	t.Run("success - пустой список", func(t *testing.T) {
		h, _, _, authCtx, _ := setupTest(t)
		req := &api.ListRequest{}

		resp, err := h.PasswordList(authCtx, req)
		require.NoError(t, err)
		assert.Empty(t, resp.GetPassword())
	})

	t.Run("success - список записей с фильтрацией", func(t *testing.T) {
		h, rep, _, authCtx, userID := setupTest(t)

		_ = rep.PasswordCreate(authCtx, &models.Password{
			UserId:   userID,
			Name:     "Google Account",
			Login:    "user@google.com",
			Password: "pass1",
		})
		_ = rep.PasswordCreate(authCtx, &models.Password{
			UserId:   userID,
			Name:     "GitHub Account",
			Login:    "user_git",
			Password: "pass2",
		})

		// Без фильтра
		listReq := &api.ListRequest{}
		resp, err := h.PasswordList(authCtx, listReq)
		require.NoError(t, err)
		assert.Len(t, resp.GetPassword(), 2)

		// С фильтром по "goog"
		filterName := "goog"
		listReqFiltered := api.ListRequest_builder{
			Name: &filterName,
		}.Build()
		respFiltered, err := h.PasswordList(authCtx, listReqFiltered)
		require.NoError(t, err)
		require.Len(t, respFiltered.GetPassword(), 1)
		assert.Equal(t, "Google Account", respFiltered.GetPassword()[0].GetName())
	})

	t.Run("repository error - Internal", func(t *testing.T) {
		h, rep, _, authCtx, _ := setupTest(t)
		rep.passwordListErr = errors.New("db error")

		req := &api.ListRequest{}
		resp, err := h.PasswordList(authCtx, req)
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.Internal, st.Code())
		assert.Nil(t, resp.GetPassword())
	})
}

func TestHandler_PasswordUpdate(t *testing.T) {
	t.Run("unauthenticated - пустой контекст", func(t *testing.T) {
		h, _, _, _, _ := setupTest(t)
		req := pbPassword.UpdateRequest_builder{
			Id:       stringPtr("pwd-1"),
			Name:     stringPtr("Google"),
			Login:    stringPtr("john@gmail.com"),
			Password: stringPtr("newpass123"),
		}.Build()

		resp, err := h.PasswordUpdate(context.Background(), req)
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.Unauthenticated, st.Code())
		assert.Nil(t, resp.GetPassword())
	})

	t.Run("validation error - невалидные данные", func(t *testing.T) {
		h, _, _, authCtx, _ := setupTest(t)
		req := pbPassword.UpdateRequest_builder{
			Id:       stringPtr("pwd-1"),
			Name:     stringPtr(""),
			Login:    stringPtr("john@gmail.com"),
			Password: stringPtr("newpass123"),
		}.Build()

		resp, err := h.PasswordUpdate(authCtx, req)
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, st.Code())
		assert.Nil(t, resp.GetPassword())
	})

	t.Run("not found - запись не существует", func(t *testing.T) {
		h, _, _, authCtx, _ := setupTest(t)
		req := pbPassword.UpdateRequest_builder{
			Id:       stringPtr("nonexistent-id"),
			Name:     stringPtr("Google"),
			Login:    stringPtr("john@gmail.com"),
			Password: stringPtr("newpass123"),
		}.Build()

		resp, err := h.PasswordUpdate(authCtx, req)
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.NotFound, st.Code())
		assert.Nil(t, resp.GetPassword())
	})

	t.Run("repository error - Internal", func(t *testing.T) {
		h, rep, _, authCtx, _ := setupTest(t)
		rep.passwordUpdateErr = errors.New("db error")

		req := pbPassword.UpdateRequest_builder{
			Id:       stringPtr("pwd-1"),
			Name:     stringPtr("Google"),
			Login:    stringPtr("john@gmail.com"),
			Password: stringPtr("newpass123"),
		}.Build()

		resp, err := h.PasswordUpdate(authCtx, req)
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.Internal, st.Code())
		assert.Nil(t, resp.GetPassword())
	})

	t.Run("success - успешное обновление пароля", func(t *testing.T) {
		h, rep, _, authCtx, userID := setupTest(t)
		pwdItem := &models.Password{
			UserId:   userID,
			Name:     "Старый пароль",
			Login:    "old_login",
			Password: "old_password",
		}
		err := rep.PasswordCreate(authCtx, pwdItem)
		require.NoError(t, err)

		req := pbPassword.UpdateRequest_builder{
			Id:       &pwdItem.Id,
			Name:     stringPtr("Новый Google"),
			Login:    stringPtr("new_login"),
			Password: stringPtr("new_password"),
		}.Build()

		resp, err := h.PasswordUpdate(authCtx, req)
		require.NoError(t, err)
		require.NotNil(t, resp.GetPassword())
		assert.Equal(t, pwdItem.Id, resp.GetPassword().GetId())
		assert.Equal(t, "Новый Google", resp.GetPassword().GetName())
		assert.Equal(t, "new_login", resp.GetPassword().GetLogin())
		assert.Equal(t, "new_password", resp.GetPassword().GetPassword())
	})
}

func TestHandler_PasswordDelete(t *testing.T) {
	t.Run("unauthenticated - пустой контекст", func(t *testing.T) {
		h, _, _, _, _ := setupTest(t)
		req := api.IDRequest_builder{
			Id: stringPtr("pwd-1"),
		}.Build()

		resp, err := h.PasswordDelete(context.Background(), req)
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.Unauthenticated, st.Code())
		assert.Nil(t, resp.GetPassword())
	})

	t.Run("validation error - пустой id", func(t *testing.T) {
		h, _, _, authCtx, _ := setupTest(t)
		req := api.IDRequest_builder{
			Id: stringPtr("   "),
		}.Build()

		resp, err := h.PasswordDelete(authCtx, req)
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, st.Code())
		assert.Nil(t, resp.GetPassword())
	})

	t.Run("not found - запись не существует", func(t *testing.T) {
		h, _, _, authCtx, _ := setupTest(t)
		req := api.IDRequest_builder{
			Id: stringPtr("nonexistent-id"),
		}.Build()

		resp, err := h.PasswordDelete(authCtx, req)
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.NotFound, st.Code())
		assert.Nil(t, resp.GetPassword())
	})

	t.Run("repository error - Internal", func(t *testing.T) {
		h, rep, _, authCtx, _ := setupTest(t)
		rep.passwordDeleteErr = errors.New("db error")

		req := api.IDRequest_builder{
			Id: stringPtr("pwd-1"),
		}.Build()

		resp, err := h.PasswordDelete(authCtx, req)
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.Internal, st.Code())
		assert.Nil(t, resp.GetPassword())
	})

	t.Run("success - успешное удаление пароля", func(t *testing.T) {
		h, rep, _, authCtx, userID := setupTest(t)
		pwdItem := &models.Password{
			UserId:   userID,
			Name:     "Пароль к удалению",
			Login:    "login_to_del",
			Password: "pass_to_del",
		}
		err := rep.PasswordCreate(authCtx, pwdItem)
		require.NoError(t, err)

		req := api.IDRequest_builder{
			Id: &pwdItem.Id,
		}.Build()

		resp, err := h.PasswordDelete(authCtx, req)
		require.NoError(t, err)
		require.NotNil(t, resp.GetPassword())
		assert.Equal(t, pwdItem.Id, resp.GetPassword().GetId())
		assert.Equal(t, "Пароль к удалению", resp.GetPassword().GetName())

		// Проверяем, что пароля больше нет
		listResp, err := h.PasswordList(authCtx, &api.ListRequest{})
		require.NoError(t, err)
		assert.Empty(t, listResp.GetPassword())
	})
}
