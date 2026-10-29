package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/jwtauth/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// TestNew Тестирование создания объекта аутентификации
func TestNew(t *testing.T) {
	secret := "secret"
	guard := New(secret)
	assert.NotNil(t, guard)
	assert.NotNil(t, guard.jwt)
}

// TestNewUserToken Проверка создания токена аутентификации пользователя
func TestNewUserToken(t *testing.T) {
	secret := "secret"
	guard := New(secret)
	userId := "user123"

	tokenString, expiredAt, err := guard.NewUserToken(userId)
	require.NoError(t, err)
	assert.NotEmpty(t, tokenString)
	assert.WithinDuration(t, time.Now().Add(TokenExpDuration), expiredAt, 10*time.Second)

	// Проверяем содержимое токена
	token, err := guard.jwt.Decode(tokenString)
	require.NoError(t, err)

	var val string
	err = token.Get("user_id", &val)
	require.NoError(t, err)
	assert.Equal(t, userId, val)
}

// TestCheckPassword Проверка сравнения пароля с хешем
func TestCheckPassword(t *testing.T) {
	guard := New("secret")
	password := "mypassword"

	hash, err := guard.GetHash(password)
	require.NoError(t, err)
	assert.NotEmpty(t, hash)

	assert.True(t, guard.CheckPassword(password, hash))
	assert.False(t, guard.CheckPassword("wrongpassword", hash))
}

// TestGuestMiddleware Проверка посредника GuestMiddleware (только для неаутентифицированных пользователей)
func TestGuestMiddleware(t *testing.T) {
	secret := "secret"
	guard := New(secret)

	// Middleware должен работать в связке с Verifier
	h := jwtauth.Verifier(guard.jwt)(guard.GuestMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})))

	t.Run("positive - GuestMiddleware для неавторизированного пользователя", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("negative - GuestMiddleware для авторизированного пользователя", func(t *testing.T) {
		tokenString, _, _ := guard.NewUserToken("user123")
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer "+tokenString)

		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusForbidden, rec.Code)
	})
}

// TestAuthGroup Проверка группы роутов только для аутентифицированных пользователей
func TestAuthGroup(t *testing.T) {
	secret := "secret"
	guard := New(secret)
	userId := "user123"

	mux := http.NewServeMux()
	handler := jwtauth.Verifier(guard.jwt)(jwtauth.Authenticator(guard.jwt)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, claims, _ := jwtauth.FromContext(r.Context())
		assert.Equal(t, userId, claims["user_id"])
		w.WriteHeader(http.StatusOK)
	})))
	mux.Handle("GET /protected", handler)

	ts := httptest.NewServer(mux)
	defer ts.Close()

	t.Run("positive - AuthGroup для авторизированного пользователя", func(t *testing.T) {
		tokenString, _, _ := guard.NewUserToken(userId)
		req, _ := http.NewRequest(http.MethodGet, ts.URL+"/protected", nil)
		req.Header.Set("Authorization", "Bearer "+tokenString)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("negative - AuthGroup для неавторизированного пользователя", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/protected")
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})
}

// TestGuestGroup Проверка группы роутов только для неаутентифицированных пользователей
func TestGuestGroup(t *testing.T) {
	secret := "secret"
	guard := New(secret)

	mux := http.NewServeMux()
	handler := jwtauth.Verifier(guard.jwt)(guard.GuestMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})))
	mux.Handle("GET /guest-only", handler)

	ts := httptest.NewServer(mux)
	defer ts.Close()

	t.Run("positive - GuestGroup для неавторизированного пользователя", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/guest-only")
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("negative - GuestGroup для авторизированного пользователя", func(t *testing.T) {
		tokenString, _, _ := guard.NewUserToken("user123")
		req, _ := http.NewRequest(http.MethodGet, ts.URL+"/guest-only", nil)
		req.Header.Set("Authorization", "Bearer "+tokenString)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})
}

