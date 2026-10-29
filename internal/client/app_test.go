package client

import (
	"context"
	"errors"
	"testing"

	pb "github.com/alexsey-popov/gkeep/internal/handler/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type mockUsersClient struct {
	registerFunc func(ctx context.Context, in *pb.UserRequest, opts ...grpc.CallOption) (*pb.UserResponse, error)
	loginFunc    func(ctx context.Context, in *pb.UserRequest, opts ...grpc.CallOption) (*pb.UserResponse, error)
}

func (m *mockUsersClient) RegisterUser(ctx context.Context, in *pb.UserRequest, opts ...grpc.CallOption) (*UserResponse, error) {
	if m.registerFunc != nil {
		return m.registerFunc(ctx, in, opts...)
	}
	return nil, errors.New("not implemented")
}

func (m *mockUsersClient) LoginUser(ctx context.Context, in *pb.UserRequest, opts ...grpc.CallOption) (*UserResponse, error) {
	if m.loginFunc != nil {
		return m.loginFunc(ctx, in, opts...)
	}
	return nil, errors.New("not implemented")
}

type UserResponse = pb.UserResponse

func TestNewApp(t *testing.T) {
	mockClient := &mockUsersClient{}
	app := NewApp(mockClient)

	require.NotNil(t, app)
	assert.NotNil(t, app.tviewApp)
	assert.NotNil(t, app.pages)
	assert.NotNil(t, app.loginForm)
	assert.NotNil(t, app.registerForm)
	assert.NotNil(t, app.userView)
	assert.Nil(t, app.GetSession())
}

func TestApp_Login(t *testing.T) {
	t.Run("empty fields validation", func(t *testing.T) {
		mockClient := &mockUsersClient{}
		app := NewApp(mockClient)

		app.loginUserField.SetText("")
		app.loginPassField.SetText("")
		app.handleLogin()

		assert.Nil(t, app.GetSession())
		assert.True(t, app.pages.HasPage("error_modal"))
	})

	t.Run("server error on login", func(t *testing.T) {
		mockClient := &mockUsersClient{
			loginFunc: func(ctx context.Context, in *pb.UserRequest, opts ...grpc.CallOption) (*pb.UserResponse, error) {
				return nil, status.Error(codes.Unauthenticated, "неверный логин или пароль")
			},
		}
		app := NewApp(mockClient)

		app.loginUserField.SetText("user1")
		app.loginPassField.SetText("wrongpass")
		app.handleLogin()

		assert.Nil(t, app.GetSession())
		assert.True(t, app.pages.HasPage("error_modal"))
	})

	t.Run("successful login with token header", func(t *testing.T) {
		mockClient := &mockUsersClient{
			loginFunc: func(ctx context.Context, in *pb.UserRequest, opts ...grpc.CallOption) (*pb.UserResponse, error) {
				assert.Equal(t, "user1", in.GetLogin())
				assert.Equal(t, "pass123", in.GetPassword())

				for _, opt := range opts {
					if headerOpt, ok := opt.(grpc.HeaderCallOption); ok {
						*headerOpt.HeaderAddr = metadata.Pairs("token", "jwt-token-123")
					}
				}

				id := "uuid-1"
				login := "user1"
				resp := &pb.UserResponse{}
				resp.SetUser(pb.User_builder{
					Id:    &id,
					Login: &login,
				}.Build())
				return resp, nil
			},
		}
		app := NewApp(mockClient)

		app.loginUserField.SetText("user1")
		app.loginPassField.SetText("pass123")
		app.handleLogin()

		session := app.GetSession()
		require.NotNil(t, session)
		assert.Equal(t, "uuid-1", session.UserID)
		assert.Equal(t, "user1", session.Login)
		assert.Equal(t, "jwt-token-123", session.Token)
	})
}

func TestApp_Register(t *testing.T) {
	t.Run("empty fields validation", func(t *testing.T) {
		mockClient := &mockUsersClient{}
		app := NewApp(mockClient)

		app.regUserField.SetText("")
		app.regPassField.SetText("")
		app.regPassField2.SetText("")
		app.handleRegister()

		assert.Nil(t, app.GetSession())
		assert.True(t, app.pages.HasPage("error_modal"))
	})

	t.Run("passwords mismatch validation", func(t *testing.T) {
		mockClient := &mockUsersClient{}
		app := NewApp(mockClient)

		app.regUserField.SetText("user1")
		app.regPassField.SetText("pass1")
		app.regPassField2.SetText("pass2")
		app.handleRegister()

		assert.Nil(t, app.GetSession())
		assert.True(t, app.pages.HasPage("error_modal"))
	})

	t.Run("server error on register", func(t *testing.T) {
		mockClient := &mockUsersClient{
			registerFunc: func(ctx context.Context, in *pb.UserRequest, opts ...grpc.CallOption) (*pb.UserResponse, error) {
				return nil, status.Error(codes.AlreadyExists, "логин занят")
			},
		}
		app := NewApp(mockClient)

		app.regUserField.SetText("user1")
		app.regPassField.SetText("pass1")
		app.regPassField2.SetText("pass1")
		app.handleRegister()

		assert.Nil(t, app.GetSession())
		assert.True(t, app.pages.HasPage("error_modal"))
	})

	t.Run("successful register with bearer authorization header", func(t *testing.T) {
		mockClient := &mockUsersClient{
			registerFunc: func(ctx context.Context, in *pb.UserRequest, opts ...grpc.CallOption) (*pb.UserResponse, error) {
				assert.Equal(t, "newuser", in.GetLogin())
				assert.Equal(t, "secretpass", in.GetPassword())

				for _, opt := range opts {
					if headerOpt, ok := opt.(grpc.HeaderCallOption); ok {
						*headerOpt.HeaderAddr = metadata.Pairs("authorization", "Bearer auth-token-456")
					}
				}

				id := "uuid-new"
				login := "newuser"
				resp := &pb.UserResponse{}
				resp.SetUser(pb.User_builder{
					Id:    &id,
					Login: &login,
				}.Build())
				return resp, nil
			},
		}
		app := NewApp(mockClient)

		app.regUserField.SetText("newuser")
		app.regPassField.SetText("secretpass")
		app.regPassField2.SetText("secretpass")
		app.handleRegister()

		session := app.GetSession()
		require.NotNil(t, session)
		assert.Equal(t, "uuid-new", session.UserID)
		assert.Equal(t, "newuser", session.Login)
		assert.Equal(t, "auth-token-456", session.Token)
	})
}

func TestExtractTokenFromHeader(t *testing.T) {
	t.Run("token header", func(t *testing.T) {
		md := metadata.Pairs("token", "my-token")
		assert.Equal(t, "my-token", extractTokenFromHeader(md))
	})

	t.Run("authorization bearer header", func(t *testing.T) {
		md := metadata.Pairs("authorization", "Bearer my-auth-token")
		assert.Equal(t, "my-auth-token", extractTokenFromHeader(md))
	})

	t.Run("empty header", func(t *testing.T) {
		md := metadata.Pairs()
		assert.Equal(t, "", extractTokenFromHeader(md))
	})
}
