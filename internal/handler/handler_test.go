package handler

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/alexsey-popov/gkeep/internal/auth"
	"github.com/alexsey-popov/gkeep/internal/models"
	"github.com/alexsey-popov/gkeep/internal/repository"
	api "github.com/alexsey-popov/gkeep/proto"
	pbPassword "github.com/alexsey-popov/gkeep/proto/password"
	pb "github.com/alexsey-popov/gkeep/proto/user"
	"github.com/go-chi/jwtauth/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

// mockRepository мок для репозитория
type mockRepository struct {
	users     map[string]*models.User
	passwords map[string]*models.Password
}

func newMockRepository() *mockRepository {
	return &mockRepository{
		users:     make(map[string]*models.User),
		passwords: make(map[string]*models.Password),
	}
}

func (m *mockRepository) CreateUser(ctx context.Context, login, password string) (*models.User, error) {
	if _, exists := m.users[login]; exists {
		return nil, repository.ErrConflictUnique
	}

	user := &models.User{
		Id:       "user-uuid-1",
		Login:    login,
		Password: password,
	}
	m.users[login] = user
	return user, nil
}

func (m *mockRepository) GetUser(ctx context.Context, login string) (*models.User, error) {
	user, exists := m.users[login]
	if !exists {
		return nil, errors.New("пользователь не найден")
	}
	return user, nil
}

func (m *mockRepository) PasswordCreate(ctx context.Context, userID, name, login, password string) (*models.Password, error) {
	for _, p := range m.passwords {
		if p.UserId == userID && p.Name == name {
			return nil, repository.ErrConflictUnique
		}
	}
	id := fmt.Sprintf("pwd-%d", len(m.passwords)+1)
	pwd := &models.Password{
		Id:        id,
		UserId:    userID,
		Name:      name,
		Login:     login,
		Password:  password,
		CreatedAt: time.Now(),
	}
	m.passwords[id] = pwd
	return pwd, nil
}

func (m *mockRepository) PasswordList(ctx context.Context, userID, name, page string) ([]models.Password, error) {
	var res []models.Password
	for _, p := range m.passwords {
		if p.UserId == userID {
			if name != "" && !strings.Contains(strings.ToLower(p.Name), strings.ToLower(name)) {
				continue
			}
			res = append(res, *p)
		}
	}
	return res, nil
}

func (m *mockRepository) PasswordUpdate(ctx context.Context, userID, login, password string) (*models.Password, error) {
	for _, p := range m.passwords {
		if p.UserId == userID && p.Login == login {
			p.Password = password
			return p, nil
		}
	}
	return nil, sql.ErrNoRows
}

func (m *mockRepository) PasswordDelete(ctx context.Context, userID, id string) (*models.Password, error) {
	p, exists := m.passwords[id]
	if !exists || p.UserId != userID {
		return nil, sql.ErrNoRows
	}
	delete(m.passwords, id)
	return p, nil
}

func TestHandler_RegisterAndLoginUser(t *testing.T) {
	secret := "secret-key"
	guard := auth.New(secret)
	rep := newMockRepository()
	h := New(rep, guard)

	t.Run("RegisterUser - успешная регистрация и отправка метаданных", func(t *testing.T) {
		req := &pb.UserRequest{}
		req.SetLogin("testuser")
		req.SetPassword("password123")

		resp, err := h.RegisterUser(context.Background(), req)
		require.NoError(t, err)
		assert.Equal(t, "user-uuid-1", resp.GetUser().GetId())
		assert.Equal(t, "testuser", resp.GetUser().GetLogin())

		// Проверяем, что в БД пароль сохранен в виде хеша
		savedUser, err := rep.GetUser(context.Background(), "testuser")
		require.NoError(t, err)
		assert.True(t, guard.CheckPassword("password123", savedUser.Password))
	})

	t.Run("RegisterUser - конфликт логина", func(t *testing.T) {
		req := &pb.UserRequest{}
		req.SetLogin("testuser")
		req.SetPassword("password123")

		resp, err := h.RegisterUser(context.Background(), req)
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.AlreadyExists, st.Code())
		assert.Nil(t, resp.GetUser())
	})

	t.Run("LoginUser - успешный вход", func(t *testing.T) {
		req := &pb.UserRequest{}
		req.SetLogin("testuser")
		req.SetPassword("password123")

		resp, err := h.LoginUser(context.Background(), req)
		require.NoError(t, err)
		assert.Equal(t, "user-uuid-1", resp.GetUser().GetId())
		assert.Equal(t, "testuser", resp.GetUser().GetLogin())
	})

	t.Run("LoginUser - неверный пароль", func(t *testing.T) {
		req := &pb.UserRequest{}
		req.SetLogin("testuser")
		req.SetPassword("wrongpassword")

		resp, err := h.LoginUser(context.Background(), req)
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.Unauthenticated, st.Code())
		assert.Nil(t, resp.GetUser())
	})

	t.Run("LoginUser - несуществующий пользователь", func(t *testing.T) {
		req := &pb.UserRequest{}
		req.SetLogin("nonexistent")
		req.SetPassword("password123")

		resp, err := h.LoginUser(context.Background(), req)
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.Unknown, st.Code())
		assert.Nil(t, resp.GetUser())
	})
}

