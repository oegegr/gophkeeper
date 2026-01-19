package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	ServerAddress    string `mapstructure:"server_address"`
	StoragePath      string `mapstructure:"storage_path"`
	StorageEncryption bool   `mapstructure:"storage_encryption"`
}

// Load загружает конфигурацию из файла и env переменных
func Load() (*Config, error) {
	v := viper.New()
	
	// Устанавливаем значения по умолчанию
	v.SetDefault("server_address", "localhost:8080")
	v.SetDefault("storage_path", getDefaultStoragePath())
	v.SetDefault("storage_encryption", true)
	
	// Настраиваем viper для работы с файлами
	v.SetConfigName("config") // имя конфиг файла без расширения
	v.SetConfigType("yaml")   // формат конфига
	
	// Стандартные пути поиска конфига
	v.AddConfigPath(".")                    // текущая директория
	v.AddConfigPath("$HOME/.gophkeeper")   // домашняя директория
	
	// Читаем конфигурационный файл
	if err := v.ReadInConfig(); err != nil {
		// Если файл не найден, используем значения по умолчанию
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			// Другие ошибки чтения файла
			return nil, fmt.Errorf("failed to read config: %w", err)
		}
		// Файл не найден - это нормально, используем значения по умолчанию
	}
	
	// Читаем переменные окружения (с префиксом GOPHKEEPER_)
	v.SetEnvPrefix("GOPHKEEPER_CLIENT")
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	
	// Десериализуем в структуру
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}
	
	return &cfg, nil
}

// getDefaultStoragePath возвращает путь к хранилищу по умолчанию
func getDefaultStoragePath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "./gophkeeper_data.json"
	}
	
	dir := filepath.Join(home, ".gophkeeper")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "./gophkeeper_data.json"
	}
	
	return filepath.Join(dir, "data.json")
}