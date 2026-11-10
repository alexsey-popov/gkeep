package handler

import (
	"context"
	"errors"
	"testing"

	"github.com/alexsey-popov/gkeep/internal/models"
	api "github.com/alexsey-popov/gkeep/proto"
	"github.com/alexsey-popov/gkeep/proto/card"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestHandler_CardCreate(t *testing.T) {
	validCardNumber := "4532015112830366" // Visa with valid Luhn checksum

	t.Run("unauthenticated - пустой контекст", func(t *testing.T) {
		h, _, _, _, _ := setupTest(t)
		req := card.CreateRequest_builder{
			Name:    stringPtr("Основная карта"),
			Number:  stringPtr(validCardNumber),
			Expired: stringPtr("1228"),
			Code:    stringPtr("123"),
		}.Build()

		resp, err := h.CardCreate(context.Background(), req)
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.Unauthenticated, st.Code())
		assert.Nil(t, resp.GetCard())
	})

	t.Run("validation error - пустое название", func(t *testing.T) {
		h, _, _, authCtx, _ := setupTest(t)
		req := card.CreateRequest_builder{
			Name:    stringPtr(""),
			Number:  stringPtr(validCardNumber),
			Expired: stringPtr("1228"),
			Code:    stringPtr("123"),
		}.Build()

		resp, err := h.CardCreate(authCtx, req)
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, st.Code())
		assert.Nil(t, resp.GetCard())
	})

	t.Run("validation error - неверный номер карты (luhn checksum)", func(t *testing.T) {
		h, _, _, authCtx, _ := setupTest(t)
		req := card.CreateRequest_builder{
			Name:    stringPtr("Карта"),
			Number:  stringPtr("4532015112830367"),
			Expired: stringPtr("1228"),
			Code:    stringPtr("123"),
		}.Build()

		resp, err := h.CardCreate(authCtx, req)
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, st.Code())
		assert.Nil(t, resp.GetCard())
	})

	t.Run("validation error - неверный срок действия (не 4 цифры)", func(t *testing.T) {
		h, _, _, authCtx, _ := setupTest(t)
		req := card.CreateRequest_builder{
			Name:    stringPtr("Карта"),
			Number:  stringPtr(validCardNumber),
			Expired: stringPtr("12/28"),
			Code:    stringPtr("123"),
		}.Build()

		resp, err := h.CardCreate(authCtx, req)
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, st.Code())
		assert.Nil(t, resp.GetCard())
	})

	t.Run("validation error - неверный код CVV (длина < 3 или > 4)", func(t *testing.T) {
		h, _, _, authCtx, _ := setupTest(t)
		req := card.CreateRequest_builder{
			Name:    stringPtr("Карта"),
			Number:  stringPtr(validCardNumber),
			Expired: stringPtr("1228"),
			Code:    stringPtr("12"),
		}.Build()

		resp, err := h.CardCreate(authCtx, req)
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, st.Code())
		assert.Nil(t, resp.GetCard())
	})

	t.Run("success - успешное создание карты", func(t *testing.T) {
		h, _, _, authCtx, _ := setupTest(t)
		req := card.CreateRequest_builder{
			Name:    stringPtr("Зарплатная карта"),
			Number:  stringPtr(validCardNumber),
			Expired: stringPtr("1228"),
			Code:    stringPtr("123"),
		}.Build()

		resp, err := h.CardCreate(authCtx, req)
		require.NoError(t, err)
		require.NotNil(t, resp.GetCard())
		assert.NotEmpty(t, resp.GetCard().GetId())
		assert.Equal(t, "Зарплатная карта", resp.GetCard().GetName())
		assert.Equal(t, validCardNumber, resp.GetCard().GetNumber())
		assert.Equal(t, "1228", resp.GetCard().GetExpired())
		assert.Equal(t, "123", resp.GetCard().GetCode())
	})

	t.Run("duplicate name - AlreadyExists", func(t *testing.T) {
		h, _, _, authCtx, _ := setupTest(t)
		req := card.CreateRequest_builder{
			Name:    stringPtr("Зарплатная карта"),
			Number:  stringPtr(validCardNumber),
			Expired: stringPtr("1228"),
			Code:    stringPtr("123"),
		}.Build()

		_, err := h.CardCreate(authCtx, req)
		require.NoError(t, err)

		// Повторный запрос с тем же именем
		resp, err := h.CardCreate(authCtx, req)
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.AlreadyExists, st.Code())
		assert.Nil(t, resp.GetCard())
	})

	t.Run("repository error - Internal", func(t *testing.T) {
		h, rep, _, authCtx, _ := setupTest(t)
		rep.cardCreateErr = errors.New("db error")

		req := card.CreateRequest_builder{
			Name:    stringPtr("Карта"),
			Number:  stringPtr(validCardNumber),
			Expired: stringPtr("1228"),
			Code:    stringPtr("123"),
		}.Build()

		resp, err := h.CardCreate(authCtx, req)
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.Internal, st.Code())
		assert.Nil(t, resp.GetCard())
	})
}

