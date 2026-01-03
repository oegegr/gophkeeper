package config

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config содержит конфигурацию сервера
type Config struct {
	Addr             string        // Адрес сервера (например: ":50051")
	EnableReflection bool          // Включить gRPC reflection
	ShutdownTimeout  time.Duration // Таймаут graceful shutdown
	Dsa              string        // database connection string
	TokenSecret      string        // secret pass frase for token
}

// Load загружает конфигурацию из различных источников
func Load() (*Config, error) {
	// Инициализируем Viper
	v := viper.New()

	// Настраиваем Viper
	setupViper(v)

	// Читаем конфигурацию
	if err := v.ReadInConfig(); err != nil {
		// Если файл конфигурации не найден, это не ошибка
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config: %w", err)
		}
		log.Println("Config file not found, using defaults and environment variables")
	}

	// Создаем конфиг со значениями из Viper
	cfg := &Config{
		Addr:             v.GetString("addr"),
		EnableReflection: v.GetBool("enable_reflection"),
		ShutdownTimeout:  v.GetDuration("shutdown_timeout"),
		Dsa:              v.GetString("dsa"),
		TokenSecret:      v.GetString("token_secret"),
	}

	// Валидируем конфигурацию
	if err := validateConfig(cfg); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	log.Printf("Configuration loaded successfully: %v", cfg)
	return cfg, nil
}

// setupViper настраивает Viper для чтения конфигурации
func setupViper(v *viper.Viper) {
	// Настройки по умолчанию
	v.SetDefault("addr", "127.0.0.1:8080")
	v.SetDefault("enable_reflection", true)
	v.SetDefault("shutdown_timeout", 10*time.Second)
	v.SetDefault("dsa", "postgres://admin:admin@127.0.0.1:5432/gophkeeper?sslmode=disable")
	v.SetDefault("token_secret", "12345")

	// Имя файла конфигурации (без расширения)
	v.SetConfigName("config")

	// Поддерживаемые форматы конфигурационных файлов
	v.SetConfigType("yaml")

	// Пути поиска файла конфигурации
	v.AddConfigPath(".")

	// Чтение из environment variables
	v.SetEnvPrefix("GOPHKEEPER")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Связываем env переменные с полями конфига
	// GOPHKEEPER_ADDR -> addr
	// GOPHKEEPER_ENABLE_REFLECTION -> enable_reflection
	// GOPHKEEPER_SHUTDOWN_TIMEOUT -> shutdown_timeout
	// GOPHKEEPER_DSA -> dsa
}

// validateConfig валидирует конфигурацию
func validateConfig(cfg *Config) error {
	if cfg.Addr == "" {
		return fmt.Errorf("server address is required")
	}

	if cfg.ShutdownTimeout <= 0 {
		cfg.ShutdownTimeout = 10 * time.Second
	}

	return nil
}

// DefaultConfig возвращает конфигурацию по умолчанию
func DefaultConfig() *Config {
	return &Config{
		Addr:             ":50051",
		Dsa:              "",
		EnableReflection: true,
		ShutdownTimeout:  10 * time.Second,
	}
}
