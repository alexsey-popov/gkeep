package handler

import (
	"context"
	"errors"
	"testing"

	"github.com/alexsey-popov/gkeep/internal/models"
	api "github.com/alexsey-popov/gkeep/proto"
	"github.com/alexsey-popov/gkeep/proto/text"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestHandler_TextCreate(t *testing.T) {
	t.Run("unauthenticated - пустой контекст", func(t *testing.T) {
		h, _, _, _, _ := setupTest(t)
		req := text.CreateRequest_builder{
			Name: stringPtr("Заметка"),
			Text: stringPtr("Секретный текст"),
		}.Build()

		resp, err := h.TextCreate(context.Background(), req)
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.Unauthenticated, st.Code())
		assert.Nil(t, resp.GetText())
	})

	t.Run("validation error - пустое название", func(t *testing.T) {
		h, _, _, authCtx, _ := setupTest(t)
		req := text.CreateRequest_builder{
			Name: stringPtr(""),
			Text: stringPtr("Текст"),
		}.Build()

		resp, err := h.TextCreate(authCtx, req)
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, st.Code())
		assert.Nil(t, resp.GetText())
	})

	t.Run("validation error - пустой текст", func(t *testing.T) {
		h, _, _, authCtx, _ := setupTest(t)
		req := text.CreateRequest_builder{
			Name: stringPtr("Название"),
			Text: stringPtr(""),
		}.Build()

		resp, err := h.TextCreate(authCtx, req)
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, st.Code())
		assert.Nil(t, resp.GetText())
	})

	t.Run("success - успешное создание текстовой записи", func(t *testing.T) {
		h, _, _, authCtx, _ := setupTest(t)
		req := text.CreateRequest_builder{
			Name: stringPtr("Пин-коды"),
			Text: stringPtr("1234, 5678"),
		}.Build()

		resp, err := h.TextCreate(authCtx, req)
		require.NoError(t, err)
		require.NotNil(t, resp.GetText())
		assert.NotEmpty(t, resp.GetText().GetId())
		assert.Equal(t, "Пин-коды", resp.GetText().GetName())
		assert.Equal(t, "1234, 5678", resp.GetText().GetText())
	})

	t.Run("duplicate name - AlreadyExists", func(t *testing.T) {
		h, _, _, authCtx, _ := setupTest(t)
		req := text.CreateRequest_builder{
			Name: stringPtr("Пин-коды"),
			Text: stringPtr("1234, 5678"),
		}.Build()

		_, err := h.TextCreate(authCtx, req)
		require.NoError(t, err)

		// Повторный запрос с тем же именем
		resp, err := h.TextCreate(authCtx, req)
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.AlreadyExists, st.Code())
		assert.Nil(t, resp.GetText())
	})

	t.Run("repository error - Internal", func(t *testing.T) {
		h, rep, _, authCtx, _ := setupTest(t)
		rep.textCreateErr = errors.New("db error")

		req := text.CreateRequest_builder{
			Name: stringPtr("Заметка"),
			Text: stringPtr("Текст"),
		}.Build()

		resp, err := h.TextCreate(authCtx, req)
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.Internal, st.Code())
		assert.Nil(t, resp.GetText())
	})
}

func TestHandler_TextList(t *testing.T) {
	t.Run("unauthenticated - пустой контекст", func(t *testing.T) {
		h, _, _, _, _ := setupTest(t)
		req := &api.ListRequest{}

		resp, err := h.TextList(context.Background(), req)
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.Unauthenticated, st.Code())
		assert.Nil(t, resp.GetText())
	})

	t.Run("success - пустой список", func(t *testing.T) {
		h, _, _, authCtx, _ := setupTest(t)
		req := &api.ListRequest{}

		resp, err := h.TextList(authCtx, req)
		require.NoError(t, err)
		assert.Empty(t, resp.GetText())
	})

	t.Run("success - список записей с фильтрацией", func(t *testing.T) {
		h, rep, _, authCtx, userID := setupTest(t)

		_ = rep.TextCreate(authCtx, &models.Text{
			UserId: userID,
			Name:   "Заметка о работе",
			Text:   "Список задач на неделю",
		})
		_ = rep.TextCreate(authCtx, &models.Text{
			UserId: userID,
			Name:   "Личные заметки",
			Text:   "Купить молоко",
		})

		// Без фильтра
		listReq := &api.ListRequest{}
		resp, err := h.TextList(authCtx, listReq)
		require.NoError(t, err)
		assert.Len(t, resp.GetText(), 2)

		// С фильтром по "работе"
		filterName := "работе"
		listReqFiltered := api.ListRequest_builder{
			Name: &filterName,
		}.Build()
		respFiltered, err := h.TextList(authCtx, listReqFiltered)
		require.NoError(t, err)
		require.Len(t, respFiltered.GetText(), 1)
		assert.Equal(t, "Заметка о работе", respFiltered.GetText()[0].GetName())
	})

	t.Run("repository error - Internal", func(t *testing.T) {
		h, rep, _, authCtx, _ := setupTest(t)
		rep.textListErr = errors.New("db error")

		req := &api.ListRequest{}
		resp, err := h.TextList(authCtx, req)
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.Internal, st.Code())
		assert.Nil(t, resp.GetText())
	})
}

