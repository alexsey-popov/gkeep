package client

import (
	"context"
	"testing"

	api "github.com/alexsey-popov/gkeep/proto"
	pbPassword "github.com/alexsey-popov/gkeep/proto/password"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

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

	t.Run("positive - загрузка паролей", func(t *testing.T) {
		app.loadPasswords()
		assert.Len(t, app.currentPasswords, 1)
		assert.Equal(t, "Email", app.currentPasswords[0].GetName())
	})

	t.Run("negative - пустое поле с названием", func(t *testing.T) {
		app.createPassNameField.SetText("")
		app.handlePasswordCreate()
		assert.True(t, app.pages.HasPage("error_modal"))
		app.pages.RemovePage("error_modal")
	})

	t.Run("positive - сохранение пароля", func(t *testing.T) {
		app.createPassNameField.SetText("GitHub")
		app.createPassLoginField.SetText("gituser")
		app.createPassValField.SetText("gitpass")
		app.handlePasswordCreate()
		assert.False(t, app.pages.HasPage("error_modal"))
	})

	t.Run("positive - форма просмотра и редактирование", func(t *testing.T) {
		app.openPasswordView(app.currentPasswords[0])
		assert.Equal(t, "Email", app.currentPassword.GetName())
		assert.True(t, app.pages.HasPage("password_view"))

		app.openPasswordEdit(app.currentPasswords[0])
		assert.Equal(t, "Email", app.editPassNameField.GetText())
		assert.Equal(t, "user@mail.com", app.editPassLoginField.GetText())

		app.editPassValField.SetText("new-password-456")
		app.handlePasswordUpdate()

		assert.Equal(t, "new-password-456", app.currentPassword.GetPassword())
	})

	t.Run("positive - удаление пароля", func(t *testing.T) {
		app.currentPassword = app.currentPasswords[0]
		app.handlePasswordDelete()
		assert.Nil(t, app.currentPassword)
	})

	t.Run("negative - запрос списка паролей без аутентификации", func(t *testing.T) {
		unauthApp := NewApp(mockClient)
		unauthApp.loadPasswords()
		assert.True(t, unauthApp.pages.HasPage("error_modal"))
	})
}

func TestApp_PasswordPagination(t *testing.T) {
	requestedPages := make([]string, 0)
	mockClient := &mockUsersClient{
		passwordListFunc: func(ctx context.Context, in *api.ListRequest, opts ...grpc.CallOption) (*pbPassword.ListResponse, error) {
			page := in.GetPage()
			requestedPages = append(requestedPages, page)

			resp := &pbPassword.ListResponse{}
			if page == "1" || page == "" {
				name := "Pass_Page1"
				resp.SetPassword([]*pbPassword.Password{
					pbPassword.Password_builder{Name: &name}.Build(),
				})
			} else if page == "2" {
				name := "Pass_Page2"
				resp.SetPassword([]*pbPassword.Password{
					pbPassword.Password_builder{Name: &name}.Build(),
				})
			} else {
				resp.SetPassword([]*pbPassword.Password{})
			}
			return resp, nil
		},
	}

	app := NewApp(mockClient)
	app.session = &Session{Token: "test-token", UserID: "uid-1", Login: "user1"}

	t.Run("positive - проверка пагинации. Первая страница", func(t *testing.T) {
		requestedPages = nil
		app.loadPasswords()
		assert.Equal(t, 1, app.GetCurrentPage())
		require.Len(t, requestedPages, 1)
		assert.Equal(t, "1", requestedPages[0])
		assert.Len(t, app.currentPasswords, 1)
		assert.Equal(t, "Pass_Page1", app.currentPasswords[0].GetName())
	})

	t.Run("positive - проверка пагинации. Предыдущая страница должна быть заблокирована", func(t *testing.T) {
		requestedPages = nil
		app.prevPage()
		assert.Equal(t, 1, app.GetCurrentPage())
		assert.Empty(t, requestedPages)
	})

	t.Run("positive - проверка пагинации. Переход на следующую страницу", func(t *testing.T) {
		requestedPages = nil
		app.nextPage()
		assert.Equal(t, 2, app.GetCurrentPage())
		require.Len(t, requestedPages, 1)
		assert.Equal(t, "2", requestedPages[0])
		assert.Len(t, app.currentPasswords, 1)
		assert.Equal(t, "Pass_Page2", app.currentPasswords[0].GetName())
	})

	t.Run("positive - проверка пагинации. Переход на предыдущую страницу", func(t *testing.T) {
		requestedPages = nil
		app.prevPage()
		assert.Equal(t, 1, app.GetCurrentPage())
		require.Len(t, requestedPages, 1)
		assert.Equal(t, "1", requestedPages[0])
		assert.Len(t, app.currentPasswords, 1)
		assert.Equal(t, "Pass_Page1", app.currentPasswords[0].GetName())
	})

	t.Run("positive - проверка пагинации. Сброс номера страницы при переходе в форму списка паролей", func(t *testing.T) {
		app.SetCurrentPage(5)
		requestedPages = nil
		app.openPasswordList()
		assert.Equal(t, 1, app.GetCurrentPage())
		require.Len(t, requestedPages, 1)
		assert.Equal(t, "1", requestedPages[0])
	})

	t.Run("negative - ошибка при загрузке списка", func(t *testing.T) {
		errMockClient := &mockUsersClient{
			passwordListFunc: func(ctx context.Context, in *api.ListRequest, opts ...grpc.CallOption) (*pbPassword.ListResponse, error) {
				return nil, status.Error(codes.Internal, "internal error")
			},
		}
		errApp := NewApp(errMockClient)
		errApp.session = &Session{Token: "test-token", UserID: "uid-1", Login: "user1"}
		errApp.loadPasswords()
		assert.True(t, errApp.pages.HasPage("error_modal"))
	})
}
