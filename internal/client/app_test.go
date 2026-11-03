package client

import (
	"context"
	"errors"
	"testing"

	pbCard "github.com/alexsey-popov/gkeep/proto/card"
	pbPassword "github.com/alexsey-popov/gkeep/proto/password"
	pbText "github.com/alexsey-popov/gkeep/proto/text"
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
	passwordShowFunc   func(ctx context.Context, in *pbPassword.PasswordShowRequest, opts ...grpc.CallOption) (*pbPassword.PasswordResponse, error)
	passwordUpdateFunc func(ctx context.Context, in *pbPassword.PasswordUpdateRequest, opts ...grpc.CallOption) (*pbPassword.PasswordResponse, error)
	passwordDeleteFunc func(ctx context.Context, in *pbPassword.PasswordDeleteRequest, opts ...grpc.CallOption) (*pbPassword.PasswordResponse, error)
	cardListFunc       func(ctx context.Context, in *pbCard.CardListRequest, opts ...grpc.CallOption) (*pbCard.CardListResponse, error)
	cardCreateFunc     func(ctx context.Context, in *pbCard.CardCreateRequest, opts ...grpc.CallOption) (*pbCard.CardResponse, error)
	cardShowFunc       func(ctx context.Context, in *pbCard.CardShowRequest, opts ...grpc.CallOption) (*pbCard.CardResponse, error)
	cardUpdateFunc     func(ctx context.Context, in *pbCard.CardUpdateRequest, opts ...grpc.CallOption) (*pbCard.CardResponse, error)
	cardDeleteFunc     func(ctx context.Context, in *pbCard.CardDeleteRequest, opts ...grpc.CallOption) (*pbCard.CardResponse, error)
	textListFunc       func(ctx context.Context, in *pbText.TextListRequest, opts ...grpc.CallOption) (*pbText.TextListResponse, error)
	textCreateFunc     func(ctx context.Context, in *pbText.TextCreateRequest, opts ...grpc.CallOption) (*pbText.TextResponse, error)
	textShowFunc       func(ctx context.Context, in *pbText.TextShowRequest, opts ...grpc.CallOption) (*pbText.TextResponse, error)
	textUpdateFunc     func(ctx context.Context, in *pbText.TextUpdateRequest, opts ...grpc.CallOption) (*pbText.TextResponse, error)
	textDeleteFunc     func(ctx context.Context, in *pbText.TextDeleteRequest, opts ...grpc.CallOption) (*pbText.TextResponse, error)
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

func (m *mockUsersClient) PasswordShow(ctx context.Context, in *pbPassword.PasswordShowRequest, opts ...grpc.CallOption) (*pbPassword.PasswordResponse, error) {
	if m.passwordShowFunc != nil {
		return m.passwordShowFunc(ctx, in, opts...)
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

func (m *mockUsersClient) CardList(ctx context.Context, in *pbCard.CardListRequest, opts ...grpc.CallOption) (*pbCard.CardListResponse, error) {
	if m.cardListFunc != nil {
		return m.cardListFunc(ctx, in, opts...)
	}
	return nil, errors.New("not implemented")
}

func (m *mockUsersClient) CardCreate(ctx context.Context, in *pbCard.CardCreateRequest, opts ...grpc.CallOption) (*pbCard.CardResponse, error) {
	if m.cardCreateFunc != nil {
		return m.cardCreateFunc(ctx, in, opts...)
	}
	return nil, errors.New("not implemented")
}

func (m *mockUsersClient) CardShow(ctx context.Context, in *pbCard.CardShowRequest, opts ...grpc.CallOption) (*pbCard.CardResponse, error) {
	if m.cardShowFunc != nil {
		return m.cardShowFunc(ctx, in, opts...)
	}
	return nil, errors.New("not implemented")
}

func (m *mockUsersClient) CardUpdate(ctx context.Context, in *pbCard.CardUpdateRequest, opts ...grpc.CallOption) (*pbCard.CardResponse, error) {
	if m.cardUpdateFunc != nil {
		return m.cardUpdateFunc(ctx, in, opts...)
	}
	return nil, errors.New("not implemented")
}

func (m *mockUsersClient) CardDelete(ctx context.Context, in *pbCard.CardDeleteRequest, opts ...grpc.CallOption) (*pbCard.CardResponse, error) {
	if m.cardDeleteFunc != nil {
		return m.cardDeleteFunc(ctx, in, opts...)
	}
	return nil, errors.New("not implemented")
}

func (m *mockUsersClient) TextList(ctx context.Context, in *pbText.TextListRequest, opts ...grpc.CallOption) (*pbText.TextListResponse, error) {
	if m.textListFunc != nil {
		return m.textListFunc(ctx, in, opts...)
	}
	return nil, errors.New("not implemented")
}

func (m *mockUsersClient) TextCreate(ctx context.Context, in *pbText.TextCreateRequest, opts ...grpc.CallOption) (*pbText.TextResponse, error) {
	if m.textCreateFunc != nil {
		return m.textCreateFunc(ctx, in, opts...)
	}
	return nil, errors.New("not implemented")
}

func (m *mockUsersClient) TextShow(ctx context.Context, in *pbText.TextShowRequest, opts ...grpc.CallOption) (*pbText.TextResponse, error) {
	if m.textShowFunc != nil {
		return m.textShowFunc(ctx, in, opts...)
	}
	return nil, errors.New("not implemented")
}

func (m *mockUsersClient) TextUpdate(ctx context.Context, in *pbText.TextUpdateRequest, opts ...grpc.CallOption) (*pbText.TextResponse, error) {
	if m.textUpdateFunc != nil {
		return m.textUpdateFunc(ctx, in, opts...)
	}
	return nil, errors.New("not implemented")
}

func (m *mockUsersClient) TextDelete(ctx context.Context, in *pbText.TextDeleteRequest, opts ...grpc.CallOption) (*pbText.TextResponse, error) {
	if m.textDeleteFunc != nil {
		return m.textDeleteFunc(ctx, in, opts...)
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
	assert.NotNil(t, app.cardsFlex)
	assert.NotNil(t, app.cardViewFlex)
	assert.NotNil(t, app.cardCreateForm)
	assert.NotNil(t, app.cardEditForm)
	assert.NotNil(t, app.textsFlex)
	assert.NotNil(t, app.textViewFlex)
	assert.NotNil(t, app.textCreateForm)
	assert.NotNil(t, app.textEditForm)
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

func TestApp_Version(t *testing.T) {
	mockClient := &mockUsersClient{}

	t.Run("default empty version", func(t *testing.T) {
		app := NewApp(mockClient)
		assert.Equal(t, "", app.GetVersion())
	})

	t.Run("version provided in NewApp", func(t *testing.T) {
		app := NewApp(mockClient, "v1.2.3")
		assert.Equal(t, "v1.2.3", app.GetVersion())
	})

	t.Run("set version dynamically updates info", func(t *testing.T) {
		app := NewApp(mockClient)
		app.session = &Session{Token: "test-token", UserID: "uid-1", Login: "user1"}
		app.updateUserInfo()
		assert.NotContains(t, app.dataTypesUserText.GetText(true), "Версия:")

		app.SetVersion("v2.0.0")
		assert.Equal(t, "v2.0.0", app.GetVersion())
		assert.Contains(t, app.dataTypesUserText.GetText(true), "Версия: v2.0.0")
	})
}

func TestApp_PasswordPagination(t *testing.T) {
	requestedPages := make([]string, 0)
	mockClient := &mockUsersClient{
		passwordListFunc: func(ctx context.Context, in *pbPassword.PasswordListRequest, opts ...grpc.CallOption) (*pbPassword.PasswordListResponse, error) {
			page := in.GetPage()
			requestedPages = append(requestedPages, page)

			resp := &pbPassword.PasswordListResponse{}
			if page == "1" || page == "" {
				pName := "Page1_Item"
				resp.SetPassword([]*pbPassword.Password{
					pbPassword.Password_builder{Name: &pName}.Build(),
				})
			} else if page == "2" {
				pName := "Page2_Item"
				resp.SetPassword([]*pbPassword.Password{
					pbPassword.Password_builder{Name: &pName}.Build(),
				})
			} else {
				resp.SetPassword([]*pbPassword.Password{})
			}
			return resp, nil
		},
	}

	app := NewApp(mockClient)
	app.session = &Session{Token: "test-token", UserID: "uid-1", Login: "user1"}

	t.Run("initial load requests page 1", func(t *testing.T) {
		requestedPages = nil
		app.loadPasswords()
		assert.Equal(t, 1, app.GetCurrentPage())
		require.Len(t, requestedPages, 1)
		assert.Equal(t, "1", requestedPages[0])
		assert.Len(t, app.currentPasswords, 1)
		assert.Equal(t, "Page1_Item", app.currentPasswords[0].GetName())
	})

	t.Run("prev button on page 1 does not decrement", func(t *testing.T) {
		requestedPages = nil
		app.prevPage()
		assert.Equal(t, 1, app.GetCurrentPage())
		assert.Empty(t, requestedPages)
	})

	t.Run("next button increments page and loads page 2", func(t *testing.T) {
		requestedPages = nil
		app.nextPage()
		assert.Equal(t, 2, app.GetCurrentPage())
		require.Len(t, requestedPages, 1)
		assert.Equal(t, "2", requestedPages[0])
		assert.Len(t, app.currentPasswords, 1)
		assert.Equal(t, "Page2_Item", app.currentPasswords[0].GetName())
	})

	t.Run("prev button on page 2 decrements to page 1", func(t *testing.T) {
		requestedPages = nil
		app.prevPage()
		assert.Equal(t, 1, app.GetCurrentPage())
		require.Len(t, requestedPages, 1)
		assert.Equal(t, "1", requestedPages[0])
		assert.Len(t, app.currentPasswords, 1)
		assert.Equal(t, "Page1_Item", app.currentPasswords[0].GetName())
	})

	t.Run("openPasswordList resets to page 1", func(t *testing.T) {
		app.SetCurrentPage(5)
		requestedPages = nil
		app.openPasswordList()
		assert.Equal(t, 1, app.GetCurrentPage())
		require.Len(t, requestedPages, 1)
		assert.Equal(t, "1", requestedPages[0])
	})

	t.Run("empty page on page > 1 shows pagination hint", func(t *testing.T) {
		app.SetCurrentPage(3)
		app.loadPasswords()
		assert.Empty(t, app.currentPasswords)
		assert.Equal(t, 3, app.GetCurrentPage())
	})
}

func TestApp_CardsOperations(t *testing.T) {
	id1 := "card-1"
	name1 := "Tinkoff Black"
	number1 := "2200 7000 1234 5678"
	exp1 := "12/28"
	code1 := "999"

	mockClient := &mockUsersClient{
		cardListFunc: func(ctx context.Context, in *pbCard.CardListRequest, opts ...grpc.CallOption) (*pbCard.CardListResponse, error) {
			resp := &pbCard.CardListResponse{}
			resp.SetCard([]*pbCard.Card{
				pbCard.Card_builder{
					Id:      &id1,
					Name:    &name1,
					Number:  &number1,
					Expired: &exp1,
					Code:    &code1,
				}.Build(),
			})
			return resp, nil
		},
		cardCreateFunc: func(ctx context.Context, in *pbCard.CardCreateRequest, opts ...grpc.CallOption) (*pbCard.CardResponse, error) {
			if in.GetName() == "" {
				return nil, status.Error(codes.InvalidArgument, "пустое имя")
			}
			id := "card-2"
			n := in.GetName()
			num := in.GetNumber()
			exp := in.GetExpired()
			c := in.GetCode()
			resp := &pbCard.CardResponse{}
			resp.SetCard(pbCard.Card_builder{
				Id:      &id,
				Name:    &n,
				Number:  &num,
				Expired: &exp,
				Code:    &c,
			}.Build())
			return resp, nil
		},
		cardUpdateFunc: func(ctx context.Context, in *pbCard.CardUpdateRequest, opts ...grpc.CallOption) (*pbCard.CardResponse, error) {
			id := in.GetId()
			n := in.GetName()
			num := in.GetNumber()
			exp := in.GetExpired()
			c := in.GetCode()
			resp := &pbCard.CardResponse{}
			resp.SetCard(pbCard.Card_builder{
				Id:      &id,
				Name:    &n,
				Number:  &num,
				Expired: &exp,
				Code:    &c,
			}.Build())
			return resp, nil
		},
		cardDeleteFunc: func(ctx context.Context, in *pbCard.CardDeleteRequest, opts ...grpc.CallOption) (*pbCard.CardResponse, error) {
			resp := &pbCard.CardResponse{}
			return resp, nil
		},
	}

	app := NewApp(mockClient)
	app.session = &Session{Token: "test-token", UserID: "uid-1", Login: "user1"}

	t.Run("load cards list", func(t *testing.T) {
		app.loadCards()
		assert.Len(t, app.currentCards, 1)
		assert.Equal(t, "Tinkoff Black", app.currentCards[0].GetName())
	})

	t.Run("open card view", func(t *testing.T) {
		app.openCardView(app.currentCards[0])
		assert.Equal(t, app.currentCards[0], app.currentCard)
		assert.Contains(t, app.cardDetailsText.GetText(true), "Tinkoff Black")
		assert.Contains(t, app.cardDetailsText.GetText(true), "2200 7000 1234 5678")
		assert.Contains(t, app.cardDetailsText.GetText(true), "12/28")
		assert.Contains(t, app.cardDetailsText.GetText(true), "999")
	})

	t.Run("create card validation empty name", func(t *testing.T) {
		app.createCardNameField.SetText("")
		app.createCardNumberField.SetText("1111 2222 3333 4444")
		app.createCardExpiredField.SetText("01/30")
		app.createCardCodeField.SetText("123")
		app.handleCardCreate()

		assert.True(t, app.pages.HasPage("error_modal"))
		app.pages.RemovePage("error_modal")
	})

	t.Run("create card success", func(t *testing.T) {
		app.createCardNameField.SetText("Sberbank Gold")
		app.createCardNumberField.SetText("4276 3800 1234 5678")
		app.createCardExpiredField.SetText("05/29")
		app.createCardCodeField.SetText("777")
		app.handleCardCreate()

		assert.False(t, app.pages.HasPage("error_modal"))
	})

	t.Run("edit card update", func(t *testing.T) {
		app.openCardEdit(app.currentCards[0])
		assert.Equal(t, "Tinkoff Black", app.editCardNameField.GetText())
		assert.Equal(t, "2200 7000 1234 5678", app.editCardNumberField.GetText())
		assert.Equal(t, "12/28", app.editCardExpiredField.GetText())
		assert.Equal(t, "999", app.editCardCodeField.GetText())

		app.editCardNumberField.SetText("2200 7000 9999 8888")
		app.handleCardUpdate()

		assert.Equal(t, "2200 7000 9999 8888", app.currentCard.GetNumber())
	})

	t.Run("delete card success", func(t *testing.T) {
		app.currentCard = app.currentCards[0]
		app.handleCardDelete()

		assert.Nil(t, app.currentCard)
	})

	t.Run("unauthenticated card operations show error", func(t *testing.T) {
		app.session = nil
		app.handleCardCreate()
		assert.True(t, app.pages.HasPage("error_modal"))
		app.pages.RemovePage("error_modal")

		app.handleCardUpdate()
		assert.True(t, app.pages.HasPage("error_modal"))
		app.pages.RemovePage("error_modal")

		app.currentCard = &pbCard.Card{}
		app.handleCardDelete()
		assert.True(t, app.pages.HasPage("error_modal"))
		app.pages.RemovePage("error_modal")
	})
}

func TestApp_CardPagination(t *testing.T) {
	requestedPages := make([]string, 0)
	mockClient := &mockUsersClient{
		cardListFunc: func(ctx context.Context, in *pbCard.CardListRequest, opts ...grpc.CallOption) (*pbCard.CardListResponse, error) {
			page := in.GetPage()
			requestedPages = append(requestedPages, page)

			resp := &pbCard.CardListResponse{}
			if page == "1" || page == "" {
				cName := "Card_Page1_Item"
				resp.SetCard([]*pbCard.Card{
					pbCard.Card_builder{Name: &cName}.Build(),
				})
			} else if page == "2" {
				cName := "Card_Page2_Item"
				resp.SetCard([]*pbCard.Card{
					pbCard.Card_builder{Name: &cName}.Build(),
				})
			} else {
				resp.SetCard([]*pbCard.Card{})
			}
			return resp, nil
		},
	}

	app := NewApp(mockClient)
	app.session = &Session{Token: "test-token", UserID: "uid-1", Login: "user1"}

	t.Run("initial load requests page 1", func(t *testing.T) {
		requestedPages = nil
		app.loadCards()
		assert.Equal(t, 1, app.GetCurrentCardPage())
		require.Len(t, requestedPages, 1)
		assert.Equal(t, "1", requestedPages[0])
		assert.Len(t, app.currentCards, 1)
		assert.Equal(t, "Card_Page1_Item", app.currentCards[0].GetName())
	})

	t.Run("prev button on page 1 does not decrement", func(t *testing.T) {
		requestedPages = nil
		app.prevCardPage()
		assert.Equal(t, 1, app.GetCurrentCardPage())
		assert.Empty(t, requestedPages)
	})

	t.Run("next button increments page and loads page 2", func(t *testing.T) {
		requestedPages = nil
		app.nextCardPage()
		assert.Equal(t, 2, app.GetCurrentCardPage())
		require.Len(t, requestedPages, 1)
		assert.Equal(t, "2", requestedPages[0])
		assert.Len(t, app.currentCards, 1)
		assert.Equal(t, "Card_Page2_Item", app.currentCards[0].GetName())
	})

	t.Run("prev button on page 2 decrements to page 1", func(t *testing.T) {
		requestedPages = nil
		app.prevCardPage()
		assert.Equal(t, 1, app.GetCurrentCardPage())
		require.Len(t, requestedPages, 1)
		assert.Equal(t, "1", requestedPages[0])
		assert.Len(t, app.currentCards, 1)
		assert.Equal(t, "Card_Page1_Item", app.currentCards[0].GetName())
	})

	t.Run("openCardList resets to page 1", func(t *testing.T) {
		app.SetCurrentCardPage(5)
		requestedPages = nil
		app.openCardList()
		assert.Equal(t, 1, app.GetCurrentCardPage())
		require.Len(t, requestedPages, 1)
		assert.Equal(t, "1", requestedPages[0])
	})

	t.Run("empty page on page > 1 shows pagination hint", func(t *testing.T) {
		app.SetCurrentCardPage(3)
		app.loadCards()
		assert.Empty(t, app.currentCards)
		assert.Equal(t, 3, app.GetCurrentCardPage())
	})
}

func TestApp_TextsOperations(t *testing.T) {
	id1 := "text-1"
	name1 := "My Secret Note"
	text1 := "This is a secret note content\nwith multiple lines"

	mockClient := &mockUsersClient{
		textListFunc: func(ctx context.Context, in *pbText.TextListRequest, opts ...grpc.CallOption) (*pbText.TextListResponse, error) {
			resp := &pbText.TextListResponse{}
			resp.SetText([]*pbText.Text{
				pbText.Text_builder{
					Id:   &id1,
					Name: &name1,
					Text: &text1,
				}.Build(),
			})
			return resp, nil
		},
		textCreateFunc: func(ctx context.Context, in *pbText.TextCreateRequest, opts ...grpc.CallOption) (*pbText.TextResponse, error) {
			if in.GetName() == "" {
				return nil, status.Error(codes.InvalidArgument, "пустое имя")
			}
			id := "text-2"
			n := in.GetName()
			val := in.GetText()
			resp := &pbText.TextResponse{}
			resp.SetText(pbText.Text_builder{
				Id:   &id,
				Name: &n,
				Text: &val,
			}.Build())
			return resp, nil
		},
		textUpdateFunc: func(ctx context.Context, in *pbText.TextUpdateRequest, opts ...grpc.CallOption) (*pbText.TextResponse, error) {
			id := in.GetId()
			n := in.GetName()
			val := in.GetText()
			resp := &pbText.TextResponse{}
			resp.SetText(pbText.Text_builder{
				Id:   &id,
				Name: &n,
				Text: &val,
			}.Build())
			return resp, nil
		},
		textDeleteFunc: func(ctx context.Context, in *pbText.TextDeleteRequest, opts ...grpc.CallOption) (*pbText.TextResponse, error) {
			resp := &pbText.TextResponse{}
			return resp, nil
		},
	}

	app := NewApp(mockClient)
	app.session = &Session{Token: "test-token", UserID: "uid-1", Login: "user1"}

	t.Run("load texts list", func(t *testing.T) {
		app.loadTexts()
		assert.Len(t, app.currentTexts, 1)
		assert.Equal(t, "My Secret Note", app.currentTexts[0].GetName())
	})

	t.Run("open text view", func(t *testing.T) {
		app.openTextView(app.currentTexts[0])
		assert.Equal(t, app.currentTexts[0], app.currentText)
		assert.Contains(t, app.textDetailsText.GetText(true), "My Secret Note")
		assert.Contains(t, app.textDetailsText.GetText(true), "This is a secret note content")
	})

	t.Run("create text validation empty name", func(t *testing.T) {
		app.createTextNameField.SetText("")
		app.createTextValField.SetText("Note content")
		app.handleTextCreate()

		assert.True(t, app.pages.HasPage("error_modal"))
		app.pages.RemovePage("error_modal")
	})

	t.Run("create text success", func(t *testing.T) {
		app.createTextNameField.SetText("New Note")
		app.createTextValField.SetText("Some notes here")
		app.handleTextCreate()

		assert.False(t, app.pages.HasPage("error_modal"))
	})

	t.Run("edit text update", func(t *testing.T) {
		app.openTextEdit(app.currentTexts[0])
		assert.Equal(t, "My Secret Note", app.editTextNameField.GetText())
		assert.Equal(t, "This is a secret note content\nwith multiple lines", app.editTextValField.GetText())

		app.editTextValField.SetText("Updated secret note content")
		app.handleTextUpdate()

		assert.Equal(t, "Updated secret note content", app.currentText.GetText())
	})

	t.Run("delete text success", func(t *testing.T) {
		app.currentText = app.currentTexts[0]
		app.handleTextDelete()

		assert.Nil(t, app.currentText)
	})

	t.Run("unauthenticated text operations show error", func(t *testing.T) {
		app.session = nil
		app.handleTextCreate()
		assert.True(t, app.pages.HasPage("error_modal"))
		app.pages.RemovePage("error_modal")

		app.handleTextUpdate()
		assert.True(t, app.pages.HasPage("error_modal"))
		app.pages.RemovePage("error_modal")

		app.currentText = &pbText.Text{}
		app.handleTextDelete()
		assert.True(t, app.pages.HasPage("error_modal"))
		app.pages.RemovePage("error_modal")
	})
}

func TestApp_TextPagination(t *testing.T) {
	requestedPages := make([]string, 0)
	mockClient := &mockUsersClient{
		textListFunc: func(ctx context.Context, in *pbText.TextListRequest, opts ...grpc.CallOption) (*pbText.TextListResponse, error) {
			page := in.GetPage()
			requestedPages = append(requestedPages, page)

			resp := &pbText.TextListResponse{}
			if page == "1" || page == "" {
				tName := "Text_Page1_Item"
				resp.SetText([]*pbText.Text{
					pbText.Text_builder{Name: &tName}.Build(),
				})
			} else if page == "2" {
				tName := "Text_Page2_Item"
				resp.SetText([]*pbText.Text{
					pbText.Text_builder{Name: &tName}.Build(),
				})
			} else {
				resp.SetText([]*pbText.Text{})
			}
			return resp, nil
		},
	}

	app := NewApp(mockClient)
	app.session = &Session{Token: "test-token", UserID: "uid-1", Login: "user1"}

	t.Run("initial load requests page 1", func(t *testing.T) {
		requestedPages = nil
		app.loadTexts()
		assert.Equal(t, 1, app.GetCurrentTextPage())
		require.Len(t, requestedPages, 1)
		assert.Equal(t, "1", requestedPages[0])
		assert.Len(t, app.currentTexts, 1)
		assert.Equal(t, "Text_Page1_Item", app.currentTexts[0].GetName())
	})

	t.Run("prev button on page 1 does not decrement", func(t *testing.T) {
		requestedPages = nil
		app.prevTextPage()
		assert.Equal(t, 1, app.GetCurrentTextPage())
		assert.Empty(t, requestedPages)
	})

	t.Run("next button increments page and loads page 2", func(t *testing.T) {
		requestedPages = nil
		app.nextTextPage()
		assert.Equal(t, 2, app.GetCurrentTextPage())
		require.Len(t, requestedPages, 1)
		assert.Equal(t, "2", requestedPages[0])
		assert.Len(t, app.currentTexts, 1)
		assert.Equal(t, "Text_Page2_Item", app.currentTexts[0].GetName())
	})

	t.Run("prev button on page 2 decrements to page 1", func(t *testing.T) {
		requestedPages = nil
		app.prevTextPage()
		assert.Equal(t, 1, app.GetCurrentTextPage())
		require.Len(t, requestedPages, 1)
		assert.Equal(t, "1", requestedPages[0])
		assert.Len(t, app.currentTexts, 1)
		assert.Equal(t, "Text_Page1_Item", app.currentTexts[0].GetName())
	})

	t.Run("openTextList resets to page 1", func(t *testing.T) {
		app.SetCurrentTextPage(5)
		requestedPages = nil
		app.openTextList()
		assert.Equal(t, 1, app.GetCurrentTextPage())
		require.Len(t, requestedPages, 1)
		assert.Equal(t, "1", requestedPages[0])
	})

	t.Run("empty page on page > 1 shows pagination hint", func(t *testing.T) {
		app.SetCurrentTextPage(3)
		app.loadTexts()
		assert.Empty(t, app.currentTexts)
		assert.Equal(t, 3, app.GetCurrentTextPage())
	})
}
