package client

import (
	"context"
	"testing"

	"github.com/rivo/tview"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/metadata"
)

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

func TestApp_SetVersion(t *testing.T) {
	app := NewApp(&mockUsersClient{})
	assert.Equal(t, "", app.GetVersion())

	app.SetVersion("v1.2.3")
	assert.Equal(t, "v1.2.3", app.GetVersion())

	app.session = &Session{Token: "token", UserID: "uid", Login: "testuser"}
	app.SetVersion("v2.0.0")
	assert.Equal(t, "v2.0.0", app.GetVersion())
}

func TestApp_AuthCtx(t *testing.T) {
	app := NewApp(&mockUsersClient{})

	t.Run("negative - Контекст без метаданных", func(t *testing.T) {
		ctx, cancel := app.authCtx(context.Background())
		defer cancel()

		md, ok := metadata.FromOutgoingContext(ctx)
		assert.False(t, ok)
		assert.Nil(t, md)
	})

	t.Run("positive - контекст с данными пользователя", func(t *testing.T) {
		app.session = &Session{Token: "test-jwt-token", UserID: "uid-1", Login: "user1"}
		ctx, cancel := app.authCtx(context.Background())
		defer cancel()

		md, ok := metadata.FromOutgoingContext(ctx)
		require.True(t, ok)
		assert.Equal(t, []string{"Bearer test-jwt-token"}, md.Get("authorization"))
		assert.Equal(t, []string{"test-jwt-token"}, md.Get("token"))
	})
}

func TestApp_Modals(t *testing.T) {
	app := NewApp(&mockUsersClient{})

	t.Run("negative - отображение и закрытие модального окна ошибки", func(t *testing.T) {
		app.showErrorModal("Something went wrong", "login")
		assert.True(t, app.pages.HasPage("error_modal"))

		app.pages.RemovePage("error_modal")
		assert.False(t, app.pages.HasPage("error_modal"))
	})

	t.Run("positive - отображение и закрытие информационного модального окна", func(t *testing.T) {
		app.showInfoModal("Operation succeeded", "file_list")
		assert.True(t, app.pages.HasPage("info_modal"))

		app.pages.RemovePage("info_modal")
		assert.False(t, app.pages.HasPage("info_modal"))
	})
}

func TestApp_ExtractTokenFromHeader(t *testing.T) {
	t.Run("negative - извлечение JWT токена из пустого набора метаданных", func(t *testing.T) {
		assert.Equal(t, "", extractTokenFromHeader(nil, nil))
	})

	t.Run("positive - извлечение JWT токена из метаданных", func(t *testing.T) {
		md := metadata.Pairs("token", "my-token-123")
		assert.Equal(t, "my-token-123", extractTokenFromHeader(md))
	})

	t.Run("positive - извлечение JWT токена из заголовков", func(t *testing.T) {
		header := metadata.Pairs()
		trailer := metadata.Pairs("jwt", "jwt-token-789")
		assert.Equal(t, "jwt-token-789", extractTokenFromHeader(header, trailer))
	})

	t.Run("positive - извлечение токена аутентификации пользователя из метеданных", func(t *testing.T) {
		md := metadata.Pairs("authorization", "Bearer auth-token-456")
		assert.Equal(t, "auth-token-456", extractTokenFromHeader(md))
	})

}

func TestApp_CenterPrimitive(t *testing.T) {
	tv := tview.NewTextView().SetText("Hello")
	centered := centerPrimitive(tv, 40, 20)
	require.NotNil(t, centered)
}
