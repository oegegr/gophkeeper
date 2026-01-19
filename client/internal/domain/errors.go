package domain

import (
	"errors"
)

// Ошибки хранилища
var (
	ErrNotFound        = errors.New("запись не найдена")
	ErrAlreadyExists   = errors.New("запись уже существует")
	ErrInvalidPassword = errors.New("неверный пароль")
	ErrNotInitialized  = errors.New("хранилище не инициализировано")
	ErrUnauthenticated = errors.New("unauthenticated - please login first")
)
