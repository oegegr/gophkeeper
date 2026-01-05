package usecases

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/oegegr/gophkeeper/server/internal/domain"
	"github.com/samber/do/v2"
)

type SecretsRepository interface {
	FindSecretByID(ctx context.Context, secretID string) (domain.Secret, error)
	FindSecretsByUserID(ctx context.Context, userID string) ([]domain.Secret, error)
	DeleteSecret(ctx context.Context, secretID string) error
	CreateSecret(ctx context.Context, secret domain.Secret) error
	UpdateSecret(ctx context.Context, secret domain.Secret) error
}

// SecretsUseCaseImpl реализация SecretsUseCase
type SecretsUseCaseImpl struct {
	repos SecretsRepository
}

func ResolveSecretsUseCase(i do.Injector) *SecretsUseCaseImpl {
	repo := do.MustInvokeAs[SecretsRepository](i)

	return NewSecretsUseCase(repo)
}

// NewSecretsUseCase создает новый SecretsUseCase
func NewSecretsUseCase(repos SecretsRepository) *SecretsUseCaseImpl {
	return &SecretsUseCaseImpl{
		repos: repos,
	}
}

// GetSecrets возвращает все секреты пользователя
func (uc *SecretsUseCaseImpl) GetSecrets(ctx context.Context, userID string) ([]domain.Secret, error) {
	// Получаем секреты
	secrets, err := uc.repos.FindSecretsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return secrets, nil
}

// GetSecret возвращает конкретный секрет
func (uc *SecretsUseCaseImpl) GetSecret(ctx context.Context, userID, secretID string) (domain.Secret, error) {
	// Получаем секрет
	secret, err := uc.repos.FindSecretByID(ctx, secretID)
	if err != nil {
		return domain.Secret{}, err
	}

	// Проверяем права доступа
	if secret.UserID != userID {
		return domain.Secret{}, domain.ErrPermissionDenied
	}

	return secret, nil
}

// AddSecret добавляет новый секрет
func (uc *SecretsUseCaseImpl) AddSecret(ctx context.Context, userID string, secretType domain.SecretType, data []byte, meta string) (domain.Secret, error) {
	// Создаем секрет
	secret := domain.Secret{
		ID:        uuid.NewString(),
		UserID:    userID,
		Type:      secretType,
		Data:      data,
		Meta:      meta,
		Version:   1,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Сохраняем
	if err := uc.repos.CreateSecret(ctx, secret); err != nil {
		return domain.Secret{}, err
	}

	return secret, nil
}

// UpdateSecret обновляет существующий секрет
func (uc *SecretsUseCaseImpl) UpdateSecret(ctx context.Context, userID, secretID string, secretType domain.SecretType, data []byte, meta string, version int) (domain.Secret, error) {
	// Получаем текущий секрет
	secret, err := uc.repos.FindSecretByID(ctx, secretID)
	if err != nil {
		return domain.Secret{}, err
	}

	// Проверяем права доступа
	if secret.UserID != userID {
		return domain.Secret{}, domain.ErrPermissionDenied
	}

	// Проверяем версию
	if secret.Version != version {
		return domain.Secret{}, domain.ErrVersionConflict
	}

	// Обновляем данные
	secret.Type = secretType
	secret.Data = data
	secret.Meta = meta
	secret.Version = version + 1
	secret.UpdatedAt = time.Now()

	// Сохраняем
	if err := uc.repos.UpdateSecret(ctx, secret); err != nil {
		return domain.Secret{}, err
	}

	return secret, nil
}

// DeleteSecret удаляет секрет
func (uc *SecretsUseCaseImpl) DeleteSecret(ctx context.Context, userID, secretID string) error {
	// Валидация
	if userID == "" || secretID == "" {
		return domain.ErrInvalidInput
	}

	// Получаем секрет
	secret, err := uc.repos.FindSecretByID(ctx, secretID)
	if err != nil {
		return err
	}

	// Проверяем права доступа
	if secret.UserID != userID {
		return domain.ErrPermissionDenied
	}

	// Удаляем
	return uc.repos.DeleteSecret(ctx, secretID)
}