// TestUnaryServerInterceptor Проверка gRPC unary интерцептора аутентификации
func TestUnaryServerInterceptor(t *testing.T) {
	secret := "secret"
	guard := New(secret)
	interceptor := guard.UnaryServerInterceptor()

	dummyHandler := func(ctx context.Context, req any) (any, error) {
		userID, _ := UserIDFromContext(ctx)
		return userID, nil
	}

	t.Run("guest method - positive (без токена для RegisterUser)", func(t *testing.T) {
		ctx := context.Background()
		info := &grpc.UnaryServerInfo{
			FullMethod: "/models.Users/RegisterUser",
		}
		resp, err := interceptor(ctx, "req", info, dummyHandler)
		require.NoError(t, err)
		assert.Equal(t, "", resp)
	})

	t.Run("guest method - positive (без токена для LoginUser)", func(t *testing.T) {
		ctx := context.Background()
		info := &grpc.UnaryServerInfo{
			FullMethod: "/models.Users/LoginUser",
		}
		resp, err := interceptor(ctx, "req", info, dummyHandler)
		require.NoError(t, err)
		assert.Equal(t, "", resp)
	})

	t.Run("guest method - negative (с валидным токеном в authorization Bearer)", func(t *testing.T) {
		tokenString, _, err := guard.NewUserToken("user123")
		require.NoError(t, err)

		md := metadata.Pairs("authorization", "Bearer "+tokenString)
		ctx := metadata.NewIncomingContext(context.Background(), md)
		info := &grpc.UnaryServerInfo{
			FullMethod: "/models.Users/RegisterUser",
		}

		resp, err := interceptor(ctx, "req", info, dummyHandler)
		require.Error(t, err)
		assert.Nil(t, resp)

		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.PermissionDenied, st.Code())
	})

	t.Run("guest method - negative (с валидным токеном в заголовке token)", func(t *testing.T) {
		tokenString, _, err := guard.NewUserToken("user123")
		require.NoError(t, err)

		md := metadata.Pairs("token", tokenString)
		ctx := metadata.NewIncomingContext(context.Background(), md)
		info := &grpc.UnaryServerInfo{
			FullMethod: "/models.Users/LoginUser",
		}

		resp, err := interceptor(ctx, "req", info, dummyHandler)
		require.Error(t, err)
		assert.Nil(t, resp)

		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.PermissionDenied, st.Code())
	})

	t.Run("guest method - positive (с невалидным токеном обратная проверка разрешает вызов)", func(t *testing.T) {
		md := metadata.Pairs("authorization", "Bearer invalid-token")
		ctx := metadata.NewIncomingContext(context.Background(), md)
		info := &grpc.UnaryServerInfo{
			FullMethod: "/models.Users/RegisterUser",
		}

		resp, err := interceptor(ctx, "req", info, dummyHandler)
		require.NoError(t, err)
		assert.Equal(t, "", resp)
	})

	t.Run("protected method - positive (с валидным токеном)", func(t *testing.T) {
		tokenString, _, err := guard.NewUserToken("user123")
		require.NoError(t, err)

		md := metadata.Pairs("authorization", "Bearer "+tokenString)
		ctx := metadata.NewIncomingContext(context.Background(), md)
		info := &grpc.UnaryServerInfo{
			FullMethod: "/models.Data/GetData",
		}

		resp, err := interceptor(ctx, "req", info, dummyHandler)
		require.NoError(t, err)
		assert.Equal(t, "user123", resp)
	})

	t.Run("protected method - negative (без метаданных/токена)", func(t *testing.T) {
		ctx := context.Background()
		info := &grpc.UnaryServerInfo{
			FullMethod: "/models.Data/GetData",
		}

		resp, err := interceptor(ctx, "req", info, dummyHandler)
		require.Error(t, err)
		assert.Nil(t, resp)

		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.Unauthenticated, st.Code())
	})

	t.Run("protected method - negative (с невалидным токеном)", func(t *testing.T) {
		md := metadata.Pairs("authorization", "Bearer invalid.token.value")
		ctx := metadata.NewIncomingContext(context.Background(), md)
		info := &grpc.UnaryServerInfo{
			FullMethod: "/models.Data/GetData",
		}

		resp, err := interceptor(ctx, "req", info, dummyHandler)
		require.Error(t, err)
		assert.Nil(t, resp)

		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.Unauthenticated, st.Code())
	})
}
