package domain

import "time"

// User представляет пользователя системы
type User struct {
	ID        string
	Login     string
	Password  string // В реальности будет хеш
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Key для получения User из контекста
type ContextKey string

const UserContextKey ContextKey = "user-context-key"
