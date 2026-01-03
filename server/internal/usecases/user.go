package usecases

import (
	"context"
	"time"

	"github.com/oegegr/gophkeeper/server/internal/domain"
	"github.com/oegegr/gophkeeper/server/internal/input/grpc"
	"github.com/samber/do/v2"
)

var _ grpc.AuthUseCase = (*AuthUseCaseImpl)(nil)

type AuthRepository interface {
	FindByLogin(ctx context.Context, login string) (*domain.User, error)
	CreateUser(ctx context.Context, user *domain.User) error
}

// TokenProvider управляет токенами
type TokenProvider interface {
	GenerateToken(userID string) (string, error)
}

// AuthUseCaseImpl реализация AuthUseCase
type AuthUseCaseImpl struct {
	repos        AuthRepository
	tokenManager TokenProvider
}

func ResolveAuthUseCase(i do.Injector) *AuthUseCaseImpl {
	repo := do.MustInvokeAs[AuthRepository](i)
	tokenManager := do.MustInvokeAs[TokenProvider](i)

	return NewAuthUseCase(repo, tokenManager)
}

// NewAuthUseCase создает новый AuthUseCase
func NewAuthUseCase(repos AuthRepository, tokenManager TokenProvider) *AuthUseCaseImpl {
	return &AuthUseCaseImpl{
		repos:        repos,
		tokenManager: tokenManager,
	}
}

// Register регистрирует нового пользователя
func (uc *AuthUseCaseImpl) Register(ctx context.Context, login, password string) (*domain.User, string, error) {
	// Валидация
	if login == "" || password == "" {
		return nil, "", domain.ErrInvalidInput
	}

	// Проверяем, нет ли уже такого пользователя
	existing, _ := uc.repos.FindByLogin(ctx, login)
	if existing != nil {
		return nil, "", domain.ErrAlreadyExists
	}

	// Создаем пользователя
	user := &domain.User{
		Login:     login,
		Password:  password, // В реальности здесь был бы хеш
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Сохраняем
	if err := uc.repos.CreateUser(ctx, user); err != nil {
		return nil, "", err
	}

	// Генерируем токен
	token, err := uc.tokenManager.GenerateToken(user.ID)
	if err != nil {
		return nil, "", err
	}

	return user, token, nil
}

// Login выполняет аутентификацию пользователя
func (uc *AuthUseCaseImpl) Login(ctx context.Context, login, password string) (*domain.User, string, error) {
	// Валидация
	if login == "" || password == "" {
		return nil, "", domain.ErrInvalidInput
	}

	// Ищем пользователя
	user, err := uc.repos.FindByLogin(ctx, login)
	if err != nil {
		return nil, "", domain.ErrNotFound
	}

	// Проверяем пароль (в реальности сравниваем хеши)
	if user.Password != password {
		return nil, "", domain.ErrInvalidCredentials
	}

	// Генерируем токен
	token, err := uc.tokenManager.GenerateToken(user.ID)
	if err != nil {
		return nil, "", err
	}

	return user, token, nil
}