func TestGRPCServer_IntegrationWithInterceptor(t *testing.T) {
	secret := "secret-jwt-token"
	guard := auth.New(secret)
	rep := newMockRepository()
	h := New(rep, guard)

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
		req := &pb.UserRequest{}
		req.SetLogin("john_doe")
		req.SetPassword("strong_password")

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
		req := &pb.UserRequest{}
		req.SetLogin("john_doe")
		req.SetPassword("strong_password")

		_, err := client.RegisterUser(context.Background(), req)
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.AlreadyExists, st.Code())
	})

	t.Run("3. Регистрация авторизованным пользователем (обратная проверка) - ошибка PermissionDenied", func(t *testing.T) {
		req := &pb.UserRequest{}
		req.SetLogin("another_user")
		req.SetPassword("strong_password")

		ctx := metadata.NewOutgoingContext(context.Background(), metadata.Pairs("authorization", "Bearer "+authToken))
		_, err := client.RegisterUser(ctx, req)
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.PermissionDenied, st.Code())
	})

	t.Run("4. Логин авторизованным пользователем (обратная проверка) - ошибка PermissionDenied", func(t *testing.T) {
		req := &pb.UserRequest{}
		req.SetLogin("john_doe")
		req.SetPassword("strong_password")

		ctx := metadata.NewOutgoingContext(context.Background(), metadata.Pairs("authorization", "Bearer "+authToken))
		_, err := client.LoginUser(ctx, req)
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.PermissionDenied, st.Code())
	})

	t.Run("5. Успешный логин неавторизованным пользователем - успех + получение токена в metadata", func(t *testing.T) {
		req := &pb.UserRequest{}
		req.SetLogin("john_doe")
		req.SetPassword("strong_password")

		var header metadata.MD
		resp, err := client.LoginUser(context.Background(), req, grpc.Header(&header))
		require.NoError(t, err)
		assert.Equal(t, "john_doe", resp.GetUser().GetLogin())

		tokens := header.Get("token")
		require.NotEmpty(t, tokens)
		assert.NotEmpty(t, tokens[0])
	})

	t.Run("6. Логин с неверным паролем - ошибка Unauthenticated", func(t *testing.T) {
		req := &pb.UserRequest{}
		req.SetLogin("john_doe")
		req.SetPassword("wrong_password")

		_, err := client.LoginUser(context.Background(), req)
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.Unauthenticated, st.Code())
	})

	var createdPasswordID string

	t.Run("7. PasswordCreate без токена - ошибка Unauthenticated", func(t *testing.T) {
		req := &pbPassword.PasswordCreateRequest{}
		req.SetName("Google")
		req.SetLogin("john@gmail.com")
		req.SetPassword("mysecret123")

		_, err := client.PasswordCreate(context.Background(), req)
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.Unauthenticated, st.Code())
	})

	t.Run("8. PasswordCreate с токеном - успех", func(t *testing.T) {
		req := &pbPassword.PasswordCreateRequest{}
		req.SetName("Google")
		req.SetLogin("john@gmail.com")
		req.SetPassword("mysecret123")

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
		req := &pbPassword.PasswordCreateRequest{}
		req.SetName("Google")
		req.SetLogin("john@gmail.com")
		req.SetPassword("otherpass")

		ctx := metadata.NewOutgoingContext(context.Background(), metadata.Pairs("authorization", "Bearer "+authToken))
		_, err := client.PasswordCreate(ctx, req)
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.AlreadyExists, st.Code())
	})

	t.Run("10. PasswordList - получение списка паролей", func(t *testing.T) {
		ctx := metadata.NewOutgoingContext(context.Background(), metadata.Pairs("authorization", "Bearer "+authToken))
		req := &pbPassword.PasswordListRequest{}

		resp, err := client.PasswordList(ctx, req)
		require.NoError(t, err)
		require.Len(t, resp.GetPassword(), 1)
		assert.Equal(t, "Google", resp.GetPassword()[0].GetName())
	})

	t.Run("11. PasswordList с фильтрацией по имени", func(t *testing.T) {
		ctx := metadata.NewOutgoingContext(context.Background(), metadata.Pairs("authorization", "Bearer "+authToken))

		// Добавляем еще одну запись
		req2 := &pbPassword.PasswordCreateRequest{}
		req2.SetName("GitHub")
		req2.SetLogin("john_git")
		req2.SetPassword("gitpass")
		_, err := client.PasswordCreate(ctx, req2)
		require.NoError(t, err)

		// Поиск по "goog"
		listReq := &pbPassword.PasswordListRequest{}
		listReq.SetName("goog")
		resp, err := client.PasswordList(ctx, listReq)
		require.NoError(t, err)
		require.Len(t, resp.GetPassword(), 1)
		assert.Equal(t, "Google", resp.GetPassword()[0].GetName())
	})

	t.Run("12. PasswordUpdate - успешное обновление пароля", func(t *testing.T) {
		ctx := metadata.NewOutgoingContext(context.Background(), metadata.Pairs("authorization", "Bearer "+authToken))
		req := &pbPassword.PasswordUpdateRequest{}
		req.SetLogin("john@gmail.com")
		req.SetPassword("newsecret456")

		resp, err := client.PasswordUpdate(ctx, req)
		require.NoError(t, err)
		require.NotNil(t, resp.GetPassword())
		assert.Equal(t, "newsecret456", resp.GetPassword().GetPassword())
	})

	t.Run("13. PasswordDelete - успешное удаление записи", func(t *testing.T) {
		ctx := metadata.NewOutgoingContext(context.Background(), metadata.Pairs("authorization", "Bearer "+authToken))
		req := &pbPassword.PasswordDeleteRequest{}
		req.SetId(createdPasswordID)

		resp, err := client.PasswordDelete(ctx, req)
		require.NoError(t, err)
		require.NotNil(t, resp.GetPassword())
		assert.Equal(t, createdPasswordID, resp.GetPassword().GetId())

		// Проверяем, что удаленная запись больше не возвращается
		listReq := &pbPassword.PasswordListRequest{}
		listReq.SetName("Google")
		listResp, err := client.PasswordList(ctx, listReq)
		require.NoError(t, err)
		assert.Empty(t, listResp.GetPassword())
	})
}

