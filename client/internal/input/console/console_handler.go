package console

import (
	"context"
	"fmt"

	"github.com/oegegr/gophkeeper/client/internal/domain"
	"github.com/samber/do/v2"
	"github.com/urfave/cli/v2"
)

type SessionUseCase interface {
	Login(ctx context.Context, username, password string) error
	Register(ctx context.Context, username, password string) error
}

type SecretUseCase interface {
	ListSecrets(ctx context.Context) ([]domain.Secret, error)
	GetSecret(ctx context.Context, id string) (domain.Secret, error)
	AddSecret(ctx context.Context, secret domain.Secret) (domain.Secret, error)
	DeleteSecret(ctx context.Context, id string) error
}

type SyncUseCase interface {
	Sync(ctx context.Context) error
}

func ResolveConsoleHandler(i do.Injector) *ConsoleHandler {
	secretUC := do.MustInvokeAs[SecretUseCase](i)
	syncUC := do.MustInvokeAs[SyncUseCase](i)
	sessionUC := do.MustInvokeAs[SessionUseCase](i)
	return NewConsoleHandler(sessionUC, secretUC, syncUC)
}

// ConsoleHandler обрабатывает команды
type ConsoleHandler struct {
	sessionUC SessionUseCase
	secretUC  SecretUseCase
	syncUC    SyncUseCase
}

// NewConsoleHandler создает новый обработчик
func NewConsoleHandler(
	userUC SessionUseCase,
	secretUC SecretUseCase,
	syncUC SyncUseCase,
) *ConsoleHandler {
	return &ConsoleHandler{
		sessionUC: userUC,
		secretUC:  secretUC,
		syncUC:    syncUC,
	}
}

// HandleLogin обрабатывает вход
func (h *ConsoleHandler) HandleLogin(ctx *cli.Context) error {
	username := ctx.String("username")
	password := ctx.String("password")

	err := h.sessionUC.Login(context.Background(), username, password)
	if err != nil {
		return fmt.Errorf("login failed: %w", err)
	}

	fmt.Printf("login completed: %s\n", username)
	return nil
}

// HandleRegister обрабатывает регистрацию
func (h *ConsoleHandler) HandleRegister(ctx *cli.Context) error {
	username := ctx.String("username")
	password := ctx.String("password")

	err := h.sessionUC.Register(context.Background(), username, password)
	if err != nil {
		return fmt.Errorf("registration failed: %w", err)
	}

	fmt.Printf("registration completed: %s\n", username)
	return nil
}

// HandleListSecrets обрабатывает список секретов
func (h *ConsoleHandler) HandleListSecrets(ctx *cli.Context) error {
	secrets, err := h.secretUC.ListSecrets(context.Background())
	if err != nil {
		return handleAuthError(err)
	}

	ShowSecretsList(secrets)
	return nil
}

// HandleGetSecret обрабатывает получение секрета
func (h *ConsoleHandler) HandleGetSecret(ctx *cli.Context) error {
	secretID := ctx.String("id")
	secret, err := h.secretUC.GetSecret(ctx.Context, secretID)
	if err != nil {
		return handleAuthError(err)
	}

	showSecretDetails(secret)
	return nil
}

// HandleAddSecret обрабатывает добавление секрета
func (h *ConsoleHandler) HandleAddSecret(ctx *cli.Context) error {
	name := ctx.String("name")
	data := ctx.String("data")
	secretType := ctx.String("type")

	secret := domain.Secret{
		Type: secretType,
		Data: []byte(data),
		Meta: name,
	}

	_, err := h.secretUC.AddSecret(ctx.Context, secret)
	if err != nil {
		return handleAuthError(err)
	}

	fmt.Printf("secret added: %s\n", name)
	return nil
}

// HandleDeleteSecret обрабатывает удаление секрета
func (h *ConsoleHandler) HandleDeleteSecret(ctx *cli.Context) error {
	secretID := ctx.String("id")

	if err := h.secretUC.DeleteSecret(ctx.Context, secretID); err != nil {
		return handleAuthError(err)
	}

	fmt.Println("secret deleted")
	return nil
}

// HandleForceSync обрабатывает синхронизацию
func (h *ConsoleHandler) HandleForceSync(ctx *cli.Context) error {
	if err := h.syncUC.Sync(ctx.Context); err != nil {
		return handleAuthError(err)
	}

	fmt.Println("sync completed")
	return nil
}

// handleAuthError обрабатывает ошибки аутентификации
func handleAuthError(err error) error {
	if err == domain.ErrUnauthenticated {
		return fmt.Errorf("%v. Please run 'login' command first", err)
	}
	return fmt.Errorf("error: %w", err)
}

func showSecretDetails(secret domain.Secret) {
	fmt.Println("\n=== Детали секрета ===")
	fmt.Printf("ID: %s\n", secret.ID)
	fmt.Printf("Тип: %s\n", secret.Type)
	fmt.Printf("Данные: %s\n", string(secret.Data))
	if !secret.CreatedAt.IsZero() {
		fmt.Printf("Создан: %s\n", secret.CreatedAt.Format("2006-01-02 15:04:05"))
	}
	if !secret.UpdatedAt.IsZero() {
		fmt.Printf("Обновлен: %s\n", secret.UpdatedAt.Format("2006-01-02 15:04:05"))
	}
	if secret.Meta != "" {
		fmt.Printf("Метаданные: %s\n", secret.Meta)
	}
}
