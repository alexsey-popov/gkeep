package client

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	api "github.com/alexsey-popov/gkeep/proto"
	pbCard "github.com/alexsey-popov/gkeep/proto/card"
	pbFile "github.com/alexsey-popov/gkeep/proto/file"
	pbPassword "github.com/alexsey-popov/gkeep/proto/password"
	pbText "github.com/alexsey-popov/gkeep/proto/text"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

func TestCache_OfflinePasswords(t *testing.T) {
	tmpDir := t.TempDir()
	cachePath := filepath.Join(tmpDir, "cache.enc")

	id := "p1"
	name := "Google Account"
	login := "user@gmail.com"
	pass := "secret123"
	item := pbPassword.Password_builder{
		Id:       &id,
		Name:     &name,
		Login:    &login,
		Password: &pass,
	}.Build()

	serverAvailable := true
	mockClient := &mockUsersClient{
		passwordListFunc: func(ctx context.Context, in *api.ListRequest, opts ...grpc.CallOption) (*pbPassword.ListResponse, error) {
			if !serverAvailable {
				return nil, errors.New("connection refused")
			}
			resp := &pbPassword.ListResponse{}
			resp.SetPassword([]*pbPassword.Password{item})
			return resp, nil
		},
	}

	app := NewApp(mockClient)
	app.SetCacheFilePath(cachePath)
	app.session = &Session{UserID: "u1", Login: "user1", Token: "jwt"}

	// 1. Успешный запрос к серверу - данные должны сохраниться в кэш
	app.loadPasswords()
	assert.False(t, app.isPasswordsCached)
	assert.Len(t, app.currentPasswords, 1)
	assert.Equal(t, "Google Account", app.currentPasswords[0].GetName())
	assert.NotContains(t, app.getPasswordsListTitle(), "Локальная копия данных")

	// 2. Сервер падает
	serverAvailable = false

	// Сбрасываем текущие данные
	app.currentPasswords = nil

	// Загружаем список паролей при недоступном сервере
	app.loadPasswords()
	assert.True(t, app.isPasswordsCached)
	require.Len(t, app.currentPasswords, 1)
	assert.Equal(t, "Google Account", app.currentPasswords[0].GetName())
	assert.Contains(t, app.getPasswordsListTitle(), "Локальная копия данных")

	// Проверяем форму просмотра
	app.openPasswordView(app.currentPasswords[0])
	assert.Contains(t, app.passwordDetailsText.GetText(false), "Локальная копия данных")
}

func TestCache_OfflineCards(t *testing.T) {
	tmpDir := t.TempDir()
	cachePath := filepath.Join(tmpDir, "cache.enc")

	id := "c1"
	name := "Tinkoff Black"
	number := "2200111122223333"
	expired := "12/28"
	code := "123"
	item := pbCard.Card_builder{
		Id:      &id,
		Name:    &name,
		Number:  &number,
		Expired: &expired,
		Code:    &code,
	}.Build()

	serverAvailable := true
	mockClient := &mockUsersClient{
		cardListFunc: func(ctx context.Context, in *api.ListRequest, opts ...grpc.CallOption) (*pbCard.ListResponse, error) {
			if !serverAvailable {
				return nil, errors.New("connection error")
			}
			resp := &pbCard.ListResponse{}
			resp.SetCard([]*pbCard.Card{item})
			return resp, nil
		},
	}

	app := NewApp(mockClient)
	app.SetCacheFilePath(cachePath)
	app.session = &Session{UserID: "u1", Login: "user1", Token: "jwt"}

	// 1. Успешная загрузка с сохранением в кэш
	app.loadCards()
	assert.False(t, app.isCardsCached)
	assert.Len(t, app.currentCards, 1)
	assert.Equal(t, "Tinkoff Black", app.currentCards[0].GetName())
	assert.NotContains(t, app.getCardsListTitle(), "Локальная копия данных")

	// 2. Сервер недоступен
	serverAvailable = false
	app.currentCards = nil

	app.loadCards()
	assert.True(t, app.isCardsCached)
	require.Len(t, app.currentCards, 1)
	assert.Equal(t, "Tinkoff Black", app.currentCards[0].GetName())
	assert.Contains(t, app.getCardsListTitle(), "Локальная копия данных")

	// Проверяем форму просмотра
	app.openCardView(app.currentCards[0])
	assert.Contains(t, app.cardDetailsText.GetText(false), "Локальная копия данных")
}

