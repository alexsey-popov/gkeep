package client

import (
	"context"
	"errors"
	"testing"

	pbPassword "github.com/alexsey-popov/gkeep/proto/password"
	pb "github.com/alexsey-popov/gkeep/proto/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type mockUsersClient struct {
	registerFunc       func(ctx context.Context, in *pb.UserRequest, opts ...grpc.CallOption) (*pb.UserResponse, error)
	loginFunc          func(ctx context.Context, in *pb.UserRequest, opts ...grpc.CallOption) (*pb.UserResponse, error)
	passwordListFunc   func(ctx context.Context, in *pbPassword.PasswordListRequest, opts ...grpc.CallOption) (*pbPassword.PasswordListResponse, error)
	passwordCreateFunc func(ctx context.Context, in *pbPassword.PasswordCreateRequest, opts ...grpc.CallOption) (*pbPassword.PasswordResponse, error)
	passwordUpdateFunc func(ctx context.Context, in *pbPassword.PasswordUpdateRequest, opts ...grpc.CallOption) (*pbPassword.PasswordResponse, error)
	passwordDeleteFunc func(ctx context.Context, in *pbPassword.PasswordDeleteRequest, opts ...grpc.CallOption) (*pbPassword.PasswordResponse, error)
}

func (m *mockUsersClient) RegisterUser(ctx context.Context, in *pb.UserRequest, opts ...grpc.CallOption) (*pb.UserResponse, error) {
	if m.registerFunc != nil {
		return m.registerFunc(ctx, in, opts...)
	}
	return nil, errors.New("not implemented")
}

func (m *mockUsersClient) LoginUser(ctx context.Context, in *pb.UserRequest, opts ...grpc.CallOption) (*pb.UserResponse, error) {
	if m.loginFunc != nil {
		return m.loginFunc(ctx, in, opts...)
	}
	return nil, errors.New("not implemented")
}

func (m *mockUsersClient) PasswordList(ctx context.Context, in *pbPassword.PasswordListRequest, opts ...grpc.CallOption) (*pbPassword.PasswordListResponse, error) {
	if m.passwordListFunc != nil {
		return m.passwordListFunc(ctx, in, opts...)
	}
	return nil, errors.New("not implemented")
}

func (m *mockUsersClient) PasswordCreate(ctx context.Context, in *pbPassword.PasswordCreateRequest, opts ...grpc.CallOption) (*pbPassword.PasswordResponse, error) {
	if m.passwordCreateFunc != nil {
		return m.passwordCreateFunc(ctx, in, opts...)
	}
	return nil, errors.New("not implemented")
}

func (m *mockUsersClient) PasswordUpdate(ctx context.Context, in *pbPassword.PasswordUpdateRequest, opts ...grpc.CallOption) (*pbPassword.PasswordResponse, error) {
	if m.passwordUpdateFunc != nil {
		return m.passwordUpdateFunc(ctx, in, opts...)
	}
	return nil, errors.New("not implemented")
}

func (m *mockUsersClient) PasswordDelete(ctx context.Context, in *pbPassword.PasswordDeleteRequest, opts ...grpc.CallOption) (*pbPassword.PasswordResponse, error) {
	if m.passwordDeleteFunc != nil {
		return m.passwordDeleteFunc(ctx, in, opts...)
	}
	return nil, errors.New("not implemented")
}

