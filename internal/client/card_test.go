package client

import (
	"context"
	"path/filepath"
	"testing"

	api "github.com/alexsey-popov/gkeep/proto"
	pbCard "github.com/alexsey-popov/gkeep/proto/card"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestApp_Cards(t *testing.T) {
	id1 := "c1"
	name1 := "Депозитарий солёных огурцов"
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
			name1 = name
			num1 = num
			exp1 = exp
			code1 = code
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

	t.Run("positive - загрузка списка карт", func(t *testing.T) {
		app.loadCards()
		assert.Len(t, app.currentCards, 1)
		assert.Equal(t, "Депозитарий солёных огурцов", app.currentCards[0].GetName())
	})

	t.Run("negative - создание карты с пустым названием", func(t *testing.T) {
		app.createCardNameField.SetText("")
		app.handleCardCreate()
		assert.True(t, app.pages.HasPage("error_modal"))
		app.pages.RemovePage("error_modal")
	})

	t.Run("positive - успешное сохранение карты", func(t *testing.T) {
		app.createCardNameField.SetText("Депозитарий солёных огурцов")
		app.createCardNumberField.SetText("5555666677778888")
		app.createCardExpiredField.SetText("1130")
		app.createCardCodeField.SetText("999")
		app.handleCardCreate()
		assert.False(t, app.pages.HasPage("error_modal"))
	})

	t.Run("positive - открытие и редактирование карты", func(t *testing.T) {
		app.openCardView(app.currentCards[0])
		assert.Equal(t, "Депозитарий солёных огурцов", app.currentCard.GetName())
		assert.True(t, app.pages.HasPage("card_view"))

		app.openCardEdit(app.currentCards[0])
		assert.Equal(t, "Депозитарий солёных огурцов", app.editCardNameField.GetText())
		app.editCardNameField.SetText("Карта для огурцов")
		app.editCardExpiredField.SetText("0130")
		app.handleCardUpdate()
		assert.Equal(t, "0130", app.currentCard.GetExpired())
		assert.Equal(t, "Карта для огурцов", app.currentCard.GetName())
		assert.Equal(t, "Карта для огурцов", app.currentCards[0].GetName())
	})

	t.Run("positive - удаление карты", func(t *testing.T) {
		app.currentCard = app.currentCards[0]
		app.handleCardDelete()
		assert.Nil(t, app.currentCard)
	})

	t.Run("negative - запрос списка карт неаутентифицированным пользователем", func(t *testing.T) {
		unauthApp := NewApp(mockClient)
		unauthApp.loadCards()
		assert.True(t, unauthApp.pages.HasPage("error_modal"))
	})
}

func TestApp_CardPagination(t *testing.T) {
	requestedPages := make([]string, 0)
	mockClient := &mockUsersClient{
		cardListFunc: func(ctx context.Context, in *api.ListRequest, opts ...grpc.CallOption) (*pbCard.ListResponse, error) {
			page := in.GetPage()
			requestedPages = append(requestedPages, page)

			resp := &pbCard.ListResponse{}
			if page == "1" || page == "" {
				name := "Card_Page1"
				resp.SetCard([]*pbCard.Card{
					pbCard.Card_builder{Name: &name}.Build(),
				})
			} else if page == "2" {
				name := "Card_Page2"
				resp.SetCard([]*pbCard.Card{
					pbCard.Card_builder{Name: &name}.Build(),
				})
			} else {
				resp.SetCard([]*pbCard.Card{})
			}
			return resp, nil
		},
	}

	app := NewApp(mockClient)
	app.session = &Session{Token: "test-token", UserID: "uid-1", Login: "user1"}

	t.Run("positive - проверка пагинации", func(t *testing.T) {
		requestedPages = nil
		app.loadCards()
		assert.Equal(t, 1, app.GetCurrentCardPage())
		require.Len(t, requestedPages, 1)
		assert.Equal(t, "1", requestedPages[0])
		assert.Len(t, app.currentCards, 1)
		assert.Equal(t, "Card_Page1", app.currentCards[0].GetName())
	})

	t.Run("negative - проверка пагинации. Кнопка предыдущей страницы должна быть заблокировано", func(t *testing.T) {
		requestedPages = nil
		app.prevCardPage()
		assert.Equal(t, 1, app.GetCurrentCardPage())
		assert.Empty(t, requestedPages)
	})

	t.Run("positive - проверка пагинации. Переход на следующую страницу", func(t *testing.T) {
		requestedPages = nil
		app.nextCardPage()
		assert.Equal(t, 2, app.GetCurrentCardPage())
		require.Len(t, requestedPages, 1)
		assert.Equal(t, "2", requestedPages[0])
		assert.Len(t, app.currentCards, 1)
		assert.Equal(t, "Card_Page2", app.currentCards[0].GetName())
	})

	t.Run("positive - проверка пагинации. Переход на предыдущую страницу", func(t *testing.T) {
		requestedPages = nil
		app.prevCardPage()
		assert.Equal(t, 1, app.GetCurrentCardPage())
		require.Len(t, requestedPages, 1)
		assert.Equal(t, "1", requestedPages[0])
		assert.Len(t, app.currentCards, 1)
		assert.Equal(t, "Card_Page1", app.currentCards[0].GetName())
	})

	t.Run("positive - проверка пагинации. Сброс счётчика страницы при открытии страницы", func(t *testing.T) {
		app.SetCurrentCardPage(5)
		requestedPages = nil
		app.openCardList()
		assert.Equal(t, 1, app.GetCurrentCardPage())
		require.Len(t, requestedPages, 1)
		assert.Equal(t, "1", requestedPages[0])
	})

	t.Run("negative - ошибка при загрузке списка карт", func(t *testing.T) {
		errMockClient := &mockUsersClient{
			cardListFunc: func(ctx context.Context, in *api.ListRequest, opts ...grpc.CallOption) (*pbCard.ListResponse, error) {
				return nil, status.Error(codes.Internal, "database error")
			},
		}
		errApp := NewApp(errMockClient)
		errApp.SetCacheFilePath(filepath.Join(t.TempDir(), "nonexistent.cache"))
		errApp.session = &Session{Token: "test-token", UserID: "uid-1", Login: "user1"}
		errApp.loadCards()
		assert.True(t, errApp.pages.HasPage("error_modal"))
	})
}
