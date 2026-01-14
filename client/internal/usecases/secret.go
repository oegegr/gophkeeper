package usecases

import (
	"context"

	"github.com/oegegr/gophkeeper/client/internal/domain"
	"github.com/oegegr/gophkeeper/client/internal/input/console"
	"github.com/pkg/errors"
	"github.com/samber/do/v2"
)

var _ console.SecretUseCase = (*SecretUseCase)(nil)

type SecretStorage interface {
	GetSecrets(ctx context.Context) ([]domain.Secret, error)
	GetSecret(ctx context.Context, id string) (domain.Secret, error)
	SaveSecret(ctx context.Context, secret domain.Secret) error
	DeleteSecret(ctx context.Context, secret domain.Secret) error
}

type SecretClient interface {
	AddSecret(ctx context.Context, secret domain.Secret, session domain.Session) (*domain.Secret, error)
	DeleteSecret(ctx context.Context, secret domain.Secret, session domain.Session) error
}

type SecretUseCase struct {
	storage SecretStorage
	sessionStorage SessionStorage
	client  SecretClient
}

func ResolveSecretUseCase(i do.Injector) *SecretUseCase {
	storage := do.MustInvokeAs[SecretStorage](i)
	sessionStorage := do.MustInvokeAs[SessionStorage](i)
	client := do.MustInvokeAs[SecretClient](i)
	return NewSecretUseCase(storage, client, sessionStorage)

}

func NewSecretUseCase(storage SecretStorage, client SecretClient, sessionStorage SessionStorage) *SecretUseCase {
	return &SecretUseCase{
		storage: storage,
		client:  client,
		sessionStorage: sessionStorage,
	}
}

func (uc *SecretUseCase) ListSecrets(ctx context.Context) ([]domain.Secret, error) {
	secrets, err := uc.storage.GetSecrets(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list secrets")
	}

	return secrets, nil
}

func (uc *SecretUseCase) GetSecret(ctx context.Context, id string) (domain.Secret, error) {
	secret, err := uc.storage.GetSecret(ctx, id)
	if err != nil {
		return domain.Secret{}, errors.Wrapf(err, "failed to get secret with id %s ", id)
	}

	return secret, nil
}

func (uc *SecretUseCase) AddSecret(ctx context.Context, secret domain.Secret) (domain.Secret, error) {
	session, err := uc.sessionStorage.GetSession(ctx)
	if err != nil {
		return domain.Secret{}, errors.Wrap(err, "missing session: try to login")
	}

	addedSecret, err := uc.client.AddSecret(ctx, secret, *session)
	if err != nil {
		return domain.Secret{}, errors.Wrapf(err, "failed to add secret %s", secret.ID)
	}

	err = uc.storage.SaveSecret(ctx, *addedSecret)
	if err != nil {
		return domain.Secret{}, errors.Wrapf(err, "failed to add secret %s to local storage", secret.ID)
	}


	return secret, nil
}

func (uc *SecretUseCase) DeleteSecret(ctx context.Context, id string) error {
	session, err := uc.sessionStorage.GetSession(ctx)
	if err != nil {
		return errors.Wrap(err, "missing session: try to login")
	}

	localSecret, err := uc.storage.GetSecret(ctx, id)
	if err != nil {
		return errors.Wrapf(err, "failed to delete secret %s in local storage", id)
	}

	err = uc.client.DeleteSecret(ctx, localSecret, *session)
	if err != nil {
		return errors.Wrapf(err, "failed to delete secret %s", id)
	}

	uc.storage.DeleteSecret(ctx, localSecret)

	return nil
}
