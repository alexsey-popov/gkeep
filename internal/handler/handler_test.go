package handler

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net"
	"testing"

	"github.com/alexsey-popov/gkeep/internal/auth"
	"github.com/alexsey-popov/gkeep/internal/crypto"
	"github.com/alexsey-popov/gkeep/internal/validator"
	api "github.com/alexsey-popov/gkeep/proto"
	pbCard "github.com/alexsey-popov/gkeep/proto/card"
	pbPassword "github.com/alexsey-popov/gkeep/proto/password"
	pbText "github.com/alexsey-popov/gkeep/proto/text"
	pbUser "github.com/alexsey-popov/gkeep/proto/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

func TestGRPCServer_IntegrationWithInterceptor(t *testing.T) {
	secret := "secret-jwt-token"
	guard := auth.New(secret)
	rep := newMockRepository()
	cipher, err := crypto.NewWithTempKey()
	require.NoError(t, err)
	val, err := validator.New()
	require.NoError(t, err)
	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	h := New(rep, guard, cipher, val, log)

	buffer := 1024 * 1024
	lis := bufconn.Listen(buffer)

	server := grpc.NewServer(
		grpc.UnaryInterceptor(guard.UnaryServerInterceptor()),
	)
	api.RegisterAPIServer(server, h)

	go func() {
		if err := server.Serve(lis); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			t.Logf("Server error: %v", err)
		}
	}()
	defer server.Stop()

	conn, err := grpc.NewClient("passthrough:///bufnet",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return lis.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)
	defer conn.Close()

	client := api.NewAPIClient(conn)

	var authToken string

	t.Run("1. Регистрация нового пользователя (без токена) - успех + получение токена в header metadata", func(t *testing.T) {
		req := pbUser.Request_builder{
			Login:    stringPtr("john_doe"),
			Password: stringPtr("strong_password"),
		}.Build()

		var header metadata.MD
		resp, err := client.RegisterUser(context.Background(), req, grpc.Header(&header))
		require.NoError(t, err)
		assert.Equal(t, "john_doe", resp.GetUser().GetLogin())

		// Проверяем метаданные в ответе
		tokens := header.Get("token")
		require.NotEmpty(t, tokens, "Заголовок token должен присутствовать в metadata ответа")
		authToken = tokens[0]
		assert.NotEmpty(t, authToken)

		authHeaders := header.Get("authorization")
		require.NotEmpty(t, authHeaders, "Заголовок authorization должен присутствовать в metadata ответа")
		assert.Contains(t, authHeaders[0], authToken)
	})

	t.Run("2. Повторная регистрация под тем же логином (без токена) - ошибка AlreadyExists", func(t *testing.T) {
		req := pbUser.Request_builder{
			Login:    stringPtr("john_doe"),
			Password: stringPtr("strong_password"),
		}.Build()

		_, err := client.RegisterUser(context.Background(), req)
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.AlreadyExists, st.Code())
	})

	t.Run("3. Регистрация авторизованным пользователем - ошибка PermissionDenied", func(t *testing.T) {
		req := pbUser.Request_builder{
			Login:    stringPtr("another_user"),
			Password: stringPtr("strong_password"),
		}.Build()

		ctx := metadata.NewOutgoingContext(context.Background(), metadata.Pairs("authorization", "Bearer "+authToken))
		_, err := client.RegisterUser(ctx, req)
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.PermissionDenied, st.Code())
	})

	t.Run("4. Логин авторизованным пользователем - ошибка PermissionDenied", func(t *testing.T) {
		req := pbUser.Request_builder{
			Login:    stringPtr("john_doe"),
			Password: stringPtr("strong_password"),
		}.Build()

		ctx := metadata.NewOutgoingContext(context.Background(), metadata.Pairs("authorization", "Bearer "+authToken))
		_, err := client.LoginUser(ctx, req)
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.PermissionDenied, st.Code())
	})

	t.Run("5. Успешный логин неавторизованным пользователем - успех + получение токена в metadata", func(t *testing.T) {
		req := pbUser.Request_builder{
			Login:    stringPtr("john_doe"),
			Password: stringPtr("strong_password"),
		}.Build()

		var header metadata.MD
		resp, err := client.LoginUser(context.Background(), req, grpc.Header(&header))
		require.NoError(t, err)
		assert.Equal(t, "john_doe", resp.GetUser().GetLogin())

		tokens := header.Get("token")
		require.NotEmpty(t, tokens)
		assert.NotEmpty(t, tokens[0])
	})

	t.Run("6. Логин с неверным паролем - ошибка Unauthenticated", func(t *testing.T) {
		req := pbUser.Request_builder{
			Login:    stringPtr("john_doe"),
			Password: stringPtr("wrong_password"),
		}.Build()

		_, err := client.LoginUser(context.Background(), req)
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.Unauthenticated, st.Code())
	})

	// === Password CRUD через gRPC клиент ===
	var createdPasswordID string

	t.Run("7. PasswordCreate без токена - ошибка Unauthenticated", func(t *testing.T) {
		req := pbPassword.CreateRequest_builder{
			Name:     stringPtr("Google"),
			Login:    stringPtr("john@gmail.com"),
			Password: stringPtr("mysecret123"),
		}.Build()

		_, err := client.PasswordCreate(context.Background(), req)
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.Unauthenticated, st.Code())
	})

	t.Run("8. PasswordCreate с токеном - успех", func(t *testing.T) {
		req := pbPassword.CreateRequest_builder{
			Name:     stringPtr("Google"),
			Login:    stringPtr("john@gmail.com"),
			Password: stringPtr("mysecret123"),
		}.Build()

		ctx := metadata.NewOutgoingContext(context.Background(), metadata.Pairs("authorization", "Bearer "+authToken))
		resp, err := client.PasswordCreate(ctx, req)
		require.NoError(t, err)
		require.NotNil(t, resp.GetPassword())
		assert.Equal(t, "Google", resp.GetPassword().GetName())
		assert.Equal(t, "john@gmail.com", resp.GetPassword().GetLogin())
		assert.Equal(t, "mysecret123", resp.GetPassword().GetPassword())
		assert.NotEmpty(t, resp.GetPassword().GetId())
		createdPasswordID = resp.GetPassword().GetId()
	})

	t.Run("9. PasswordCreate дубликат имени - ошибка AlreadyExists", func(t *testing.T) {
		req := pbPassword.CreateRequest_builder{
			Name:     stringPtr("Google"),
			Login:    stringPtr("john@gmail.com"),
			Password: stringPtr("otherpass"),
		}.Build()

		ctx := metadata.NewOutgoingContext(context.Background(), metadata.Pairs("authorization", "Bearer "+authToken))
		_, err := client.PasswordCreate(ctx, req)
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.AlreadyExists, st.Code())
	})

	t.Run("10. PasswordList - получение списка паролей", func(t *testing.T) {
		ctx := metadata.NewOutgoingContext(context.Background(), metadata.Pairs("authorization", "Bearer "+authToken))
		req := &api.ListRequest{}

		resp, err := client.PasswordList(ctx, req)
		require.NoError(t, err)
		require.Len(t, resp.GetPassword(), 1)
		assert.Equal(t, "Google", resp.GetPassword()[0].GetName())
	})

	t.Run("11. PasswordList с фильтрацией по имени", func(t *testing.T) {
		ctx := metadata.NewOutgoingContext(context.Background(), metadata.Pairs("authorization", "Bearer "+authToken))

		req2 := pbPassword.CreateRequest_builder{
			Name:     stringPtr("GitHub"),
			Login:    stringPtr("john_git"),
			Password: stringPtr("gitpass"),
		}.Build()
		_, err := client.PasswordCreate(ctx, req2)
		require.NoError(t, err)

		filterName := "goog"
		listReq := api.ListRequest_builder{
			Name: &filterName,
		}.Build()
		resp, err := client.PasswordList(ctx, listReq)
		require.NoError(t, err)
		require.Len(t, resp.GetPassword(), 1)
		assert.Equal(t, "Google", resp.GetPassword()[0].GetName())
	})

	t.Run("12. PasswordUpdate - успешное обновление пароля", func(t *testing.T) {
		ctx := metadata.NewOutgoingContext(context.Background(), metadata.Pairs("authorization", "Bearer "+authToken))
		req := pbPassword.UpdateRequest_builder{
			Id:       &createdPasswordID,
			Name:     stringPtr("Google"),
			Login:    stringPtr("john@gmail.com"),
			Password: stringPtr("newsecret456"),
		}.Build()

		resp, err := client.PasswordUpdate(ctx, req)
		require.NoError(t, err)
		require.NotNil(t, resp.GetPassword())
		assert.Equal(t, "newsecret456", resp.GetPassword().GetPassword())
	})

	t.Run("13. PasswordDelete - успешное удаление записи", func(t *testing.T) {
		ctx := metadata.NewOutgoingContext(context.Background(), metadata.Pairs("authorization", "Bearer "+authToken))
		req := api.IDRequest_builder{
			Id: &createdPasswordID,
		}.Build()

		resp, err := client.PasswordDelete(ctx, req)
		require.NoError(t, err)
		require.NotNil(t, resp.GetPassword())
		assert.Equal(t, createdPasswordID, resp.GetPassword().GetId())

		filterName := "Google"
		listReq := api.ListRequest_builder{
			Name: &filterName,
		}.Build()
		listResp, err := client.PasswordList(ctx, listReq)
		require.NoError(t, err)
		assert.Empty(t, listResp.GetPassword())
	})

	// === Card CRUD через gRPC клиент ===
	var createdCardID string
	validCardNum := "4532015112830366"

	t.Run("14. CardCreate с токеном - успех", func(t *testing.T) {
		req := pbCard.CreateRequest_builder{
			Name:    stringPtr("Моя карта"),
			Number:  stringPtr(validCardNum),
			Expired: stringPtr("1228"),
			Code:    stringPtr("123"),
		}.Build()

		ctx := metadata.NewOutgoingContext(context.Background(), metadata.Pairs("authorization", "Bearer "+authToken))
		resp, err := client.CardCreate(ctx, req)
		require.NoError(t, err)
		require.NotNil(t, resp.GetCard())
		assert.Equal(t, "Моя карта", resp.GetCard().GetName())
		assert.Equal(t, validCardNum, resp.GetCard().GetNumber())
		createdCardID = resp.GetCard().GetId()
		assert.NotEmpty(t, createdCardID)
	})

	t.Run("15. CardList - получение списка карт", func(t *testing.T) {
		ctx := metadata.NewOutgoingContext(context.Background(), metadata.Pairs("authorization", "Bearer "+authToken))
		req := &api.ListRequest{}

		resp, err := client.CardList(ctx, req)
		require.NoError(t, err)
		require.Len(t, resp.GetCard(), 1)
		assert.Equal(t, "Моя карта", resp.GetCard()[0].GetName())
	})

	t.Run("16. CardUpdate - обновление карты", func(t *testing.T) {
		ctx := metadata.NewOutgoingContext(context.Background(), metadata.Pairs("authorization", "Bearer "+authToken))
		req := pbCard.UpdateRequest_builder{
			Id:      &createdCardID,
			Name:    stringPtr("Обновленная карта"),
			Number:  stringPtr(validCardNum),
			Expired: stringPtr("0130"),
			Code:    stringPtr("999"),
		}.Build()

		resp, err := client.CardUpdate(ctx, req)
		require.NoError(t, err)
		require.NotNil(t, resp.GetCard())
		assert.Equal(t, "Обновленная карта", resp.GetCard().GetName())
	})

	t.Run("17. CardDelete - удаление карты", func(t *testing.T) {
		ctx := metadata.NewOutgoingContext(context.Background(), metadata.Pairs("authorization", "Bearer "+authToken))
		req := api.IDRequest_builder{
			Id: &createdCardID,
		}.Build()

		resp, err := client.CardDelete(ctx, req)
		require.NoError(t, err)
		require.NotNil(t, resp.GetCard())
		assert.Equal(t, createdCardID, resp.GetCard().GetId())
	})

	// === Text CRUD через gRPC клиент ===
	var createdTextID string

	t.Run("18. TextCreate с токеном - успех", func(t *testing.T) {
		req := pbText.CreateRequest_builder{
			Name: stringPtr("Секретная заметка"),
			Text: stringPtr("Очень важная информация"),
		}.Build()

		ctx := metadata.NewOutgoingContext(context.Background(), metadata.Pairs("authorization", "Bearer "+authToken))
		resp, err := client.TextCreate(ctx, req)
		require.NoError(t, err)
		require.NotNil(t, resp.GetText())
		assert.Equal(t, "Секретная заметка", resp.GetText().GetName())
		assert.Equal(t, "Очень важная информация", resp.GetText().GetText())
		createdTextID = resp.GetText().GetId()
		assert.NotEmpty(t, createdTextID)
	})

	t.Run("19. TextList - получение списка заметок", func(t *testing.T) {
		ctx := metadata.NewOutgoingContext(context.Background(), metadata.Pairs("authorization", "Bearer "+authToken))
		req := &api.ListRequest{}

		resp, err := client.TextList(ctx, req)
		require.NoError(t, err)
		require.Len(t, resp.GetText(), 1)
		assert.Equal(t, "Секретная заметка", resp.GetText()[0].GetName())
	})

	t.Run("20. TextUpdate - обновление заметки", func(t *testing.T) {
		ctx := metadata.NewOutgoingContext(context.Background(), metadata.Pairs("authorization", "Bearer "+authToken))
		req := pbText.UpdateRequest_builder{
			Id:   &createdTextID,
			Name: stringPtr("Обновленная заметка"),
			Text: stringPtr("Обновленная информация"),
		}.Build()

		resp, err := client.TextUpdate(ctx, req)
		require.NoError(t, err)
		require.NotNil(t, resp.GetText())
		assert.Equal(t, "Обновленная заметка", resp.GetText().GetName())
	})

	t.Run("21. TextDelete - удаление заметки", func(t *testing.T) {
		ctx := metadata.NewOutgoingContext(context.Background(), metadata.Pairs("authorization", "Bearer "+authToken))
		req := api.IDRequest_builder{
			Id: &createdTextID,
		}.Build()

		resp, err := client.TextDelete(ctx, req)
		require.NoError(t, err)
		require.NotNil(t, resp.GetText())
		assert.Equal(t, createdTextID, resp.GetText().GetId())
	})
}
