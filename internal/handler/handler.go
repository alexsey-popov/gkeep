// Обработчик для gRPC запросов
package handler

import (
	"context"
	"errors"

	pb "github.com/alexsey-popov/gkeep/internal/handler/proto"
	"github.com/alexsey-popov/gkeep/internal/repository"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Handler Обработчик
type Handler struct {
	pb.UnimplementedUsersServer

	rep repository.Repository
}

// New создание нового обработчика
func New(rep repository.Repository) *Handler {
	return &Handler{
		rep: rep,
	}
}

// RegisterUser регистрация пользователя
func (h *Handler) RegisterUser(ctx context.Context, in *pb.UserRequest) (*pb.UserResponse, error) {
	var response pb.UserResponse

	user, err := h.rep.CreateUser(ctx, in.GetLogin(), in.GetPassword())
	if err != nil {
		// Если произошла ошибка валидации - пишем одно
		//response.SetError(fmt.Sprintf(`Пользователь с email %s уже существует`, in.GetLogin()))
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

	return &response, nil
}

// LoginUser аутентификация пользователя
func (h *Handler) LoginUser(ctx context.Context, in *pb.UserRequest) (*pb.UserResponse, error) {
	var response pb.UserResponse

	user, err := h.rep.GetUser(ctx, in.GetLogin())
	if err != nil {
		return &response, status.Error(codes.Unknown, err.Error())
	}

	// А тут заполняем response данными пользователя
	response.SetUser(
		pb.User_builder{
			Id:    &user.Id,
			Login: &user.Login,
		}.Build(),
	)

	return &response, nil
}