func TestCache_OfflineTexts(t *testing.T) {
	tmpDir := t.TempDir()
	cachePath := filepath.Join(tmpDir, "cache.enc")

	id := "t1"
	name := "Secret Note"
	textVal := "My top secret notes"
	item := pbText.Text_builder{
		Id:   &id,
		Name: &name,
		Text: &textVal,
	}.Build()

	serverAvailable := true
	mockClient := &mockUsersClient{
		textListFunc: func(ctx context.Context, in *api.ListRequest, opts ...grpc.CallOption) (*pbText.ListResponse, error) {
			if !serverAvailable {
				return nil, errors.New("network error")
			}
			resp := &pbText.ListResponse{}
			resp.SetText([]*pbText.Text{item})
			return resp, nil
		},
	}

	app := NewApp(mockClient)
	app.SetCacheFilePath(cachePath)
	app.session = &Session{UserID: "u1", Login: "user1", Token: "jwt"}

	// 1. Успешная загрузка
	app.loadTexts()
	assert.False(t, app.isTextsCached)
	assert.Len(t, app.currentTexts, 1)
	assert.Equal(t, "Secret Note", app.currentTexts[0].GetName())
	assert.NotContains(t, app.getTextsListTitle(), "Локальная копия данных")

	// 2. Сервер недоступен
	serverAvailable = false
	app.currentTexts = nil

	app.loadTexts()
	assert.True(t, app.isTextsCached)
	require.Len(t, app.currentTexts, 1)
	assert.Equal(t, "Secret Note", app.currentTexts[0].GetName())
	assert.Contains(t, app.getTextsListTitle(), "Локальная копия данных")

	// Проверяем форму просмотра текста
	app.openTextView(app.currentTexts[0])
	assert.Contains(t, app.textDetailsText.GetText(false), "Локальная копия данных")
	assert.Contains(t, app.textDetailsText.GetText(false), "Secret Note")
	assert.Contains(t, app.textDetailsText.GetText(false), "My top secret notes")
}

func TestCache_OfflineFiles(t *testing.T) {
	tmpDir := t.TempDir()
	cachePath := filepath.Join(tmpDir, "cache.enc")

	id := "f1"
	name := "Passport Scan"
	filename := "passport.pdf"
	size := int64(1048576)
	checksum := "abcdef123456"
	item := pbFile.File_builder{
		Id:       &id,
		Name:     &name,
		Filename: &filename,
		Size:     &size,
		Checksum: &checksum,
	}.Build()

	serverAvailable := true
	mockClient := &mockUsersClient{
		fileListFunc: func(ctx context.Context, in *api.ListRequest, opts ...grpc.CallOption) (*pbFile.ListResponse, error) {
			if !serverAvailable {
				return nil, errors.New("network timeout")
			}
			resp := &pbFile.ListResponse{}
			resp.SetFiles([]*pbFile.File{item})
			return resp, nil
		},
	}

	app := NewApp(mockClient)
	app.SetCacheFilePath(cachePath)
	app.session = &Session{UserID: "u1", Login: "user1", Token: "jwt"}

	// 1. Успешная загрузка
	app.loadFiles()
	assert.False(t, app.isFilesCached)
	assert.Len(t, app.currentFiles, 1)
	assert.Equal(t, "Passport Scan", app.currentFiles[0].GetName())
	assert.NotContains(t, app.getFilesListTitle(), "Локальная копия данных")

	// 2. Сервер недоступен
	serverAvailable = false
	app.currentFiles = nil

	app.loadFiles()
	assert.True(t, app.isFilesCached)
	require.Len(t, app.currentFiles, 1)
	assert.Equal(t, "Passport Scan", app.currentFiles[0].GetName())
	assert.Contains(t, app.getFilesListTitle(), "Локальная копия данных")

	// Проверяем форму просмотра файла
	app.openFileView(app.currentFiles[0])
	assert.Contains(t, app.fileDetailsText.GetText(false), "Локальная копия данных")
	assert.Contains(t, app.fileDetailsText.GetText(false), "Passport Scan")
}

func TestCache_DifferentUsersIsolation(t *testing.T) {
	tmpDir := t.TempDir()
	cachePath := filepath.Join(tmpDir, "shared_cache.enc")

	id := "p1"
	name := "User1 Secret"
	login := "u1"
	pass := "pass1"
	item := pbPassword.Password_builder{
		Id:       &id,
		Name:     &name,
		Login:    &login,
		Password: &pass,
	}.Build()

	mockClient := &mockUsersClient{
		passwordListFunc: func(ctx context.Context, in *api.ListRequest, opts ...grpc.CallOption) (*pbPassword.ListResponse, error) {
			resp := &pbPassword.ListResponse{}
			resp.SetPassword([]*pbPassword.Password{item})
			return resp, nil
		},
	}

	// Пользователь 1 загружает данные
	app1 := NewApp(mockClient)
	app1.SetCacheFilePath(cachePath)
	app1.session = &Session{UserID: "user_id_1", Login: "alice", Token: "jwt1"}
	app1.loadPasswords()

	// Пользователь 2 пытается прочитать кэш пользователя 1 в офлайн режиме
	mockClientErr := &mockUsersClient{
		passwordListFunc: func(ctx context.Context, in *api.ListRequest, opts ...grpc.CallOption) (*pbPassword.ListResponse, error) {
			return nil, errors.New("server down")
		},
	}
	app2 := NewApp(mockClientErr)
	app2.SetCacheFilePath(cachePath)
	app2.session = &Session{UserID: "user_id_2", Login: "bob", Token: "jwt2"}

	app2.loadPasswords()
	// Из-за другого ключа шифрования (другой userID/login) кэш пользователя 1 не прочитается
	assert.Empty(t, app2.currentPasswords)
}

