package handler

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/alexsey-popov/gkeep/internal/auth"
	"github.com/alexsey-popov/gkeep/internal/repository"
	"github.com/alexsey-popov/gkeep/proto/password"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// PasswordList получение списка записей пользователя
func (h *Handler) PasswordList(ctx context.Context, in *password.PasswordListRequest) (*password.PasswordListResponse, error) {
	var response password.PasswordListResponse

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
func (h *Handler) PasswordCreate(ctx context.Context, in *password.PasswordCreateRequest) (*password.PasswordResponse, error) {
	var response password.PasswordResponse

	userID, err := auth.UserIDFromContext(ctx)
	if err != nil || userID == "" {
		return &response, status.Error(codes.Unauthenticated, "пользователь не аутентифицирован")
	}

	name := strings.TrimSpace(in.GetName())
	if name == "" {
		return &response, status.Error(codes.InvalidArgument, "название не может быть пустым")
	}

	login := strings.TrimSpace(in.GetLogin())
	pwd := in.GetPassword()

	item, err := h.rep.PasswordCreate(ctx, userID, name, login, pwd)
	if err != nil {
		if errors.Is(err, repository.ErrConflictUnique) {
			return &response, status.Error(codes.AlreadyExists, "запись с таким названием уже существует")
		}
		return &response, status.Error(codes.Internal, "ошибка при создании записи")
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

// PasswordUpdate обновление пароля по логину
func (h *Handler) PasswordUpdate(ctx context.Context, in *password.PasswordUpdateRequest) (*password.PasswordResponse, error) {
	var response password.PasswordResponse

	userID, err := auth.UserIDFromContext(ctx)
	if err != nil || userID == "" {
		return &response, status.Error(codes.Unauthenticated, "пользователь не аутентифицирован")
	}

	login := strings.TrimSpace(in.GetLogin())
	if login == "" {
		return &response, status.Error(codes.InvalidArgument, "логин не может быть пустым")
	}

	pwd := in.GetPassword()

	item, err := h.rep.PasswordUpdate(ctx, userID, login, pwd)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &response, status.Error(codes.NotFound, "запись не найдена")
		}
		return &response, status.Error(codes.Internal, "ошибка при обновлении записи")
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

// PasswordDelete удаление записи пароля по идентификатору
func (h *Handler) PasswordDelete(ctx context.Context, in *password.PasswordDeleteRequest) (*password.PasswordResponse, error) {
	var response password.PasswordResponse

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
