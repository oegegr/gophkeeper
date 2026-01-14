package internal

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/oegegr/gophkeeper/client/internal/config"
	"github.com/oegegr/gophkeeper/client/internal/input/console"
	"github.com/samber/do/v2"
	"github.com/urfave/cli/v2"
)

type App struct {
	config  *config.Config
	ctx     context.Context
	console *cli.App
	cancel  context.CancelFunc
}

// New создает новое приложение
func New(cfg *config.Config) (*App, error) {
	// Инициализируем DI
	if err := InitDependencies(cfg); err != nil {
		return nil, fmt.Errorf("failed to init dependencies: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	// Создаем консольное приложение
	consoleHandler := do.MustInvoke[*console.ConsoleHandler](di)
	cliApp, err := console.New(consoleHandler)
	if err != nil {
		cancel()
		ShutdownDependencies(ctx)
		return nil, fmt.Errorf("failed to create CLI app: %w", err)
	}

	app := &App{
		config:  cfg,
		ctx:     ctx,
		console: cliApp,
		cancel:  cancel,
	}

	// Настраиваем graceful shutdown
	setupGracefulShutdown(app)

	return app, nil
}

// Run запускает приложение
func (a *App) Run(ctx context.Context) error {
	// Запускаем CLI приложение
	return a.console.RunContext(ctx, os.Args)
}

// Stop останавливает приложение
func (a *App) Stop(ctx context.Context) error {
	if a.cancel != nil {
		a.cancel()
	}
	
	// Завершаем зависимости
	if err := ShutdownDependencies(ctx); err != nil {
		return fmt.Errorf("failed to shutdown dependencies: %w", err)
	}
	
	return nil
}

// setupGracefulShutdown настраивает graceful shutdown
func setupGracefulShutdown(app *App) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\nShutting down gracefully...")
		
		// Создаем контекст с таймаутом для graceful shutdown
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		
		// Останавливаем приложение
		if err := app.Stop(shutdownCtx); err != nil {
			fmt.Printf("Error during shutdown: %v\n", err)
		}
		
		fmt.Println("Shutdown completed")
		os.Exit(0)
	}()
}
