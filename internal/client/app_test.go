package client

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"

	api "github.com/alexsey-popov/gkeep/proto"
	pbCard "github.com/alexsey-popov/gkeep/proto/card"
	pbFile "github.com/alexsey-popov/gkeep/proto/file"
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
	registerFunc       func(ctx context.Context, in *pb.Request, opts ...grpc.CallOption) (*pb.Response, error)
	loginFunc          func(ctx context.Context, in *pb.Request, opts ...grpc.CallOption) (*pb.Response, error)
	passwordListFunc   func(ctx context.Context, in *api.ListRequest, opts ...grpc.CallOption) (*pbPassword.ListResponse, error)
	passwordCreateFunc func(ctx context.Context, in *pbPassword.CreateRequest, opts ...grpc.CallOption) (*pbPassword.Response, error)
	passwordShowFunc   func(ctx context.Context, in *api.IDRequest, opts ...grpc.CallOption) (*pbPassword.Response, error)
	passwordUpdateFunc func(ctx context.Context, in *pbPassword.UpdateRequest, opts ...grpc.CallOption) (*pbPassword.Response, error)
	passwordDeleteFunc func(ctx context.Context, in *api.IDRequest, opts ...grpc.CallOption) (*pbPassword.Response, error)
	cardListFunc       func(ctx context.Context, in *api.ListRequest, opts ...grpc.CallOption) (*pbCard.ListResponse, error)
	cardCreateFunc     func(ctx context.Context, in *pbCard.CreateRequest, opts ...grpc.CallOption) (*pbCard.Response, error)
	cardShowFunc       func(ctx context.Context, in *api.IDRequest, opts ...grpc.CallOption) (*pbCard.Response, error)
	cardUpdateFunc     func(ctx context.Context, in *pbCard.UpdateRequest, opts ...grpc.CallOption) (*pbCard.Response, error)
	cardDeleteFunc     func(ctx context.Context, in *api.IDRequest, opts ...grpc.CallOption) (*pbCard.Response, error)
	textListFunc       func(ctx context.Context, in *api.ListRequest, opts ...grpc.CallOption) (*pbText.ListResponse, error)
	textCreateFunc     func(ctx context.Context, in *pbText.CreateRequest, opts ...grpc.CallOption) (*pbText.Response, error)
	textShowFunc       func(ctx context.Context, in *api.IDRequest, opts ...grpc.CallOption) (*pbText.Response, error)
	textUpdateFunc     func(ctx context.Context, in *pbText.UpdateRequest, opts ...grpc.CallOption) (*pbText.Response, error)
	textDeleteFunc     func(ctx context.Context, in *api.IDRequest, opts ...grpc.CallOption) (*pbText.Response, error)
	fileListFunc       func(ctx context.Context, in *api.ListRequest, opts ...grpc.CallOption) (*pbFile.ListResponse, error)
	fileShowFunc       func(ctx context.Context, in *api.IDRequest, opts ...grpc.CallOption) (*pbFile.Response, error)
	fileDeleteFunc     func(ctx context.Context, in *api.IDRequest, opts ...grpc.CallOption) (*pbFile.Response, error)
	fileUploadFunc     func(ctx context.Context, opts ...grpc.CallOption) (grpc.ClientStreamingClient[pbFile.UploadFileRequest, pbFile.UploadFileResponse], error)
	fileDownloadFunc   func(ctx context.Context, in *pbFile.DownloadFileRequest, opts ...grpc.CallOption) (grpc.ServerStreamingClient[pbFile.DownloadFileResponse], error)
}

func (m *mockUsersClient) RegisterUser(ctx context.Context, in *pb.Request, opts ...grpc.CallOption) (*pb.Response, error) {
	if m.registerFunc != nil {
		return m.registerFunc(ctx, in, opts...)
	}
	return nil, errors.New("not implemented")
}

func (m *mockUsersClient) LoginUser(ctx context.Context, in *pb.Request, opts ...grpc.CallOption) (*pb.Response, error) {
	if m.loginFunc != nil {
		return m.loginFunc(ctx, in, opts...)
	}
	return nil, errors.New("not implemented")
}

