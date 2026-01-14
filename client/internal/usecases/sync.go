package usecases

import (
	"context"

	"github.com/oegegr/gophkeeper/client/internal/domain"
	"github.com/oegegr/gophkeeper/client/internal/input/console"
	"github.com/pkg/errors"
	"github.com/samber/do/v2"
)

var _ console.SyncUseCase = (*SyncUseCase)(nil)

type SyncClient interface {
	Sync(ctx context.Context, secrets []domain.Secret, session domain.Session) ([]domain.Secret, error)
}

type SyncStorage interface {
	GetSecrets(ctx context.Context) ([]domain.Secret, error)
	SaveSecrets(ctx context.Context, secrets []domain.Secret) error
	DeleteSecret(ctx context.Context, secret domain.Secret) error
}

type SyncUseCase struct {
	client  SyncClient
	storage SyncStorage
	sessionStorage SessionStorage
}

func ResolveSyncUseCase(i do.Injector) *SyncUseCase {
	storage := do.MustInvokeAs[SyncStorage](i)
	sessionStorage := do.MustInvokeAs[SessionStorage](i)
	client := do.MustInvokeAs[SyncClient](i)
	return NewSyncUseCase(client, storage, sessionStorage)

}

func NewSyncUseCase(client SyncClient, storage SyncStorage, sessionStorage SessionStorage) *SyncUseCase {
	return &SyncUseCase{
		client:  client,
		storage: storage,
		sessionStorage: sessionStorage,
	}
}

func (uc *SyncUseCase) Sync(ctx context.Context) error {
	session, err := uc.sessionStorage.GetSession(ctx)
	if err != nil {
		return errors.Wrap(err, "missing session: try to login")
	}

	localSecrets, err := uc.storage.GetSecrets(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get local secrets")
	}

	serverSecrets, err := uc.client.Sync(ctx, localSecrets, *session)
	if err != nil {
		return errors.Wrap(err, "failed to sync secrets")
	}

	localMap := make(map[string]domain.Secret)
	for _, s := range localSecrets {
		localMap[s.ID] = s
	}

	serverMap := make(map[string]domain.Secret)
	for _, s := range serverSecrets {
		serverMap[s.ID] = s
	}

	toSave := []domain.Secret{}
	for _, serverSecret := range serverSecrets {
		localSecret, exists := localMap[serverSecret.ID]

		if !exists {
			toSave = append(toSave, serverSecret)
		} else if serverSecret.UpdatedAt.After(localSecret.UpdatedAt) {
			toSave = append(toSave, serverSecret)
		}
	}

	for id := range localMap {
		if secret, existsOnServer := serverMap[id]; !existsOnServer {
			uc.storage.DeleteSecret(ctx, secret)
		}
	}

	if len(toSave) > 0 {
		return uc.storage.SaveSecrets(ctx, toSave)
	}

	return nil
}