func TestHandler_PasswordUnitMethods(t *testing.T) {
	secret := "secret-jwt-token"
	guard := auth.New(secret)
	rep := newMockRepository()
	h := New(rep, guard)

	tokenString, _, err := guard.NewUserToken("user-123")
	require.NoError(t, err)

	token, err := guard.JWTAuth().Decode(tokenString)
	require.NoError(t, err)

	authCtx := jwtauth.NewContext(context.Background(), token, nil)

	t.Run("PasswordCreate - пустой контекст без токена", func(t *testing.T) {
		req := &pbPassword.PasswordCreateRequest{}
		req.SetName("Test")
		req.SetLogin("login")
		req.SetPassword("pass")

		resp, err := h.PasswordCreate(context.Background(), req)
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.Unauthenticated, st.Code())
		assert.Nil(t, resp.GetPassword())
	})

	t.Run("PasswordCreate - пустое имя", func(t *testing.T) {
		req := &pbPassword.PasswordCreateRequest{}
		req.SetName("   ")
		req.SetLogin("login")
		req.SetPassword("pass")

		resp, err := h.PasswordCreate(authCtx, req)
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, st.Code())
		assert.Nil(t, resp.GetPassword())
	})

	t.Run("PasswordUpdate - пустой логин", func(t *testing.T) {
		req := &pbPassword.PasswordUpdateRequest{}
		req.SetLogin("   ")
		req.SetPassword("pass")

		resp, err := h.PasswordUpdate(authCtx, req)
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, st.Code())
		assert.Nil(t, resp.GetPassword())
	})

	t.Run("PasswordUpdate - запись не найдена", func(t *testing.T) {
		req := &pbPassword.PasswordUpdateRequest{}
		req.SetLogin("unknown_login")
		req.SetPassword("pass")

		resp, err := h.PasswordUpdate(authCtx, req)
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.NotFound, st.Code())
		assert.Nil(t, resp.GetPassword())
	})

	t.Run("PasswordDelete - пустой id", func(t *testing.T) {
		req := &pbPassword.PasswordDeleteRequest{}
		req.SetId("   ")

		resp, err := h.PasswordDelete(authCtx, req)
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, st.Code())
		assert.Nil(t, resp.GetPassword())
	})

	t.Run("PasswordDelete - несуществующий id", func(t *testing.T) {
		req := &pbPassword.PasswordDeleteRequest{}
		req.SetId("non-existent-id")

		resp, err := h.PasswordDelete(authCtx, req)
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.NotFound, st.Code())
		assert.Nil(t, resp.GetPassword())
	})
}
