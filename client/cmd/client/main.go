package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/oegegr/gophkeeper/client/internal"
	"github.com/oegegr/gophkeeper/client/internal/config"
)

var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

func main() {
	// printBuildInfo()
	// Загружаем конфигурацию из стандартных мест
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Создаем приложение
	app, err := internal.New(cfg)
	if err != nil {
		log.Fatalf("Failed to create app: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Запускаем приложение
	if err := app.Run(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// printBuildInfo выводит информацию о сборке
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

	fmt.Printf("GophKeeper Client\n")
	fmt.Printf("Version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)
}