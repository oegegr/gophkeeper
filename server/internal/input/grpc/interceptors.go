package grpc

import (
	"context"
	"log"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/oegegr/gophkeeper/server/internal/domain"
	"github.com/samber/do/v2"
)

// TokenValidator управляет токенами
type TokenValidator interface {
	ValidateToken(token string) (string, error)
}

// Константы для публичных методов (не требуют аутентификации)
const (
	publicMethodRegister = "/gophkeeper.GophKeeper/Register"
	publicMethodLogin    = "/gophkeeper.GophKeeper/Login"
)

// AuthInterceptor интерцептор для аутентификации
func ResolveAuthInterceptor(i do.Injector) grpc.UnaryServerInterceptor {
	tokenValidator := do.MustInvokeAs[TokenValidator](i)

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Публичные методы не требуют аутентификации
		if isPublicMethod(info.FullMethod) {
			return handler(ctx, req)
		}

		// Для приватных методов проверяем токен
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "missing metadata")
		}

		authHeaders := md.Get("authorization")
		if len(authHeaders) == 0 {
			return nil, status.Error(codes.Unauthenticated, "missing authorization token")
		}

		token := authHeaders[0]
		if !strings.HasPrefix(token, "Bearer ") {
			return nil, status.Error(codes.Unauthenticated, "invalid token format")
		}

		token = strings.TrimPrefix(token, "Bearer ")
		userID, err := tokenValidator.ValidateToken(token)
		if err != nil {
			log.Printf("Invalid JWT token: %v", err)
			return nil, status.Error(codes.Unauthenticated, "invalid token")
		}

		// Добавляем userID в контекст
		ctx = context.WithValue(ctx, domain.UserContextKey, userID)

		return handler(ctx, req)
	}
}

// isPublicMethod проверяет, является ли метод публичным
func isPublicMethod(fullMethod string) bool {
	switch fullMethod {
	case publicMethodRegister, publicMethodLogin:
		return true
	default:
		return false
	}
}
