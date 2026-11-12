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
	"github.com/alexsey-popov/gkeep/proto/password"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// PasswordList получение списка записей пользователя
func (h *Handler) PasswordList(ctx context.Context, in *proto.ListRequest) (*password.ListResponse, error) {
	var response password.ListResponse

	userID, err := auth.UserIDFromContext(ctx)
	if err != nil || userID == "" {
		return &response, status.Error(codes.Unauthenticated, "пользователь не аутентифицирован")
	}

	name := in.GetName()
	page := in.GetPage()

	items, err := h.rep.PasswordList(ctx, userID, name, page)
	if err != nil {
		return &response, status.Error(codes.Internal, "ошибка при получении списка записей")
	}

	pbList := make([]*password.Password, 0, len(items))
	for _, item := range items {
		// Переприсваивать уже не надо, но хуже от этого не будет
		item := item
		pbList = append(pbList, password.Password_builder{
			Id:       &item.Id,
			Name:     &item.Name,
			Login:    &item.Login,
			Password: &item.Password,
		}.Build())
	}

	response.SetPassword(pbList)

	return &response, nil
}

// PasswordCreate создание новой записи пароля
func (h *Handler) PasswordCreate(ctx context.Context, in *password.CreateRequest) (*password.Response, error) {
	var response password.Response

	// Получаем ID пользователя
	userID, err := auth.UserIDFromContext(ctx)
	if err != nil || userID == "" {
		return &response, status.Error(codes.Unauthenticated, "пользователь не аутентифицирован")
	}

	pass := models.Password{
		UserId:   userID,
		Name:     strings.TrimSpace(in.GetName()),
		Login:    strings.TrimSpace(in.GetLogin()),
		Password: in.GetPassword(),
	}

	// Валидируем данные
	if err = h.validator.Validate(pass); err != nil {
		return &response, status.Error(codes.InvalidArgument, h.validator.TransErrors(err).Error())
	}

	// Отражаем данные в базе
	if err = h.rep.PasswordCreate(ctx, &pass); err != nil {
		if errors.Is(err, repository.ErrConflictUnique) {
			return &response, status.Error(codes.AlreadyExists, "запись с таким названием уже существует")
		}

		return &response, status.Error(codes.Internal, "ошибка при создании записи")
	}

	response.SetPassword(
		password.Password_builder{
			Id:       &pass.Id,
			Name:     &pass.Name,
			Login:    &pass.Login,
			Password: &pass.Password,
		}.Build(),
	)

	return &response, nil
}

// PasswordUpdate обновление пароля по логину
func (h *Handler) PasswordUpdate(ctx context.Context, in *password.UpdateRequest) (*password.Response, error) {
	var response password.Response

	userID, err := auth.UserIDFromContext(ctx)
	if err != nil || userID == "" {
		return &response, status.Error(codes.Unauthenticated, "пользователь не аутентифицирован")
	}

	pass := models.Password{
		Id:       in.GetId(),
		UserId:   userID,
		Name:     strings.TrimSpace(in.GetName()),
		Login:    strings.TrimSpace(in.GetLogin()),
		Password: in.GetPassword(),
	}

	// Валидируем данные
	if err = h.validator.Validate(pass); err != nil {
		return &response, status.Error(codes.InvalidArgument, h.validator.TransErrors(err).Error())
	}

	if err = h.rep.PasswordUpdate(ctx, &pass); err != nil {
		if errors.Is(err, repository.ErrConflictUnique) {
			return &response, status.Error(codes.AlreadyExists, "запись с таким названием уже существует")
		} else if errors.Is(err, sql.ErrNoRows) {
			return &response, status.Error(codes.NotFound, "запись не найдена")
		}

		return &response, status.Error(codes.Internal, "ошибка при обновлении записи")
	}

	response.SetPassword(
		password.Password_builder{
			Id:       &pass.Id,
			Name:     &pass.Name,
			Login:    &pass.Login,
			Password: &pass.Password,
		}.Build(),
	)

	return &response, nil
}

// PasswordDelete удаление записи пароля по идентификатору
func (h *Handler) PasswordDelete(ctx context.Context, in *proto.IDRequest) (*password.Response, error) {
	var response password.Response

	userID, err := auth.UserIDFromContext(ctx)
	if err != nil || userID == "" {
		return &response, status.Error(codes.Unauthenticated, "пользователь не аутентифицирован")
	}

	id := strings.TrimSpace(in.GetId())
	if id == "" {
		return &response, status.Error(codes.InvalidArgument, "id не указан")
	}

	item, err := h.rep.PasswordDelete(ctx, userID, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &response, status.Error(codes.NotFound, "запись не найдена")
		}
		return &response, status.Error(codes.Internal, "ошибка при удалении записи")
	}

	response.SetPassword(
		password.Password_builder{
			Id:       &item.Id,
			Name:     &item.Name,
			Login:    &item.Login,
			Password: &item.Password,
		}.Build(),
	)

	return &response, nil
}
