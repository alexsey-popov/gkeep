package client

import (
	"context"
	"path/filepath"
	"testing"

	api "github.com/alexsey-popov/gkeep/proto"
	pbText "github.com/alexsey-popov/gkeep/proto/text"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

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
			name1 = name
			val1 = txt
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

	t.Run("positive - загрузка списка", func(t *testing.T) {
		app.loadTexts()
		assert.Len(t, app.currentTexts, 1)
		assert.Equal(t, "My Secret Note", app.currentTexts[0].GetName())
	})

	t.Run("negative - валидация пустой формы", func(t *testing.T) {
		app.createTextNameField.SetText("")
		app.handleTextCreate()
		assert.True(t, app.pages.HasPage("error_modal"))
		app.pages.RemovePage("error_modal")
	})

	t.Run("positive - создание записи", func(t *testing.T) {
		app.createTextNameField.SetText("New Note")
		app.createTextValField.SetText("Some notes here")
		app.handleTextCreate()
		assert.False(t, app.pages.HasPage("error_modal"))
	})

	t.Run("positive - открытие и редактирование записи", func(t *testing.T) {
		app.openTextView(app.currentTexts[0])
		assert.Equal(t, "My Secret Note", app.currentText.GetName())
		assert.True(t, app.pages.HasPage("text_view"))

		app.openTextEdit(app.currentTexts[0])
		assert.Equal(t, "My Secret Note", app.editTextNameField.GetText())
		app.editTextNameField.SetText("Updated Secret Note")
		app.editTextValField.SetText("Updated secret note content")
		app.handleTextUpdate()
		assert.Equal(t, "Updated secret note content", app.currentText.GetText())
		assert.Equal(t, "Updated Secret Note", app.currentText.GetName())
		assert.Equal(t, "Updated Secret Note", app.currentTexts[0].GetName())
	})

	t.Run("positive - удаление записи", func(t *testing.T) {
		app.currentText = app.currentTexts[0]
		app.handleTextDelete()
		assert.Nil(t, app.currentText)
	})

	t.Run("negative - запрос списка без аутентификации", func(t *testing.T) {
		unauthApp := NewApp(mockClient)
		unauthApp.loadTexts()
		assert.True(t, unauthApp.pages.HasPage("error_modal"))
	})
}

func TestApp_TextPagination(t *testing.T) {
	requestedPages := make([]string, 0)
	mockClient := &mockUsersClient{
		textListFunc: func(ctx context.Context, in *api.ListRequest, opts ...grpc.CallOption) (*pbText.ListResponse, error) {
			page := in.GetPage()
			requestedPages = append(requestedPages, page)

			resp := &pbText.ListResponse{}
			if page == "1" || page == "" {
				name := "Text_Page1"
				resp.SetText([]*pbText.Text{
					pbText.Text_builder{Name: &name}.Build(),
				})
			} else if page == "2" {
				name := "Text_Page2"
				resp.SetText([]*pbText.Text{
					pbText.Text_builder{Name: &name}.Build(),
				})
			} else {
				resp.SetText([]*pbText.Text{})
			}
			return resp, nil
		},
	}

	app := NewApp(mockClient)
	app.session = &Session{Token: "test-token", UserID: "uid-1", Login: "user1"}

	t.Run("positive - проверка пагинации. Проверка 1-й страницы", func(t *testing.T) {
		requestedPages = nil
		app.loadTexts()
		assert.Equal(t, 1, app.GetCurrentTextPage())
		require.Len(t, requestedPages, 1)
		assert.Equal(t, "1", requestedPages[0])
		assert.Len(t, app.currentTexts, 1)
		assert.Equal(t, "Text_Page1", app.currentTexts[0].GetName())
	})

	t.Run("positive - проверка пагинации. На 1-й странице кнопка предыдущей страницы должна быть заблокирована", func(t *testing.T) {
		requestedPages = nil
		app.prevTextPage()
		assert.Equal(t, 1, app.GetCurrentTextPage())
		assert.Empty(t, requestedPages)
	})

	t.Run("positive - проверка пагинации. Переход на следующую страницу", func(t *testing.T) {
		requestedPages = nil
		app.nextTextPage()
		assert.Equal(t, 2, app.GetCurrentTextPage())
		require.Len(t, requestedPages, 1)
		assert.Equal(t, "2", requestedPages[0])
		assert.Len(t, app.currentTexts, 1)
		assert.Equal(t, "Text_Page2", app.currentTexts[0].GetName())
	})

	t.Run("positive - проверка пагинации. Переход на предыдущую страницу", func(t *testing.T) {
		requestedPages = nil
		app.prevTextPage()
		assert.Equal(t, 1, app.GetCurrentTextPage())
		require.Len(t, requestedPages, 1)
		assert.Equal(t, "1", requestedPages[0])
		assert.Len(t, app.currentTexts, 1)
		assert.Equal(t, "Text_Page1", app.currentTexts[0].GetName())
	})

	t.Run("positive - проверка пагинации. Сброс номера страницы при открытии формы списка записей", func(t *testing.T) {
		app.SetCurrentTextPage(5)
		requestedPages = nil
		app.openTextList()
		assert.Equal(t, 1, app.GetCurrentTextPage())
		require.Len(t, requestedPages, 1)
		assert.Equal(t, "1", requestedPages[0])
	})

	t.Run("negative - ошибка при загрузке списка записей", func(t *testing.T) {
		errMockClient := &mockUsersClient{
			textListFunc: func(ctx context.Context, in *api.ListRequest, opts ...grpc.CallOption) (*pbText.ListResponse, error) {
				return nil, status.Error(codes.Internal, "database error")
			},
		}
		errApp := NewApp(errMockClient)
		errApp.SetCacheFilePath(filepath.Join(t.TempDir(), "nonexistent.cache"))
		errApp.session = &Session{Token: "test-token", UserID: "uid-1", Login: "user1"}
		errApp.loadTexts()
		assert.True(t, errApp.pages.HasPage("error_modal"))
	})
}