func (m *mockUsersClient) PasswordList(ctx context.Context, in *api.ListRequest, opts ...grpc.CallOption) (*pbPassword.ListResponse, error) {
	if m.passwordListFunc != nil {
		return m.passwordListFunc(ctx, in, opts...)
	}
	return nil, errors.New("not implemented")
}

func (m *mockUsersClient) PasswordCreate(ctx context.Context, in *pbPassword.CreateRequest, opts ...grpc.CallOption) (*pbPassword.Response, error) {
	if m.passwordCreateFunc != nil {
		return m.passwordCreateFunc(ctx, in, opts...)
	}
	return nil, errors.New("not implemented")
}

func (m *mockUsersClient) PasswordShow(ctx context.Context, in *api.IDRequest, opts ...grpc.CallOption) (*pbPassword.Response, error) {
	if m.passwordShowFunc != nil {
		return m.passwordShowFunc(ctx, in, opts...)
	}
	return nil, errors.New("not implemented")
}

func (m *mockUsersClient) PasswordUpdate(ctx context.Context, in *pbPassword.UpdateRequest, opts ...grpc.CallOption) (*pbPassword.Response, error) {
	if m.passwordUpdateFunc != nil {
		return m.passwordUpdateFunc(ctx, in, opts...)
	}
	return nil, errors.New("not implemented")
}

func (m *mockUsersClient) PasswordDelete(ctx context.Context, in *api.IDRequest, opts ...grpc.CallOption) (*pbPassword.Response, error) {
	if m.passwordDeleteFunc != nil {
		return m.passwordDeleteFunc(ctx, in, opts...)
	}
	return nil, errors.New("not implemented")
}

func (m *mockUsersClient) CardList(ctx context.Context, in *api.ListRequest, opts ...grpc.CallOption) (*pbCard.ListResponse, error) {
	if m.cardListFunc != nil {
		return m.cardListFunc(ctx, in, opts...)
	}
	return nil, errors.New("not implemented")
}

func (m *mockUsersClient) CardCreate(ctx context.Context, in *pbCard.CreateRequest, opts ...grpc.CallOption) (*pbCard.Response, error) {
	if m.cardCreateFunc != nil {
		return m.cardCreateFunc(ctx, in, opts...)
	}
	return nil, errors.New("not implemented")
}

func (m *mockUsersClient) CardShow(ctx context.Context, in *api.IDRequest, opts ...grpc.CallOption) (*pbCard.Response, error) {
	if m.cardShowFunc != nil {
		return m.cardShowFunc(ctx, in, opts...)
	}
	return nil, errors.New("not implemented")
}

func (m *mockUsersClient) CardUpdate(ctx context.Context, in *pbCard.UpdateRequest, opts ...grpc.CallOption) (*pbCard.Response, error) {
	if m.cardUpdateFunc != nil {
		return m.cardUpdateFunc(ctx, in, opts...)
	}
	return nil, errors.New("not implemented")
}

func (m *mockUsersClient) CardDelete(ctx context.Context, in *api.IDRequest, opts ...grpc.CallOption) (*pbCard.Response, error) {
	if m.cardDeleteFunc != nil {
		return m.cardDeleteFunc(ctx, in, opts...)
	}
	return nil, errors.New("not implemented")
}

func (m *mockUsersClient) TextList(ctx context.Context, in *api.ListRequest, opts ...grpc.CallOption) (*pbText.ListResponse, error) {
	if m.textListFunc != nil {
		return m.textListFunc(ctx, in, opts...)
	}
	return nil, errors.New("not implemented")
}

func (m *mockUsersClient) TextCreate(ctx context.Context, in *pbText.CreateRequest, opts ...grpc.CallOption) (*pbText.Response, error) {
	if m.textCreateFunc != nil {
		return m.textCreateFunc(ctx, in, opts...)
	}
	return nil, errors.New("not implemented")
}

