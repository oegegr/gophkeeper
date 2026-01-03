package usecases

import (
	"context"
	"time"

	"github.com/oegegr/gophkeeper/server/internal/domain"
	"github.com/samber/do/v2"
)

type SyncRepository interface {
	FindSecretsByUserID(ctx context.Context, userID string) ([]domain.Secret, error)
}

// SyncUseCaseImpl реализация SyncUseCase
type SyncUseCaseImpl struct {
	repos SyncRepository
}

func ResolveSyncUseCase(i do.Injector) *SyncUseCaseImpl {
	repo := do.MustInvokeAs[SyncRepository](i)

	return NewSyncUseCase(repo)

}

// NewSyncUseCase создает новый SyncUseCase
func NewSyncUseCase(repos SyncRepository) *SyncUseCaseImpl {
	return &SyncUseCaseImpl{
		repos: repos,
	}
}

// Sync синхронизирует секреты между клиентом и сервером
func (uc *SyncUseCaseImpl) Sync(ctx context.Context, userID string, clientSecrets []domain.Secret) ([]domain.Secret, error) {
	// Валидация
	if userID == "" {
		return nil, domain.ErrInvalidInput
	}

	// Если клиентские секреты пусты, возвращаем серверные
	if len(clientSecrets) == 0 {
		return uc.repos.FindSecretsByUserID(ctx, userID)
	}

	// В реальности здесь была бы сложная логика синхронизации:
	// 1. Разрешение конфликтов
	// 2. Объединение изменений
	// 3. Возврат актуальной версии

	// Для мока просто возвращаем тестовые данные
	return uc.generateSyncSecrets(userID), nil
}

// generateSyncSecrets создает тестовые секреты для синхронизации
func (uc *SyncUseCaseImpl) generateSyncSecrets(userID string) []domain.Secret {
	now := time.Now()

	return []domain.Secret{
		{
			ID:        "sync_secret_1",
			UserID:    userID,
			Type:      domain.SecretTypeLoginPassword,
			Data:      []byte(`{"login":"sync@example.com","password":"sync123"}`),
			Meta:      "Sync test website",
			Version:   1,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:        "sync_secret_2",
			UserID:    userID,
			Type:      domain.SecretTypeCard,
			Data:      []byte(`{"number":"****1234","holder":"John Doe"}`),
			Meta:      "Credit card",
			Version:   1,
			CreatedAt: now,
			UpdatedAt: now,
		},
	}
}
