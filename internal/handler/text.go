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
	"github.com/alexsey-popov/gkeep/proto/text"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// TextList получение списка записей пользователя
func (h *Handler) TextList(ctx context.Context, in *proto.ListRequest) (*text.ListResponse, error) {
	var response text.ListResponse

	userID, err := auth.UserIDFromContext(ctx)
	if err != nil || userID == "" {
		return &response, status.Error(codes.Unauthenticated, "пользователь не аутентифицирован")
	}

	name := in.GetName()
	page := in.GetPage()

	items, err := h.rep.TextList(ctx, userID, name, page)
	if err != nil {
		return &response, status.Error(codes.Internal, "ошибка при получении списка записей")
	}

	pbList := make([]*text.Text, 0, len(items))
	for _, item := range items {
		// Переприсваивать уже не надо, но хуже от этого не будет
		item := item
		pbList = append(pbList, text.Text_builder{
			Id:   &item.Id,
			Name: &item.Name,
			Text: &item.Text,
		}.Build())
	}

	response.SetText(pbList)

	return &response, nil
}

// TextCreate создание новой записи пароля
func (h *Handler) TextCreate(ctx context.Context, in *text.CreateRequest) (*text.Response, error) {
	var response text.Response

	userID, err := auth.UserIDFromContext(ctx)
	if err != nil || userID == "" {
		return &response, status.Error(codes.Unauthenticated, "пользователь не аутентифицирован")
	}

	item := models.Text{
		UserId: userID,
		Name:   in.GetName(),
		Text:   in.GetText(),
	}

	// Валидируем данные
	if err = h.validator.Validate(item); err != nil {
		return &response, status.Error(codes.InvalidArgument, h.validator.TransErrors(err).Error())
	}

	if err = h.rep.TextCreate(ctx, &item); err != nil {
		if errors.Is(err, repository.ErrConflictUnique) {
			return &response, status.Error(codes.AlreadyExists, "запись с таким названием уже существует")
		}

		return &response, status.Error(codes.Internal, "ошибка при создании записи")
	}

	response.SetText(
		text.Text_builder{
			Id:   &item.Id,
			Name: &item.Name,
			Text: &item.Text,
		}.Build(),
	)

	return &response, nil
}

// TextUpdate обновление пароля по логину
func (h *Handler) TextUpdate(ctx context.Context, in *text.UpdateRequest) (*text.Response, error) {
	var response text.Response

	userID, err := auth.UserIDFromContext(ctx)
	if err != nil || userID == "" {
		return &response, status.Error(codes.Unauthenticated, "пользователь не аутентифицирован")
	}

	item := models.Text{
		Id:     in.GetId(),
		UserId: userID,
		Name:   in.GetName(),
		Text:   in.GetText(),
	}

	// Валидируем данные
	if err = h.validator.Validate(item); err != nil {
		return &response, status.Error(codes.InvalidArgument, h.validator.TransErrors(err).Error())
	}

	if err = h.rep.TextUpdate(ctx, &item); err != nil {
		if errors.Is(err, repository.ErrConflictUnique) {
			return &response, status.Error(codes.AlreadyExists, "запись с таким названием уже существует")
		} else if errors.Is(err, sql.ErrNoRows) {
			return &response, status.Error(codes.NotFound, "запись не найдена")
		}
		return &response, status.Error(codes.Internal, "ошибка при обновлении записи")
	}

	response.SetText(
		text.Text_builder{
			Id:   &item.Id,
			Name: &item.Name,
			Text: &item.Text,
		}.Build(),
	)

	return &response, nil
}

// TextDelete удаление записи пароля по идентификатору
func (h *Handler) TextDelete(ctx context.Context, in *proto.IDRequest) (*text.Response, error) {
	var response text.Response

	userID, err := auth.UserIDFromContext(ctx)
	if err != nil || userID == "" {
		return &response, status.Error(codes.Unauthenticated, "пользователь не аутентифицирован")
	}

	id := strings.TrimSpace(in.GetId())
	if id == "" {
		return &response, status.Error(codes.InvalidArgument, "id не указан")
	}

	item, err := h.rep.TextDelete(ctx, userID, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &response, status.Error(codes.NotFound, "запись не найдена")
		}
		return &response, status.Error(codes.Internal, "ошибка при удалении записи")
	}

	response.SetText(
		text.Text_builder{
			Id:   &item.Id,
			Name: &item.Name,
			Text: &item.Text,
		}.Build(),
	)

	return &response, nil
}
