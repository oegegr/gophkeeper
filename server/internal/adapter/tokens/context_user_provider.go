package tokens

import (
	"context"
	"fmt"

	"github.com/oegegr/gophkeeper/server/internal/domain"
	"github.com/samber/do/v2"
)

// AuthContextUserIDPovider предоставляет провайдер для получения идентификатора пользователя из контекста запроса.
type ContextUserPovider struct{}

func ResolveContextUserProvider(i do.Injector) *ContextUserPovider {
	return &ContextUserPovider{}
}

// Get возвращает идентификатор пользователя из контекста запроса.
func (a *ContextUserPovider) Get(ctx context.Context) (string, error) {
	value := ctx.Value(domain.UserContextKey)
	if value == nil {
		return "", fmt.Errorf("failed to get userID from context")
	}

	userID, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("failed to convert userID to string")
	}
	return userID, nil
}
