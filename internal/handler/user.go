// Обработчик для gRPC запросов
package handler

import (
	"context"
	"errors"

	"github.com/alexsey-popov/gkeep/internal/models"
	"github.com/alexsey-popov/gkeep/internal/repository"
	pb "github.com/alexsey-popov/gkeep/proto/user"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// RegisterUser регистрация пользователя
func (h *Handler) RegisterUser(ctx context.Context, in *pb.Request) (*pb.Response, error) {
	var response pb.Response

	// Создаём модель пользователя и проводим валидацию
	user := models.User{
		Login:    in.GetLogin(),
		Password: in.GetPassword(),
	}
	if err := h.validator.Validate(user); err != nil {
		return &response, status.Error(codes.InvalidArgument, h.validator.TransErrors(err).Error())
	}

	// Далее шифруем пароль и подменяем пароль на зашифрованную версию
	if h.guard != nil {
		hash, err := h.guard.GetHash(user.Password)
		if err != nil {
			return &response, status.Error(codes.Internal, "ошибка при обработке пароля")
		}
		user.Password = hash
	}

	err := h.rep.CreateUser(ctx, &user)
	if err != nil {
		// Если произошла ошибка уникальности - пишем другое
		if errors.Is(err, repository.ErrConflictUnique) {
			return &response, status.Error(codes.AlreadyExists, "Указанный логин занят")
		}

		// Иначе высылаем общее сообщение об ошибке (логирование происходит на уровне репозитория)
		return &response, status.Error(codes.Internal, "Ошибка на стороне сервера")
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
func (h *Handler) LoginUser(ctx context.Context, in *pb.Request) (*pb.Response, error) {
	var response pb.Response

	user, err := h.rep.GetUser(ctx, in.GetLogin())
	if err != nil {
		return &response, status.Error(codes.Internal, "Ошибка на стороне сервера")
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
