// Пакет для реализации аутентификации
package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/go-chi/jwtauth/v5"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// TokenExpDuration Длительность жизни токена
const TokenExpDuration = 7 * 24 * time.Hour

// Guard Объект для реализации аутентификации
type Guard struct {
	jwt *jwtauth.JWTAuth
}

// New Создание нового объекта аутентификации
func New(secret string) *Guard {
	return &Guard{
		jwt: jwtauth.New("HS256", []byte(secret), nil),
	}
}

// GetHash Хеширование строки
func (g Guard) GetHash(s string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(s), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("ошибка при хешировании строки: %w", err)
	}

	return string(hash), nil
}

// CheckPassword Проверка соответствия строки хешу
func (g Guard) CheckPassword(s, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(s))
	return err == nil
}

// NewUserToken Создание нового токена пользователя
func (g Guard) NewUserToken(userId string) (string, time.Time, error) {
	// Данные пользователя (Claims)
	claims := map[string]interface{}{
		"user_id": userId,
	}

	// Устанавливаем время жизни токена
	jwtauth.SetExpiry(claims, time.Now().Add(TokenExpDuration))

	// Генерируем и подписываем токен
	token, tokenString, err := g.jwt.Encode(claims)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("ошибка при создании токена пользователя: %w", err)
	}

	// Момент окончания действия токена
	expiredAt, _ := token.Expiration()

	return tokenString, expiredAt, nil
}

// JWTAuth Возвращает объект JWTAuth
func (g Guard) JWTAuth() *jwtauth.JWTAuth {
	return g.jwt
}

// ExtractTokenFromMetadata извлекает токен авторизации из входящих метаданных gRPC
func ExtractTokenFromMetadata(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}

	// Проверяем заголовок authorization
	if vals := md.Get("authorization"); len(vals) > 0 && vals[0] != "" {
		val := vals[0]
		if strings.HasPrefix(strings.ToLower(val), "bearer ") {
			return strings.TrimSpace(val[7:])
		}
		return strings.TrimSpace(val)
	}

	// Проверяем заголовок token
	if vals := md.Get("token"); len(vals) > 0 && vals[0] != "" {
		return strings.TrimSpace(vals[0])
	}

	// Проверяем заголовок jwt
	if vals := md.Get("jwt"); len(vals) > 0 && vals[0] != "" {
		return strings.TrimSpace(vals[0])
	}

	return ""
}

// isGuestMethod проверяет, относится ли вызываемый gRPC метод к гостевым (требующим отсутствия аутентификации)
func isGuestMethod(fullMethod string) bool {
	return strings.HasSuffix(fullMethod, "/RegisterUser") ||
		strings.HasSuffix(fullMethod, "/LoginUser") ||
		strings.EqualFold(fullMethod, "RegisterUser") ||
		strings.EqualFold(fullMethod, "LoginUser") ||
		strings.Contains(fullMethod, "Register") ||
		strings.Contains(fullMethod, "Login")
}

// UnaryServerInterceptor возвращает unary interceptor для gRPC сервера с аутентификацией через metadata
func (g Guard) UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		tokenString := ExtractTokenFromMetadata(ctx)

		if isGuestMethod(info.FullMethod) {
			// Если эндпоинт предназначен для неаутентифицированного пользователя - проверяем отсутствие аутентификации
			if tokenString != "" {
				token, err := g.jwt.Decode(tokenString)
				if err == nil && token != nil {
					return nil, status.Error(codes.PermissionDenied, "пользователь уже аутентифицирован в системе")
				}
			}
			return handler(ctx, req)
		}

		// Проверяем наличие токена
		if tokenString == "" {
			return nil, status.Error(codes.Unauthenticated, "отсутствует токен аутентификации")
		}

		token, err := g.jwt.Decode(tokenString)
		if err != nil || token == nil {
			return nil, status.Error(codes.Unauthenticated, "недействительный токен аутентификации")
		}

		// Помещаем токен в контекст для обработчика
		ctx = jwtauth.NewContext(ctx, token, nil)

		return handler(ctx, req)
	}
}

// UnaryInterceptor псевдоним для UnaryServerInterceptor
func (g Guard) UnaryInterceptor() grpc.UnaryServerInterceptor {
	return g.UnaryServerInterceptor()
}

// UnaryServerInterceptor возвращает unary interceptor для переданного Guard
func UnaryServerInterceptor(g *Guard) grpc.UnaryServerInterceptor {
	return g.UnaryServerInterceptor()
}

// UnaryInterceptor возвращает unary interceptor для переданного Guard
func UnaryInterceptor(g *Guard) grpc.UnaryServerInterceptor {
	return g.UnaryServerInterceptor()
}

// UserIDFromContext извлекает user_id из токена в контексте
func UserIDFromContext(ctx context.Context) (string, error) {
	_, claims, err := jwtauth.FromContext(ctx)
	if err != nil {
		return "", err
	}
	if claims == nil {
		return "", errors.New("не удалось извлечь пользователя из контекста")
	}
	userID, ok := claims["user_id"].(string)
	if !ok {
		return "", errors.New("user_id не найден в claims")
	}
	return userID, nil
}
