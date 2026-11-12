// Обработчик для gRPC запросов
package handler

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/alexsey-popov/gkeep/internal/auth"
	"github.com/alexsey-popov/gkeep/internal/models"
	"github.com/alexsey-popov/gkeep/internal/repository"
	"github.com/alexsey-popov/gkeep/proto"
	"github.com/alexsey-popov/gkeep/proto/card"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// CardList получение списка записей пользователя
func (h *Handler) CardList(ctx context.Context, in *proto.ListRequest) (*card.ListResponse, error) {
	var response card.ListResponse

	userID, err := auth.UserIDFromContext(ctx)
	if err != nil || userID == "" {
		return &response, status.Error(codes.Unauthenticated, "пользователь не аутентифицирован")
	}

	name := in.GetName()
	page := in.GetPage()

	items, err := h.rep.CardList(ctx, userID, name, page)
	if err != nil {
		return &response, status.Error(codes.Internal, "ошибка при получении списка записей")
	}

	pbList := make([]*card.Card, 0, len(items))
	for _, item := range items {
		// Переприсваивать уже не надо, но хуже от этого не будет
		item := item
		pbList = append(pbList, card.Card_builder{
			Id:      &item.Id,
			Name:    &item.Name,
			Number:  &item.Number,
			Expired: &item.Expired,
			Code:    &item.Code,
		}.Build())
	}

	response.SetCard(pbList)

	return &response, nil
}

// CardCreate создание новой записи пароля
func (h *Handler) CardCreate(ctx context.Context, in *card.CreateRequest) (*card.Response, error) {
	var response card.Response

	userID, err := auth.UserIDFromContext(ctx)
	if err != nil || userID == "" {
		return &response, status.Error(codes.Unauthenticated, "пользователь не аутентифицирован")
	}

	item := models.Card{
		UserId:  userID,
		Name:    in.GetName(),
		Number:  in.GetNumber(),
		Expired: in.GetExpired(),
		Code:    in.GetCode(),
	}

	// Валидируем данные
	if err = h.validator.Validate(item); err != nil {
		return &response, status.Error(codes.InvalidArgument, h.validator.TransErrors(err).Error())
	}

	if err = h.rep.CardCreate(ctx, &item); err != nil {
		if errors.Is(err, repository.ErrConflictUnique) {
			return &response, status.Error(codes.AlreadyExists, "запись с таким названием уже существует")
		}
		return &response, status.Error(codes.Internal, "ошибка при создании записи")
	}

	response.SetCard(
		card.Card_builder{
			Id:      &item.Id,
			Name:    &item.Name,
			Number:  &item.Number,
			Expired: &item.Expired,
			Code:    &item.Code,
		}.Build(),
	)

	return &response, nil
}

// CardUpdate обновление пароля по логину
func (h *Handler) CardUpdate(ctx context.Context, in *card.UpdateRequest) (*card.Response, error) {
	var response card.Response

	userID, err := auth.UserIDFromContext(ctx)
	if err != nil || userID == "" {
		return &response, status.Error(codes.Unauthenticated, "пользователь не аутентифицирован")
	}

	item := &models.Card{
		Id:      in.GetId(),
		UserId:  userID,
		Name:    in.GetName(),
		Number:  in.GetNumber(),
		Expired: in.GetExpired(),
		Code:    in.GetCode(),
	}

	// Валидируем данные
	if err = h.validator.Validate(item); err != nil {
		return &response, status.Error(codes.InvalidArgument, h.validator.TransErrors(err).Error())
	}

	// Обновляем данные
	if err = h.rep.CardUpdate(ctx, item); err != nil {
		if errors.Is(err, repository.ErrConflictUnique) {
			return &response, status.Error(codes.AlreadyExists, "запись с таким названием уже существует")
		} else if errors.Is(err, sql.ErrNoRows) {
			return &response, status.Error(codes.NotFound, "запись не найдена")
		}

		return &response, status.Error(codes.Internal, "ошибка при обновлении записи")
	}

	response.SetCard(
		card.Card_builder{
			Id:      &item.Id,
			Name:    &item.Name,
			Number:  &item.Number,
			Expired: &item.Expired,
			Code:    &item.Code,
		}.Build(),
	)

	return &response, nil
}

// CardDelete удаление записи пароля по идентификатору
func (h *Handler) CardDelete(ctx context.Context, in *proto.IDRequest) (*card.Response, error) {
	var response card.Response

	userID, err := auth.UserIDFromContext(ctx)
	if err != nil || userID == "" {
		return &response, status.Error(codes.Unauthenticated, "пользователь не аутентифицирован")
	}

	id := strings.TrimSpace(in.GetId())
	if id == "" {
		return &response, status.Error(codes.InvalidArgument, "id не указан")
	}

	item, err := h.rep.CardDelete(ctx, userID, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &response, status.Error(codes.NotFound, "запись не найдена")
		}
		return &response, status.Error(codes.Internal, "ошибка при удалении записи")
	}

	response.SetCard(
		card.Card_builder{
			Id:      &item.Id,
			Name:    &item.Name,
			Number:  &item.Number,
			Expired: &item.Expired,
			Code:    &item.Code,
		}.Build(),
	)

	return &response, nil
}
