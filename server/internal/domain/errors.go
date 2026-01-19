package domain

import "errors"

// Общие ошибки доменного уровня
var (
	ErrNotFound           = errors.New("not found")
	ErrAlreadyExists      = errors.New("already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidInput       = errors.New("invalid input")
	ErrPermissionDenied   = errors.New("permission denied")
	ErrVersionConflict    = errors.New("version conflict")
	ErrEncryptionFailed   = errors.New("encryption failed")
	ErrDecryptionFailed   = errors.New("decryption failed")
	ErrUserNotFound       = errors.New("user not found")
)
