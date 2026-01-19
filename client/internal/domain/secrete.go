package domain

import (
	"time"
)

// Secret представляет секретную запись
type Secret struct {
    ID          string         `json:"id"`
    UserID      string         `json:"user_id"`
    Type        string         `json:"type"`
    Data        []byte         `json:"data"`      // Зашифрованные данные
    Meta        string         `json:"meta"`      // Метаинформация
    Version     int32          `json:"version"`
    CreatedAt   time.Time      `json:"created_at"`
    UpdatedAt   time.Time      `json:"updated_at"`
}
