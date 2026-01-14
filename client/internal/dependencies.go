package internal

import (
	"context"

	"github.com/oegegr/gophkeeper/client/internal/adapter/client"
	"github.com/oegegr/gophkeeper/client/internal/adapter/storage"
	"github.com/oegegr/gophkeeper/client/internal/config"
	"github.com/oegegr/gophkeeper/client/internal/input/console"
	"github.com/oegegr/gophkeeper/client/internal/usecases"
	"github.com/samber/do/v2"
)

var di do.Injector

func init() {
	di = do.New()
}

// InitDependencies инициализирует все зависимости
func InitDependencies(cfg *config.Config) error {
	// Конфигурация
	do.Provide(di, func(i do.Injector) (*config.Config, error) {
		return cfg, nil
	})

	// DataProvider с учетом шифрования
	do.Provide(di, func(i do.Injector) (storage.DataProvider, error) {
		cfg := do.MustInvoke[*config.Config](i)

		// Создаем базовый файловый провайдер
		fileProvider := storage.NewFileDataProvider(cfg.StoragePath)

		// Если шифрование включено, оборачиваем в EncryptedDataProvider
		if cfg.StorageEncryption {
			// Мастер-пароль пока пустой, будет установлен при логине
			return storage.NewEncryptedDataProvider(fileProvider, ""), nil
		}

		// Иначе используем обычный файловый провайдер
		return fileProvider, nil
	})

	do.ProvideValue(di, storage.ResolveJsonStorage(di))
	do.ProvideValue(di, client.ResolveClient(di))
	do.ProvideValue(di, usecases.ResolveSessionUseCase(di))
	do.ProvideValue(di, usecases.ResolveSecretUseCase(di))
	do.ProvideValue(di, usecases.ResolveSyncUseCase(di))
	do.ProvideValue(di, console.ResolveConsoleHandler(di))

	return nil
}

// GetInjector возвращает инжектор зависимостей
func GetInjector() do.Injector {
	return di
}

// ShutdownDependencies завершает работу зависимостей
func ShutdownDependencies(ctx context.Context) error {
	return di.Shutdown()
}
