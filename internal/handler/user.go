package handler

import (
	"context"
	"errors"

	"github.com/alexsey-popov/gkeep/internal/repository"
	pb "github.com/alexsey-popov/gkeep/proto/user"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// RegisterUser регистрация пользователя
func (h *Handler) RegisterUser(ctx context.Context, in *pb.UserRequest) (*pb.UserResponse, error) {
	var response pb.UserResponse

	password := in.GetPassword()
	if h.guard != nil {
		hash, err := h.guard.GetHash(password)
		if err != nil {
			return &response, status.Error(codes.Internal, "ошибка при обработке пароля")
		}
		password = hash
	}

	user, err := h.rep.CreateUser(ctx, in.GetLogin(), password)
	if err != nil {
		// Если произошла ошибка уникальности - пишем другое
		if errors.Is(err, repository.ErrConflictUnique) {
			return &response, status.Error(codes.AlreadyExists, "Указанный логин занят")
		}

		// Иначе - логируем и высылаем общее сообщение
		return &response, status.Error(codes.Unknown, "Указанный логин занят")
	}

	// А тут заполняем response данными пользователя
	response.SetUser(
		pb.User_builder{
			Id:    &user.Id,
			Login: &user.Login,
		}.Build(),
	)

	// Добавляем метаданные при успешной регистрации
	if h.guard != nil {
		tokenString, _, err := h.guard.NewUserToken(user.Id)
		if err == nil && tokenString != "" {
			md := metadata.Pairs(
				"authorization", "Bearer "+tokenString,
				"token", tokenString,
			)
			_ = grpc.SetHeader(ctx, md)
			_ = grpc.SendHeader(ctx, md)
			_ = grpc.SetTrailer(ctx, md)
		}
	}

	return &response, nil
}

// LoginUser аутентификация пользователя
func (h *Handler) LoginUser(ctx context.Context, in *pb.UserRequest) (*pb.UserResponse, error) {
	var response pb.UserResponse

	user, err := h.rep.GetUser(ctx, in.GetLogin())
	if err != nil {
		return &response, status.Error(codes.Unknown, err.Error())
	}

	if h.guard != nil {
		if !h.guard.CheckPassword(in.GetPassword(), user.Password) {
			return &response, status.Error(codes.Unauthenticated, "неверный логин или пароль")
		}
	}

	// А тут заполняем response данными пользователя
	response.SetUser(
		pb.User_builder{
			Id:    &user.Id,
			Login: &user.Login,
		}.Build(),
	)

	// Добавляем метаданные при успешной аутентификации
	if h.guard != nil {
		tokenString, _, err := h.guard.NewUserToken(user.Id)
		if err == nil && tokenString != "" {
			md := metadata.Pairs(
				"authorization", "Bearer "+tokenString,
				"token", tokenString,
			)
			_ = grpc.SetHeader(ctx, md)
			_ = grpc.SendHeader(ctx, md)
			_ = grpc.SetTrailer(ctx, md)
		}
	}

	return &response, nil
}