func TestHandler_CardList(t *testing.T) {
	validCardNumber := "4532015112830366"

	t.Run("unauthenticated - пустой контекст", func(t *testing.T) {
		h, _, _, _, _ := setupTest(t)
		req := &api.ListRequest{}

		resp, err := h.CardList(context.Background(), req)
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.Unauthenticated, st.Code())
		assert.Nil(t, resp.GetCard())
	})

	t.Run("success - пустой список", func(t *testing.T) {
		h, _, _, authCtx, _ := setupTest(t)
		req := &api.ListRequest{}

		resp, err := h.CardList(authCtx, req)
		require.NoError(t, err)
		assert.Empty(t, resp.GetCard())
	})

	t.Run("success - список записей с фильтрацией", func(t *testing.T) {
		h, rep, _, authCtx, userID := setupTest(t)

		_ = rep.CardCreate(authCtx, &models.Card{
			UserId:  userID,
			Name:    "Сбербанк Дебетовая",
			Number:  validCardNumber,
			Expired: "1228",
			Code:    "123",
		})
		_ = rep.CardCreate(authCtx, &models.Card{
			UserId:  userID,
			Name:    "Тинькофф Кредитная",
			Number:  validCardNumber,
			Expired: "1127",
			Code:    "456",
		})

		// Без фильтра
		listReq := &api.ListRequest{}
		resp, err := h.CardList(authCtx, listReq)
		require.NoError(t, err)
		assert.Len(t, resp.GetCard(), 2)

		// С фильтром по "сбер"
		filterName := "сбер"
		listReqFiltered := api.ListRequest_builder{
			Name: &filterName,
		}.Build()
		respFiltered, err := h.CardList(authCtx, listReqFiltered)
		require.NoError(t, err)
		require.Len(t, respFiltered.GetCard(), 1)
		assert.Equal(t, "Сбербанк Дебетовая", respFiltered.GetCard()[0].GetName())
	})

	t.Run("repository error - Internal", func(t *testing.T) {
		h, rep, _, authCtx, _ := setupTest(t)
		rep.cardListErr = errors.New("db error")

		req := &api.ListRequest{}
		resp, err := h.CardList(authCtx, req)
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.Internal, st.Code())
		assert.Nil(t, resp.GetCard())
	})
}

