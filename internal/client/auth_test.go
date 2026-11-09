package client

import (
	"context"
	"testing"

	pbFile "github.com/alexsey-popov/gkeep/proto/file"
	pb "github.com/alexsey-popov/gkeep/proto/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestApp_Login(t *testing.T) {
	t.Run("negative - валидация пустой формы логина", func(t *testing.T) {
		mockClient := &mockUsersClient{}
		app := NewApp(mockClient)

		app.loginUserField.SetText("")
		app.loginPassField.SetText("")
		app.handleLogin()

		assert.Nil(t, app.GetSession())
		assert.True(t, app.pages.HasPage("error_modal"))
	})

	t.Run("negative - ошибка сервера при логине", func(t *testing.T) {
		mockClient := &mockUsersClient{
			loginFunc: func(ctx context.Context, in *pb.Request, opts ...grpc.CallOption) (*pb.Response, error) {
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

	t.Run("negative - сервер вернул данных без токена аутентификации", func(t *testing.T) {
		mockClient := &mockUsersClient{
			loginFunc: func(ctx context.Context, in *pb.Request, opts ...grpc.CallOption) (*pb.Response, error) {
				id := "uuid-1"
				login := "user1"
				resp := &pb.Response{}
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

		assert.Nil(t, app.GetSession())
		assert.True(t, app.pages.HasPage("error_modal"))
	})

	t.Run("positive - успешная аутентификация", func(t *testing.T) {
		mockClient := &mockUsersClient{
			loginFunc: func(ctx context.Context, in *pb.Request, opts ...grpc.CallOption) (*pb.Response, error) {
				assert.Equal(t, "user1", in.GetLogin())
				assert.Equal(t, "pass123", in.GetPassword())

				for _, opt := range opts {
					if headerOpt, ok := opt.(grpc.HeaderCallOption); ok {
						*headerOpt.HeaderAddr = metadata.Pairs("token", "jwt-token-123")
					}
				}

				id := "uuid-1"
				login := "user1"
				resp := &pb.Response{}
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
	t.Run("negative - пустые поля в форме регистрации", func(t *testing.T) {
		mockClient := &mockUsersClient{}
		app := NewApp(mockClient)

		app.regUserField.SetText("")
		app.regPassField.SetText("")
		app.regPassField2.SetText("")
		app.handleRegister()

		assert.Nil(t, app.GetSession())
		assert.True(t, app.pages.HasPage("error_modal"))
	})

	t.Run("negative - некорректное подтверждение пароля", func(t *testing.T) {
		mockClient := &mockUsersClient{}
		app := NewApp(mockClient)

		app.regUserField.SetText("user1")
		app.regPassField.SetText("pass1")
		app.regPassField2.SetText("pass2")
		app.handleRegister()

		assert.Nil(t, app.GetSession())
		assert.True(t, app.pages.HasPage("error_modal"))
	})

	t.Run("negative - ошибка сервера при регистрации", func(t *testing.T) {
		mockClient := &mockUsersClient{
			registerFunc: func(ctx context.Context, in *pb.Request, opts ...grpc.CallOption) (*pb.Response, error) {
				return nil, status.Error(codes.AlreadyExists, "пользователь уже существует")
			},
		}
		app := NewApp(mockClient)

		app.regUserField.SetText("existinguser")
		app.regPassField.SetText("secret123")
		app.regPassField2.SetText("secret123")
		app.handleRegister()

		assert.Nil(t, app.GetSession())
		assert.True(t, app.pages.HasPage("error_modal"))
	})

	t.Run("positive - успешная регистрация", func(t *testing.T) {
		mockClient := &mockUsersClient{
			registerFunc: func(ctx context.Context, in *pb.Request, opts ...grpc.CallOption) (*pb.Response, error) {
				for _, opt := range opts {
					if headerOpt, ok := opt.(grpc.HeaderCallOption); ok {
						*headerOpt.HeaderAddr = metadata.Pairs("authorization", "Bearer jwt-token-reg")
					}
				}

				id := "uuid-reg"
				login := "newuser"
				resp := &pb.Response{}
				resp.SetUser(pb.User_builder{
					Id:    &id,
					Login: &login,
				}.Build())
				return resp, nil
			},
		}
		app := NewApp(mockClient)

		app.regUserField.SetText("newuser")
		app.regPassField.SetText("secret123")
		app.regPassField2.SetText("secret123")
		app.handleRegister()

		session := app.GetSession()
		require.NotNil(t, session)
		assert.Equal(t, "uuid-reg", session.UserID)
		assert.Equal(t, "newuser", session.Login)
		assert.Equal(t, "jwt-token-reg", session.Token)
	})
}

func TestApp_Logout(t *testing.T) {
	app := NewApp(&mockUsersClient{})
	app.session = &Session{Token: "some-token", UserID: "uid", Login: "user"}
	app.currentFiles = []*pbFile.File{{}}
	app.currentFile = &pbFile.File{}
	app.currentFilePage = 3
	app.currentCardPage = 2
	app.currentPage = 4
	app.currentTextPage = 5

	app.handleLogout()

	assert.Nil(t, app.GetSession())
	assert.Nil(t, app.currentFiles)
	assert.Nil(t, app.currentFile)
	assert.Equal(t, 1, app.currentFilePage)
	assert.Equal(t, 1, app.currentCardPage)
	assert.Equal(t, 1, app.currentPage)
	assert.Equal(t, 1, app.currentTextPage)
}
