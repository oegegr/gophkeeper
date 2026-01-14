package usecases

import (
	"context"

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

func (uc *SyncUseCaseImpl) Sync(ctx context.Context, userID string, clientSecrets []domain.Secret) ([]domain.Secret, error) {
	serverSecrets, err := uc.repos.FindSecretsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return serverSecrets, nil
}