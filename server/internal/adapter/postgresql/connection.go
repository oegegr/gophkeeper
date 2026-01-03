package postgresql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// NewPGConn создает новое подключение к PostgreSQL с пулом, проверяет его и применяет миграции
func NewPGConn(ctx context.Context, connectionString string) (*sql.DB, error) {
	log.Println("Creating PostgreSQL connection pool...")

	// 1. Создаем подключение с пулом
	db, err := createConnectionPool(connectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// 2. Проверяем подключение
	if err := pingConnection(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to verify connection: %w", err)
	}

	// 3. Применяем миграции
	// if err := applyMigrations(connectionString); err != nil {
	// 	db.Close()
	// 	return nil, fmt.Errorf("failed to apply migrations: %w", err)
	// }

	log.Println("PostgreSQL connection pool created successfully")
	return db, nil
}

// createConnectionPool создает пул подключений к базе данных
func createConnectionPool(connectionString string) (*sql.DB, error) {
	log.Printf("Connecting to PostgreSQL: %s", connectionString)

	db, err := sql.Open("pgx", connectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Настраиваем пул соединений
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(5 * time.Minute)

	log.Println("PostgreSQL connection pool configured")
	return db, nil
}

// pingConnection проверяет соединение с базой
func pingConnection(db *sql.DB) error {
	log.Println("Verifying PostgreSQL connection...")

	if err := db.Ping(); err != nil {
		return fmt.Errorf("ping failed: %w", err)
	}

	log.Println("PostgreSQL connection verified")
	return nil
}

// applyMigrations применяет миграции к базе данных
func applyMigrations(connectionString string) error {
	log.Println("Applying PostgreSQL migrations...")

	m, err := migrate.New("file://server/migrations", connectionString)
	if err != nil {
		log.Printf("failed to configure db migrations: %v", err)
		return fmt.Errorf("failed to create migration: %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Printf("failed to apply db migrations: %v", err)
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	log.Println("PostgreSQL migrations applied successfully")
	return nil
}
