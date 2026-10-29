package handler

import (
	"context"

	"github.com/alexsey-popov/gkeep/proto/password"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (h *Handler) PasswordList(context.Context, *password.PasswordListRequest) (*password.PasswordListResponse, error) {
	return nil, status.Error(codes.NotFound, "method PasswordList not implemented")
}
func (h *Handler) PasswordCreate(context.Context, *password.PasswordCreateRequest) (*password.PasswordResponse, error) {
	return nil, status.Error(codes.NotFound, "method PasswordCreate not implemented")
}
func (h *Handler) PasswordUpdate(context.Context, *password.PasswordUpdateRequest) (*password.PasswordResponse, error) {
	return nil, status.Error(codes.NotFound, "method PasswordUpdate not implemented")
}
func (h *Handler) PasswordDelete(context.Context, *password.PasswordDeleteRequest) (*password.PasswordResponse, error) {
	return nil, status.Error(codes.NotFound, "method PasswordDelete not implemented")
}