func (m *mockUsersClient) TextShow(ctx context.Context, in *api.IDRequest, opts ...grpc.CallOption) (*pbText.Response, error) {
	if m.textShowFunc != nil {
		return m.textShowFunc(ctx, in, opts...)
	}
	return nil, errors.New("not implemented")
}

func (m *mockUsersClient) TextUpdate(ctx context.Context, in *pbText.UpdateRequest, opts ...grpc.CallOption) (*pbText.Response, error) {
	if m.textUpdateFunc != nil {
		return m.textUpdateFunc(ctx, in, opts...)
	}
	return nil, errors.New("not implemented")
}

func (m *mockUsersClient) TextDelete(ctx context.Context, in *api.IDRequest, opts ...grpc.CallOption) (*pbText.Response, error) {
	if m.textDeleteFunc != nil {
		return m.textDeleteFunc(ctx, in, opts...)
	}
	return nil, errors.New("not implemented")
}

func (m *mockUsersClient) FileList(ctx context.Context, in *api.ListRequest, opts ...grpc.CallOption) (*pbFile.ListResponse, error) {
	if m.fileListFunc != nil {
		return m.fileListFunc(ctx, in, opts...)
	}
	return nil, errors.New("not implemented")
}

func (m *mockUsersClient) FileShow(ctx context.Context, in *api.IDRequest, opts ...grpc.CallOption) (*pbFile.Response, error) {
	if m.fileShowFunc != nil {
		return m.fileShowFunc(ctx, in, opts...)
	}
	return nil, errors.New("not implemented")
}

func (m *mockUsersClient) FileDelete(ctx context.Context, in *api.IDRequest, opts ...grpc.CallOption) (*pbFile.Response, error) {
	if m.fileDeleteFunc != nil {
		return m.fileDeleteFunc(ctx, in, opts...)
	}
	return nil, errors.New("not implemented")
}

func (m *mockUsersClient) FileUpload(ctx context.Context, opts ...grpc.CallOption) (grpc.ClientStreamingClient[pbFile.UploadFileRequest, pbFile.UploadFileResponse], error) {
	if m.fileUploadFunc != nil {
		return m.fileUploadFunc(ctx, opts...)
	}
	return nil, errors.New("not implemented")
}

func (m *mockUsersClient) FileDownload(ctx context.Context, in *pbFile.DownloadFileRequest, opts ...grpc.CallOption) (grpc.ServerStreamingClient[pbFile.DownloadFileResponse], error) {
	if m.fileDownloadFunc != nil {
		return m.fileDownloadFunc(ctx, in, opts...)
	}
	return nil, errors.New("not implemented")
}

type mockUploadStreamClient struct {
	grpc.ClientStreamingClient[pbFile.UploadFileRequest, pbFile.UploadFileResponse]
	sentReqs     []*pbFile.UploadFileRequest
	sendErr      error
	closeResp    *pbFile.UploadFileResponse
	closeRecvErr error
}

func (m *mockUploadStreamClient) Send(req *pbFile.UploadFileRequest) error {
	if m.sendErr != nil {
		return m.sendErr
	}
	m.sentReqs = append(m.sentReqs, req)
	return nil
}

func (m *mockUploadStreamClient) CloseAndRecv() (*pbFile.UploadFileResponse, error) {
	if m.closeRecvErr != nil {
		return nil, m.closeRecvErr
	}
	return m.closeResp, nil
}

type mockDownloadStreamClient struct {
	grpc.ServerStreamingClient[pbFile.DownloadFileResponse]
	msgs    []*pbFile.DownloadFileResponse
	recvIdx int
	recvErr error
}

