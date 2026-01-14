package usecases

import (
	"context"

	"github.com/oegegr/gophkeeper/client/internal/domain"
	"github.com/oegegr/gophkeeper/client/internal/input/console"
	"github.com/pkg/errors"
	"github.com/samber/do/v2"
)

var _ console.SessionUseCase = (*SessionUseCase)(nil)

type SessionStorage interface {
	SaveSession(ctx context.Context, session domain.Session) error
	GetSession(ctx context.Context) (*domain.Session, error)
}

type SessionClient interface {
	Login(ctx context.Context, username, password string) (domain.Session, error)
	Register(ctx context.Context, username, password string) (domain.Session, error)
}

type SessionUseCase struct {
	client  SessionClient
	storage SessionStorage
}

func ResolveSessionUseCase(i do.Injector) *SessionUseCase {
	client := do.MustInvokeAs[SessionClient](i)
	storage := do.MustInvokeAs[SessionStorage](i)
	return NewSessionUseCase(client, storage)

}

func NewSessionUseCase(client SessionClient, storage SessionStorage) *SessionUseCase {
	return &SessionUseCase{
		client:  client,
		storage: storage,
	}
}

func (uc *SessionUseCase) Login(ctx context.Context, username, password string) error {
	session, err := uc.client.Login(ctx, username, password)
	if err != nil {
		return errors.Wrap(err, "failed to login")
	}
	uc.storage.SaveSession(ctx, session)
	return nil
}

func (uc *SessionUseCase) Register(ctx context.Context, username, password string) error {
	session, err := uc.client.Register(ctx, username, password)
	if err != nil {
		return errors.Wrap(err, "failed to register user")
	}
	uc.storage.SaveSession(ctx, session)
	return nil
}