func TestCache_RefreshNotifications(t *testing.T) {
	tmpDir := t.TempDir()
	cachePath := filepath.Join(tmpDir, "refresh_test.cache")

	passId := "p1"
	passName := "Pass1"
	passItem := pbPassword.Password_builder{Id: &passId, Name: &passName}.Build()

	cardId := "c1"
	cardName := "Card1"
	cardItem := pbCard.Card_builder{Id: &cardId, Name: &cardName}.Build()

	textId := "t1"
	textName := "Text1"
	textItem := pbText.Text_builder{Id: &textId, Name: &textName}.Build()

	fileId := "f1"
	fileName := "File1"
	fileItem := pbFile.File_builder{Id: &fileId, Name: &fileName}.Build()

	serverAvailable := true
	mockClient := &mockUsersClient{
		passwordListFunc: func(ctx context.Context, in *api.ListRequest, opts ...grpc.CallOption) (*pbPassword.ListResponse, error) {
			if !serverAvailable {
				return nil, errors.New("server down")
			}
			resp := &pbPassword.ListResponse{}
			resp.SetPassword([]*pbPassword.Password{passItem})
			return resp, nil
		},
		cardListFunc: func(ctx context.Context, in *api.ListRequest, opts ...grpc.CallOption) (*pbCard.ListResponse, error) {
			if !serverAvailable {
				return nil, errors.New("server down")
			}
			resp := &pbCard.ListResponse{}
			resp.SetCard([]*pbCard.Card{cardItem})
			return resp, nil
		},
		textListFunc: func(ctx context.Context, in *api.ListRequest, opts ...grpc.CallOption) (*pbText.ListResponse, error) {
			if !serverAvailable {
				return nil, errors.New("server down")
			}
			resp := &pbText.ListResponse{}
			resp.SetText([]*pbText.Text{textItem})
			return resp, nil
		},
		fileListFunc: func(ctx context.Context, in *api.ListRequest, opts ...grpc.CallOption) (*pbFile.ListResponse, error) {
			if !serverAvailable {
				return nil, errors.New("server down")
			}
			resp := &pbFile.ListResponse{}
			resp.SetFiles([]*pbFile.File{fileItem})
			return resp, nil
		},
	}

	app := NewApp(mockClient)
	app.SetCacheFilePath(cachePath)
	app.session = &Session{UserID: "u1", Login: "user1", Token: "jwt"}

	t.Run("online refresh - пароли, карты, тексты, файлы", func(t *testing.T) {
		serverAvailable = true

		// Пароли
		app.handleRefreshPasswords()
		assert.False(t, app.isPasswordsCached)
		assert.True(t, app.pages.HasPage("info_modal"))
		app.pages.RemovePage("info_modal")

		// Карты
		app.handleRefreshCards()
		assert.False(t, app.isCardsCached)
		assert.True(t, app.pages.HasPage("info_modal"))
		app.pages.RemovePage("info_modal")

		// Тексты
		app.handleRefreshTexts()
		assert.False(t, app.isTextsCached)
		assert.True(t, app.pages.HasPage("info_modal"))
		app.pages.RemovePage("info_modal")

		// Файлы
		app.handleRefreshFiles()
		assert.False(t, app.isFilesCached)
		assert.True(t, app.pages.HasPage("info_modal"))
		app.pages.RemovePage("info_modal")
	})

	t.Run("cached refresh - пароли, карты, тексты, файлы", func(t *testing.T) {
		serverAvailable = false

		// Пароли из кэша
		app.handleRefreshPasswords()
		assert.True(t, app.isPasswordsCached)
		assert.True(t, app.pages.HasPage("info_modal"))
		app.pages.RemovePage("info_modal")

		// Карты из кэша
		app.handleRefreshCards()
		assert.True(t, app.isCardsCached)
		assert.True(t, app.pages.HasPage("info_modal"))
		app.pages.RemovePage("info_modal")

		// Тексты из кэша
		app.handleRefreshTexts()
		assert.True(t, app.isTextsCached)
		assert.True(t, app.pages.HasPage("info_modal"))
		app.pages.RemovePage("info_modal")

		// Файлы из кэша
		app.handleRefreshFiles()
		assert.True(t, app.isFilesCached)
		assert.True(t, app.pages.HasPage("info_modal"))
		app.pages.RemovePage("info_modal")
	})
}
