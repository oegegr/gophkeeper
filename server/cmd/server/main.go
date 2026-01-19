package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/oegegr/gophkeeper/server/internal"
	"github.com/oegegr/gophkeeper/server/internal/config"
)

// Переменные для хранения информации о сборке
// Заполняются при сборке через ldflags
var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

func main() {
	printBuildInfo()
	// Загружаем конфигурацию
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Создаем приложение
	app := internal.NewApplication(cfg)

	// Запускаем сервер
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := app.Start(ctx); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}

	log.Printf("Server started on %s", cfg.Addr)

	// Ждем сигнала завершения
	select {
	case <-ctx.Done():
		// Получен сигнал завершения
		log.Println("Shutdown signal received")
	case <-app.Wait():
		// Сервер сам завершился с ошибкой
		log.Println("Server stopped unexpectedly")
	}

	// Останавливаем сервер
	if err := app.Stop(ctx); err != nil {
		log.Printf("Error during shutdown: %v", err)
	}

}

// Функция для вывода информации о сборке
func printBuildInfo() {
	if buildVersion == "" {
		buildVersion = "N/A"
	}
	if buildDate == "" {
		buildDate = "N/A"
	}
	if buildCommit == "" {
		buildCommit = "N/A"
	}

	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)
}