func (m *mockDownloadStreamClient) Recv() (*pbFile.DownloadFileResponse, error) {
	if m.recvErr != nil {
		return nil, m.recvErr
	}
	if m.recvIdx >= len(m.msgs) {
		return nil, io.EOF
	}
	msg := m.msgs[m.recvIdx]
	m.recvIdx++
	return msg, nil
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
	assert.NotNil(t, app.filesFlex)
	assert.NotNil(t, app.fileViewFlex)
	assert.NotNil(t, app.fileUploadForm)
	assert.NotNil(t, app.fileDownloadForm)
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

	t.Run("successful login with token header", func(t *testing.T) {
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

	t.Run("successful registration with auth header", func(t *testing.T) {
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

	app.handleLogout()

	assert.Nil(t, app.GetSession())
	assert.Nil(t, app.currentFiles)
	assert.Nil(t, app.currentFile)
	assert.Equal(t, 1, app.currentFilePage)
}

func TestApp_Passwords(t *testing.T) {
	id1 := "p1"
	name1 := "Email"
	login1 := "user@mail.com"
	pass1 := "password123"

	mockClient := &mockUsersClient{
		passwordListFunc: func(ctx context.Context, in *api.ListRequest, opts ...grpc.CallOption) (*pbPassword.ListResponse, error) {
			resp := &pbPassword.ListResponse{}
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
		passwordCreateFunc: func(ctx context.Context, in *pbPassword.CreateRequest, opts ...grpc.CallOption) (*pbPassword.Response, error) {
			id := "p2"
			name := in.GetName()
			login := in.GetLogin()
			pass := in.GetPassword()
			resp := &pbPassword.Response{}
			resp.SetPassword(pbPassword.Password_builder{
				Id:       &id,
				Name:     &name,
				Login:    &login,
				Password: &pass,
			}.Build())
			return resp, nil
		},
		passwordUpdateFunc: func(ctx context.Context, in *pbPassword.UpdateRequest, opts ...grpc.CallOption) (*pbPassword.Response, error) {
			id := in.GetId()
			name := in.GetName()
			login := in.GetLogin()
			pass := in.GetPassword()
			resp := &pbPassword.Response{}
			resp.SetPassword(pbPassword.Password_builder{
				Id:       &id,
				Name:     &name,
				Login:    &login,
				Password: &pass,
			}.Build())
			return resp, nil
		},
		passwordDeleteFunc: func(ctx context.Context, in *api.IDRequest, opts ...grpc.CallOption) (*pbPassword.Response, error) {
			return &pbPassword.Response{}, nil
		},
	}

	app := NewApp(mockClient)
	app.session = &Session{Token: "test-token", UserID: "uid-1", Login: "user1"}

	t.Run("load passwords", func(t *testing.T) {
		app.loadPasswords()
		assert.Len(t, app.currentPasswords, 1)
		assert.Equal(t, "Email", app.currentPasswords[0].GetName())
	})

	t.Run("create password validation", func(t *testing.T) {
		app.createPassNameField.SetText("")
		app.handlePasswordCreate()
		assert.True(t, app.pages.HasPage("error_modal"))
		app.pages.RemovePage("error_modal")
	})

	t.Run("create password success", func(t *testing.T) {
		app.createPassNameField.SetText("GitHub")
		app.createPassLoginField.SetText("gituser")
		app.createPassValField.SetText("gitpass")
		app.handlePasswordCreate()
		assert.False(t, app.pages.HasPage("error_modal"))
	})

	t.Run("edit password update", func(t *testing.T) {
		app.openPasswordEdit(app.currentPasswords[0])
		assert.Equal(t, "Email", app.editPassNameField.GetText())
		assert.Equal(t, "user@mail.com", app.editPassLoginField.GetText())

		app.editPassValField.SetText("new-password-456")
		app.handlePasswordUpdate()

		assert.Equal(t, "new-password-456", app.currentPassword.GetPassword())
	})

	t.Run("delete password success", func(t *testing.T) {
		app.currentPassword = app.currentPasswords[0]
		app.handlePasswordDelete()
		assert.Nil(t, app.currentPassword)
	})
}

func TestApp_Cards(t *testing.T) {
	id1 := "c1"
	name1 := "Bank Visa"
	num1 := "1111222233334444"
	exp1 := "12/28"
	code1 := "123"

	mockClient := &mockUsersClient{
		cardListFunc: func(ctx context.Context, in *api.ListRequest, opts ...grpc.CallOption) (*pbCard.ListResponse, error) {
			resp := &pbCard.ListResponse{}
			resp.SetCard([]*pbCard.Card{
				pbCard.Card_builder{
					Id:      &id1,
					Name:    &name1,
					Number:  &num1,
					Expired: &exp1,
					Code:    &code1,
				}.Build(),
			})
			return resp, nil
		},
		cardCreateFunc: func(ctx context.Context, in *pbCard.CreateRequest, opts ...grpc.CallOption) (*pbCard.Response, error) {
			id := "c2"
			name := in.GetName()
			resp := &pbCard.Response{}
			resp.SetCard(pbCard.Card_builder{
				Id:   &id,
				Name: &name,
			}.Build())
			return resp, nil
		},
		cardUpdateFunc: func(ctx context.Context, in *pbCard.UpdateRequest, opts ...grpc.CallOption) (*pbCard.Response, error) {
			id := in.GetId()
			name := in.GetName()
			num := in.GetNumber()
			exp := in.GetExpired()
			code := in.GetCode()
			resp := &pbCard.Response{}
			resp.SetCard(pbCard.Card_builder{
				Id:      &id,
				Name:    &name,
				Number:  &num,
				Expired: &exp,
				Code:    &code,
			}.Build())
			return resp, nil
		},
		cardDeleteFunc: func(ctx context.Context, in *api.IDRequest, opts ...grpc.CallOption) (*pbCard.Response, error) {
			return &pbCard.Response{}, nil
		},
	}

	app := NewApp(mockClient)
	app.session = &Session{Token: "test-token", UserID: "uid-1", Login: "user1"}

	t.Run("load cards", func(t *testing.T) {
		app.loadCards()
		assert.Len(t, app.currentCards, 1)
		assert.Equal(t, "Bank Visa", app.currentCards[0].GetName())
	})

	t.Run("create card success", func(t *testing.T) {
		app.createCardNameField.SetText("Mastercard")
		app.createCardNumberField.SetText("5555666677778888")
		app.createCardExpiredField.SetText("11/30")
		app.createCardCodeField.SetText("999")
		app.handleCardCreate()
		assert.False(t, app.pages.HasPage("error_modal"))
	})

	t.Run("edit card update", func(t *testing.T) {
		app.openCardEdit(app.currentCards[0])
		assert.Equal(t, "Bank Visa", app.editCardNameField.GetText())
		app.editCardExpiredField.SetText("01/30")
		app.handleCardUpdate()
		assert.Equal(t, "01/30", app.currentCard.GetExpired())
	})

	t.Run("delete card success", func(t *testing.T) {
		app.currentCard = app.currentCards[0]
		app.handleCardDelete()
		assert.Nil(t, app.currentCard)
	})
}

func TestApp_Texts(t *testing.T) {
	id1 := "t1"
	name1 := "My Secret Note"
	val1 := "This is a secret note content"

	mockClient := &mockUsersClient{
		textListFunc: func(ctx context.Context, in *api.ListRequest, opts ...grpc.CallOption) (*pbText.ListResponse, error) {
			resp := &pbText.ListResponse{}
			resp.SetText([]*pbText.Text{
				pbText.Text_builder{
					Id:   &id1,
					Name: &name1,
					Text: &val1,
				}.Build(),
			})
			return resp, nil
		},
		textCreateFunc: func(ctx context.Context, in *pbText.CreateRequest, opts ...grpc.CallOption) (*pbText.Response, error) {
			id := "t2"
			name := in.GetName()
			txt := in.GetText()
			resp := &pbText.Response{}
			resp.SetText(pbText.Text_builder{
				Id:   &id,
				Name: &name,
				Text: &txt,
			}.Build())
			return resp, nil
		},
		textUpdateFunc: func(ctx context.Context, in *pbText.UpdateRequest, opts ...grpc.CallOption) (*pbText.Response, error) {
			id := in.GetId()
			name := in.GetName()
			txt := in.GetText()
			resp := &pbText.Response{}
			resp.SetText(pbText.Text_builder{
				Id:   &id,
				Name: &name,
				Text: &txt,
			}.Build())
			return resp, nil
		},
		textDeleteFunc: func(ctx context.Context, in *api.IDRequest, opts ...grpc.CallOption) (*pbText.Response, error) {
			return &pbText.Response{}, nil
		},
	}

	app := NewApp(mockClient)
	app.session = &Session{Token: "test-token", UserID: "uid-1", Login: "user1"}

	t.Run("load texts", func(t *testing.T) {
		app.loadTexts()
		assert.Len(t, app.currentTexts, 1)
		assert.Equal(t, "My Secret Note", app.currentTexts[0].GetName())
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
		app.editTextValField.SetText("Updated secret note content")
		app.handleTextUpdate()
		assert.Equal(t, "Updated secret note content", app.currentText.GetText())
	})

	t.Run("delete text success", func(t *testing.T) {
		app.currentText = app.currentTexts[0]
		app.handleTextDelete()
		assert.Nil(t, app.currentText)
	})
}

func TestApp_Files(t *testing.T) {
	id1 := "f1"
	name1 := "Скан паспорта"
	filename1 := "passport.pdf"
	size1 := int64(1024)
	checksum1 := "sha256-hash-value"

	mockClient := &mockUsersClient{
		fileListFunc: func(ctx context.Context, in *api.ListRequest, opts ...grpc.CallOption) (*pbFile.ListResponse, error) {
			resp := &pbFile.ListResponse{}
			resp.SetFiles([]*pbFile.File{
				pbFile.File_builder{
					Id:       &id1,
					Name:     &name1,
					Filename: &filename1,
					Size:     &size1,
					Checksum: &checksum1,
				}.Build(),
			})
			return resp, nil
		},
		fileDeleteFunc: func(ctx context.Context, in *api.IDRequest, opts ...grpc.CallOption) (*pbFile.Response, error) {
			return &pbFile.Response{}, nil
		},
	}

	app := NewApp(mockClient)
	app.session = &Session{Token: "test-token", UserID: "uid-1", Login: "user1"}

	t.Run("load files", func(t *testing.T) {
		app.openFileList()
		assert.Len(t, app.currentFiles, 1)
		assert.Equal(t, "Скан паспорта", app.currentFiles[0].GetName())
		assert.Equal(t, "passport.pdf", app.currentFiles[0].GetFilename())
		assert.Equal(t, int64(1024), app.currentFiles[0].GetSize())
	})

	t.Run("open file detail view", func(t *testing.T) {
		app.openFileView(app.currentFiles[0])
		assert.Equal(t, "Скан паспорта", app.currentFile.GetName())
		assert.True(t, app.pages.HasPage("file_view"))
	})

	t.Run("open file download pre-fills filename", func(t *testing.T) {
		app.openFileDownload(app.currentFiles[0])
		assert.Equal(t, "passport.pdf", app.downloadSavePathField.GetText())
		assert.True(t, app.pages.HasPage("file_download"))
	})

	t.Run("delete file success", func(t *testing.T) {
		app.currentFile = app.currentFiles[0]
		app.handleFileDelete()
		assert.Nil(t, app.currentFile)
	})
}

func TestApp_FileUploadStreaming(t *testing.T) {
	// Создаем временный файл на диске для тестирования загрузки
	tmpDir := t.TempDir()
	testFilePath := filepath.Join(tmpDir, "upload_sample.txt")
	testData := []byte("hello world this is a test upload file content stream")
	err := os.WriteFile(testFilePath, testData, 0644)
	require.NoError(t, err)

	hasher := sha256.New()
	hasher.Write(testData)
	expectedChecksum := hex.EncodeToString(hasher.Sum(nil))

	var streamClient *mockUploadStreamClient

	mockClient := &mockUsersClient{
		fileUploadFunc: func(ctx context.Context, opts ...grpc.CallOption) (grpc.ClientStreamingClient[pbFile.UploadFileRequest, pbFile.UploadFileResponse], error) {
			fId := "new-file-id"
			fName := "Тестовый файл"
			fFilename := "upload_sample.txt"
			fSize := int64(len(testData))

			streamClient = &mockUploadStreamClient{
				closeResp: &pbFile.UploadFileResponse{},
			}
			streamClient.closeResp.SetFile(pbFile.File_builder{
				Id:       &fId,
				Name:     &fName,
				Filename: &fFilename,
				Size:     &fSize,
				Checksum: &expectedChecksum,
			}.Build())
			return streamClient, nil
		},
		fileListFunc: func(ctx context.Context, in *api.ListRequest, opts ...grpc.CallOption) (*pbFile.ListResponse, error) {
			return &pbFile.ListResponse{}, nil
		},
	}

	app := NewApp(mockClient)
	app.session = &Session{Token: "test-token", UserID: "uid-1", Login: "user1"}

	t.Run("validation empty fields", func(t *testing.T) {
		app.uploadFileNameField.SetText("")
		app.uploadFilePathField.SetText("")
		app.handleFileUpload()
		assert.True(t, app.pages.HasPage("error_modal"))
		app.pages.RemovePage("error_modal")
	})

	t.Run("validation nonexistent file", func(t *testing.T) {
		app.uploadFileNameField.SetText("Test")
		app.uploadFilePathField.SetText("/nonexistent/file/path.dat")
		app.handleFileUpload()
		assert.True(t, app.pages.HasPage("error_modal"))
		app.pages.RemovePage("error_modal")
	})

	t.Run("successful streaming upload with metadata and chunks", func(t *testing.T) {
		app.uploadFileNameField.SetText("Тестовый файл")
		app.uploadFilePathField.SetText(testFilePath)
		app.handleFileUpload()

		assert.False(t, app.pages.HasPage("error_modal"))
		require.NotNil(t, streamClient)
		require.GreaterOrEqual(t, len(streamClient.sentReqs), 2) // 1 metadata + >=1 chunk

		// Проверяем метаданные в первом сообщении
		firstReq := streamClient.sentReqs[0]
		require.NotNil(t, firstReq.GetMetadata())
		assert.Equal(t, "Тестовый файл", firstReq.GetMetadata().GetName())
		assert.Equal(t, "upload_sample.txt", firstReq.GetMetadata().GetFilename())
		assert.Equal(t, int64(len(testData)), firstReq.GetMetadata().GetSize())
		assert.Equal(t, expectedChecksum, firstReq.GetMetadata().GetChecksum())

		// Проверяем чанк во втором сообщении
		secondReq := streamClient.sentReqs[1]
		assert.Equal(t, testData, secondReq.GetChunk())
	})
}

func TestApp_FileDownloadStreaming(t *testing.T) {
	tmpDir := t.TempDir()
	downloadSavePath := filepath.Join(tmpDir, "saved_sample.txt")
	testData := []byte("downloaded binary stream chunk payload")

	hasher := sha256.New()
	hasher.Write(testData)
	expectedChecksum := hex.EncodeToString(hasher.Sum(nil))

	fName := "Скачиваемый файл"
	fFilename := "saved_sample.txt"
	fSize := int64(len(testData))

	metaMsg := &pbFile.DownloadFileResponse{}
	metaMsg.SetMetadata(pbFile.FileMetadata_builder{
		Name:     &fName,
		Filename: &fFilename,
		Size:     &fSize,
		Checksum: &expectedChecksum,
	}.Build())

	chunkMsg := &pbFile.DownloadFileResponse{}
	chunkMsg.SetChunk(testData)

	mockClient := &mockUsersClient{
		fileDownloadFunc: func(ctx context.Context, in *pbFile.DownloadFileRequest, opts ...grpc.CallOption) (grpc.ServerStreamingClient[pbFile.DownloadFileResponse], error) {
			assert.Equal(t, "f123", in.GetId())
			return &mockDownloadStreamClient{
				msgs: []*pbFile.DownloadFileResponse{metaMsg, chunkMsg},
			}, nil
		},
	}

	app := NewApp(mockClient)
	app.session = &Session{Token: "test-token", UserID: "uid-1", Login: "user1"}
	fId := "f123"
	app.currentFile = pbFile.File_builder{
		Id:       &fId,
		Name:     &fName,
		Filename: &fFilename,
		Size:     &fSize,
		Checksum: &expectedChecksum,
	}.Build()

	t.Run("empty save path validation", func(t *testing.T) {
		app.downloadSavePathField.SetText("")
		app.handleFileDownload()
		assert.True(t, app.pages.HasPage("error_modal"))
		app.pages.RemovePage("error_modal")
	})

	t.Run("successful streaming download and disk verification", func(t *testing.T) {
		app.downloadSavePathField.SetText(downloadSavePath)
		app.handleFileDownload()

		assert.False(t, app.pages.HasPage("error_modal"))
		assert.True(t, app.pages.HasPage("info_modal"))

		// Проверяем, что файл действительно записан на диск и его содержимое совпадает
		savedContent, err := os.ReadFile(downloadSavePath)
		require.NoError(t, err)
		assert.Equal(t, testData, savedContent)
	})
}

func TestApp_FilePagination(t *testing.T) {
	requestedPages := make([]string, 0)
	mockClient := &mockUsersClient{
		fileListFunc: func(ctx context.Context, in *api.ListRequest, opts ...grpc.CallOption) (*pbFile.ListResponse, error) {
			page := in.GetPage()
			requestedPages = append(requestedPages, page)

			resp := &pbFile.ListResponse{}
			if page == "1" || page == "" {
				fName := "File_Page1"
				fFilename := "file1.txt"
				resp.SetFiles([]*pbFile.File{
					pbFile.File_builder{Name: &fName, Filename: &fFilename}.Build(),
				})
			} else if page == "2" {
				fName := "File_Page2"
				fFilename := "file2.txt"
				resp.SetFiles([]*pbFile.File{
					pbFile.File_builder{Name: &fName, Filename: &fFilename}.Build(),
				})
			} else {
				resp.SetFiles([]*pbFile.File{})
			}
			return resp, nil
		},
	}

	app := NewApp(mockClient)
	app.session = &Session{Token: "test-token", UserID: "uid-1", Login: "user1"}

	t.Run("initial load requests page 1", func(t *testing.T) {
		requestedPages = nil
		app.loadFiles()
		assert.Equal(t, 1, app.GetCurrentFilePage())
		require.Len(t, requestedPages, 1)
		assert.Equal(t, "1", requestedPages[0])
		assert.Len(t, app.currentFiles, 1)
		assert.Equal(t, "File_Page1", app.currentFiles[0].GetName())
	})

	t.Run("prev button on page 1 does not decrement", func(t *testing.T) {
		requestedPages = nil
		app.prevFilePage()
		assert.Equal(t, 1, app.GetCurrentFilePage())
		assert.Empty(t, requestedPages)
	})

	t.Run("next button increments page and loads page 2", func(t *testing.T) {
		requestedPages = nil
		app.nextFilePage()
		assert.Equal(t, 2, app.GetCurrentFilePage())
		require.Len(t, requestedPages, 1)
		assert.Equal(t, "2", requestedPages[0])
		assert.Len(t, app.currentFiles, 1)
		assert.Equal(t, "File_Page2", app.currentFiles[0].GetName())
	})

	t.Run("prev button on page 2 decrements to page 1", func(t *testing.T) {
		requestedPages = nil
		app.prevFilePage()
		assert.Equal(t, 1, app.GetCurrentFilePage())
		require.Len(t, requestedPages, 1)
		assert.Equal(t, "1", requestedPages[0])
		assert.Len(t, app.currentFiles, 1)
		assert.Equal(t, "File_Page1", app.currentFiles[0].GetName())
	})

	t.Run("openFileList resets to page 1", func(t *testing.T) {
		app.SetCurrentFilePage(5)
		requestedPages = nil
		app.openFileList()
		assert.Equal(t, 1, app.GetCurrentFilePage())
		require.Len(t, requestedPages, 1)
		assert.Equal(t, "1", requestedPages[0])
	})
}