func TestNewApp(t *testing.T) {
	mockClient := &mockUsersClient{}
	app := NewApp(mockClient)

	require.NotNil(t, app)
	assert.NotNil(t, app.tviewApp)
	assert.NotNil(t, app.pages)
	assert.NotNil(t, app.loginForm)
	assert.NotNil(t, app.registerForm)
	assert.NotNil(t, app.userView)
	assert.NotNil(t, app.dataTypesFlex)
	assert.NotNil(t, app.passwordsFlex)
	assert.NotNil(t, app.passwordViewFlex)
	assert.NotNil(t, app.passwordCreateForm)
	assert.NotNil(t, app.passwordEditForm)
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

func TestApp_PasswordsOperations(t *testing.T) {
	id1 := "pwd-1"
	name1 := "Google"
	login1 := "user@gmail.com"
	pass1 := "googlePass123"

	mockClient := &mockUsersClient{
		passwordListFunc: func(ctx context.Context, in *pbPassword.PasswordListRequest, opts ...grpc.CallOption) (*pbPassword.PasswordListResponse, error) {
			resp := &pbPassword.PasswordListResponse{}
			resp.SetPassword([]*pbPassword.Password{
				pbPassword.Password_builder{
					Id:       &id1,
					Name:     &name1,
					Login:    &login1,
					Password: &pass1,
				}.Build(),
			})
			return resp, nil
		},
		passwordCreateFunc: func(ctx context.Context, in *pbPassword.PasswordCreateRequest, opts ...grpc.CallOption) (*pbPassword.PasswordResponse, error) {
			if in.GetName() == "" {
				return nil, status.Error(codes.InvalidArgument, "пустое имя")
			}
			id := "pwd-2"
			n := in.GetName()
			l := in.GetLogin()
			p := in.GetPassword()
			resp := &pbPassword.PasswordResponse{}
			resp.SetPassword(pbPassword.Password_builder{
				Id:       &id,
				Name:     &n,
				Login:    &l,
				Password: &p,
			}.Build())
			return resp, nil
		},
		passwordUpdateFunc: func(ctx context.Context, in *pbPassword.PasswordUpdateRequest, opts ...grpc.CallOption) (*pbPassword.PasswordResponse, error) {
			id := "pwd-1"
			n := "Google"
			l := in.GetLogin()
			p := in.GetPassword()
			resp := &pbPassword.PasswordResponse{}
			resp.SetPassword(pbPassword.Password_builder{
				Id:       &id,
				Name:     &n,
				Login:    &l,
				Password: &p,
			}.Build())
			return resp, nil
		},
		passwordDeleteFunc: func(ctx context.Context, in *pbPassword.PasswordDeleteRequest, opts ...grpc.CallOption) (*pbPassword.PasswordResponse, error) {
			resp := &pbPassword.PasswordResponse{}
			return resp, nil
		},
	}

	app := NewApp(mockClient)
	app.session = &Session{Token: "test-token", UserID: "uid-1", Login: "user1"}

	t.Run("load passwords list", func(t *testing.T) {
		app.loadPasswords()
		assert.Len(t, app.currentPasswords, 1)
		assert.Equal(t, "Google", app.currentPasswords[0].GetName())
	})

	t.Run("open password view", func(t *testing.T) {
		app.openPasswordView(app.currentPasswords[0])
		assert.Equal(t, app.currentPasswords[0], app.currentPassword)
		assert.Contains(t, app.passwordDetailsText.GetText(true), "Google")
		assert.Contains(t, app.passwordDetailsText.GetText(true), "user@gmail.com")
	})

	t.Run("create password validation empty name", func(t *testing.T) {
		app.createPassNameField.SetText("")
		app.createPassLoginField.SetText("login")
		app.createPassValField.SetText("pass")
		app.handlePasswordCreate()

		assert.True(t, app.pages.HasPage("error_modal"))
		app.pages.RemovePage("error_modal")
	})

	t.Run("create password success", func(t *testing.T) {
		app.createPassNameField.SetText("Yandex")
		app.createPassLoginField.SetText("yuser")
		app.createPassValField.SetText("ypass")
		app.handlePasswordCreate()

		assert.False(t, app.pages.HasPage("error_modal"))
	})

	t.Run("edit password update", func(t *testing.T) {
		app.openPasswordEdit(app.currentPasswords[0])
		assert.Equal(t, "Google", app.editPassNameField.GetText())
		assert.Equal(t, "user@gmail.com", app.editPassLoginField.GetText())

		app.editPassValField.SetText("newGooglePass")
		app.handlePasswordUpdate()

		assert.Equal(t, "newGooglePass", app.currentPassword.GetPassword())
	})

	t.Run("delete password success", func(t *testing.T) {
		app.currentPassword = app.currentPasswords[0]
		app.handlePasswordDelete()

		assert.Nil(t, app.currentPassword)
	})

	t.Run("logout clears session", func(t *testing.T) {
		app.handleLogout()
		assert.Nil(t, app.GetSession())
	})

	t.Run("unauthenticated operations show error", func(t *testing.T) {
		app.session = nil
		app.handlePasswordCreate()
		assert.True(t, app.pages.HasPage("error_modal"))
		app.pages.RemovePage("error_modal")

		app.handlePasswordUpdate()
		assert.True(t, app.pages.HasPage("error_modal"))
		app.pages.RemovePage("error_modal")

		app.currentPassword = &pbPassword.Password{}
		app.handlePasswordDelete()
		assert.True(t, app.pages.HasPage("error_modal"))
		app.pages.RemovePage("error_modal")
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

	t.Run("token in trailer", func(t *testing.T) {
		header := metadata.Pairs()
		trailer := metadata.Pairs("token", "trailer-token")
		assert.Equal(t, "trailer-token", extractTokenFromHeader(header, trailer))
	})

	t.Run("jwt header", func(t *testing.T) {
		md := metadata.Pairs("jwt", "jwt-token")
		assert.Equal(t, "jwt-token", extractTokenFromHeader(md))
	})

	t.Run("empty header", func(t *testing.T) {
		md := metadata.Pairs()
		assert.Equal(t, "", extractTokenFromHeader(md))
	})
}