func TestHandler_TextUpdate(t *testing.T) {
	t.Run("unauthenticated - пустой контекст", func(t *testing.T) {
		h, _, _, _, _ := setupTest(t)
		req := text.UpdateRequest_builder{
			Id:   stringPtr("text-1"),
			Name: stringPtr("Новое имя"),
			Text: stringPtr("Новый текст"),
		}.Build()

		resp, err := h.TextUpdate(context.Background(), req)
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.Unauthenticated, st.Code())
		assert.Nil(t, resp.GetText())
	})

	t.Run("validation error - невалидные данные", func(t *testing.T) {
		h, _, _, authCtx, _ := setupTest(t)
		req := text.UpdateRequest_builder{
			Id:   stringPtr("text-1"),
			Name: stringPtr(""),
			Text: stringPtr("Новый текст"),
		}.Build()

		resp, err := h.TextUpdate(authCtx, req)
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, st.Code())
		assert.Nil(t, resp.GetText())
	})

	t.Run("not found - запись не существует", func(t *testing.T) {
		h, _, _, authCtx, _ := setupTest(t)
		req := text.UpdateRequest_builder{
			Id:   stringPtr("nonexistent-id"),
			Name: stringPtr("Заметка"),
			Text: stringPtr("Текст"),
		}.Build()

		resp, err := h.TextUpdate(authCtx, req)
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.NotFound, st.Code())
		assert.Nil(t, resp.GetText())
	})

	t.Run("repository error - Internal", func(t *testing.T) {
		h, rep, _, authCtx, _ := setupTest(t)
		rep.textUpdateErr = errors.New("db error")

		req := text.UpdateRequest_builder{
			Id:   stringPtr("text-1"),
			Name: stringPtr("Заметка"),
			Text: stringPtr("Текст"),
		}.Build()

		resp, err := h.TextUpdate(authCtx, req)
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.Internal, st.Code())
		assert.Nil(t, resp.GetText())
	})

	t.Run("success - успешное обновление текста", func(t *testing.T) {
		h, rep, _, authCtx, userID := setupTest(t)
		textItem := &models.Text{
			UserId: userID,
			Name:   "Старая заметка",
			Text:   "Старый текст",
		}
		err := rep.TextCreate(authCtx, textItem)
		require.NoError(t, err)

		req := text.UpdateRequest_builder{
			Id:   &textItem.Id,
			Name: stringPtr("Обновленная заметка"),
			Text: stringPtr("Обновленный текст"),
		}.Build()

		resp, err := h.TextUpdate(authCtx, req)
		require.NoError(t, err)
		require.NotNil(t, resp.GetText())
		assert.Equal(t, textItem.Id, resp.GetText().GetId())
		assert.Equal(t, "Обновленная заметка", resp.GetText().GetName())
		assert.Equal(t, "Обновленный текст", resp.GetText().GetText())
	})
}

func TestHandler_TextDelete(t *testing.T) {
	t.Run("unauthenticated - пустой контекст", func(t *testing.T) {
		h, _, _, _, _ := setupTest(t)
		req := api.IDRequest_builder{
			Id: stringPtr("text-1"),
		}.Build()

		resp, err := h.TextDelete(context.Background(), req)
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.Unauthenticated, st.Code())
		assert.Nil(t, resp.GetText())
	})

	t.Run("validation error - пустой id", func(t *testing.T) {
		h, _, _, authCtx, _ := setupTest(t)
		req := api.IDRequest_builder{
			Id: stringPtr("   "),
		}.Build()

		resp, err := h.TextDelete(authCtx, req)
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, st.Code())
		assert.Nil(t, resp.GetText())
	})

	t.Run("not found - запись не существует", func(t *testing.T) {
		h, _, _, authCtx, _ := setupTest(t)
		req := api.IDRequest_builder{
			Id: stringPtr("nonexistent-id"),
		}.Build()

		resp, err := h.TextDelete(authCtx, req)
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.NotFound, st.Code())
		assert.Nil(t, resp.GetText())
	})

	t.Run("repository error - Internal", func(t *testing.T) {
		h, rep, _, authCtx, _ := setupTest(t)
		rep.textDeleteErr = errors.New("db error")

		req := api.IDRequest_builder{
			Id: stringPtr("text-1"),
		}.Build()

		resp, err := h.TextDelete(authCtx, req)
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.Internal, st.Code())
		assert.Nil(t, resp.GetText())
	})

	t.Run("success - успешное удаление текста", func(t *testing.T) {
		h, rep, _, authCtx, userID := setupTest(t)
		textItem := &models.Text{
			UserId: userID,
			Name:   "Заметка к удалению",
			Text:   "Текст к удалению",
		}
		err := rep.TextCreate(authCtx, textItem)
		require.NoError(t, err)

		req := api.IDRequest_builder{
			Id: &textItem.Id,
		}.Build()

		resp, err := h.TextDelete(authCtx, req)
		require.NoError(t, err)
		require.NotNil(t, resp.GetText())
		assert.Equal(t, textItem.Id, resp.GetText().GetId())
		assert.Equal(t, "Заметка к удалению", resp.GetText().GetName())

		// Проверяем, что заметки больше нет
		listResp, err := h.TextList(authCtx, &api.ListRequest{})
		require.NoError(t, err)
		assert.Empty(t, listResp.GetText())
	})
}