func TestHandler_CardUpdate(t *testing.T) {
	validCardNumber := "4532015112830366"

	t.Run("unauthenticated - пустой контекст", func(t *testing.T) {
		h, _, _, _, _ := setupTest(t)
		req := card.UpdateRequest_builder{
			Id:      stringPtr("card-1"),
			Name:    stringPtr("Новое имя"),
			Number:  stringPtr(validCardNumber),
			Expired: stringPtr("1228"),
			Code:    stringPtr("123"),
		}.Build()

		resp, err := h.CardUpdate(context.Background(), req)
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.Unauthenticated, st.Code())
		assert.Nil(t, resp.GetCard())
	})

	t.Run("validation error - невалидные данные", func(t *testing.T) {
		h, _, _, authCtx, _ := setupTest(t)
		req := card.UpdateRequest_builder{
			Id:      stringPtr("card-1"),
			Name:    stringPtr(""),
			Number:  stringPtr(validCardNumber),
			Expired: stringPtr("1228"),
			Code:    stringPtr("123"),
		}.Build()

		resp, err := h.CardUpdate(authCtx, req)
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, st.Code())
		assert.Nil(t, resp.GetCard())
	})

	t.Run("not found - запись не существует", func(t *testing.T) {
		h, _, _, authCtx, _ := setupTest(t)
		req := card.UpdateRequest_builder{
			Id:      stringPtr("nonexistent-id"),
			Name:    stringPtr("Карта"),
			Number:  stringPtr(validCardNumber),
			Expired: stringPtr("1228"),
			Code:    stringPtr("123"),
		}.Build()

		resp, err := h.CardUpdate(authCtx, req)
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.NotFound, st.Code())
		assert.Nil(t, resp.GetCard())
	})

	t.Run("repository error - Internal", func(t *testing.T) {
		h, rep, _, authCtx, _ := setupTest(t)
		rep.cardUpdateErr = errors.New("db error")

		req := card.UpdateRequest_builder{
			Id:      stringPtr("card-1"),
			Name:    stringPtr("Карта"),
			Number:  stringPtr(validCardNumber),
			Expired: stringPtr("1228"),
			Code:    stringPtr("123"),
		}.Build()

		resp, err := h.CardUpdate(authCtx, req)
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.Internal, st.Code())
		assert.Nil(t, resp.GetCard())
	})

	t.Run("success - успешное обновление карты", func(t *testing.T) {
		h, rep, _, authCtx, userID := setupTest(t)
		cardItem := &models.Card{
			UserId:  userID,
			Name:    "Старая карта",
			Number:  validCardNumber,
			Expired: "1228",
			Code:    "123",
		}
		err := rep.CardCreate(authCtx, cardItem)
		require.NoError(t, err)

		req := card.UpdateRequest_builder{
			Id:      &cardItem.Id,
			Name:    stringPtr("Обновленная карта"),
			Number:  stringPtr(validCardNumber),
			Expired: stringPtr("0130"),
			Code:    stringPtr("999"),
		}.Build()

		resp, err := h.CardUpdate(authCtx, req)
		require.NoError(t, err)
		require.NotNil(t, resp.GetCard())
		assert.Equal(t, cardItem.Id, resp.GetCard().GetId())
		assert.Equal(t, "Обновленная карта", resp.GetCard().GetName())
		assert.Equal(t, "0130", resp.GetCard().GetExpired())
		assert.Equal(t, "999", resp.GetCard().GetCode())
	})
}

func TestHandler_CardDelete(t *testing.T) {
	validCardNumber := "4532015112830366"

	t.Run("unauthenticated - пустой контекст", func(t *testing.T) {
		h, _, _, _, _ := setupTest(t)
		req := api.IDRequest_builder{
			Id: stringPtr("card-1"),
		}.Build()

		resp, err := h.CardDelete(context.Background(), req)
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.Unauthenticated, st.Code())
		assert.Nil(t, resp.GetCard())
	})

	t.Run("validation error - пустой id", func(t *testing.T) {
		h, _, _, authCtx, _ := setupTest(t)
		req := api.IDRequest_builder{
			Id: stringPtr("   "),
		}.Build()

		resp, err := h.CardDelete(authCtx, req)
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, st.Code())
		assert.Nil(t, resp.GetCard())
	})

	t.Run("not found - запись не существует", func(t *testing.T) {
		h, _, _, authCtx, _ := setupTest(t)
		req := api.IDRequest_builder{
			Id: stringPtr("nonexistent-id"),
		}.Build()

		resp, err := h.CardDelete(authCtx, req)
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.NotFound, st.Code())
		assert.Nil(t, resp.GetCard())
	})

	t.Run("repository error - Internal", func(t *testing.T) {
		h, rep, _, authCtx, _ := setupTest(t)
		rep.cardDeleteErr = errors.New("db error")

		req := api.IDRequest_builder{
			Id: stringPtr("card-1"),
		}.Build()

		resp, err := h.CardDelete(authCtx, req)
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.Internal, st.Code())
		assert.Nil(t, resp.GetCard())
	})

	t.Run("success - успешное удаление карты", func(t *testing.T) {
		h, rep, _, authCtx, userID := setupTest(t)
		cardItem := &models.Card{
			UserId:  userID,
			Name:    "Карта к удалению",
			Number:  validCardNumber,
			Expired: "1228",
			Code:    "123",
		}
		err := rep.CardCreate(authCtx, cardItem)
		require.NoError(t, err)

		req := api.IDRequest_builder{
			Id: &cardItem.Id,
		}.Build()

		resp, err := h.CardDelete(authCtx, req)
		require.NoError(t, err)
		require.NotNil(t, resp.GetCard())
		assert.Equal(t, cardItem.Id, resp.GetCard().GetId())
		assert.Equal(t, "Карта к удалению", resp.GetCard().GetName())

		// Проверяем, что карты больше нет
		listResp, err := h.CardList(authCtx, &api.ListRequest{})
		require.NoError(t, err)
		assert.Empty(t, listResp.GetCard())
	})
}

func stringPtr(s string) *string {
	return &s
}
