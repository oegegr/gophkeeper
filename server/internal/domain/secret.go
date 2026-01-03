package domain

import "time"

// SecretType тип секретных данных
type SecretType string

const (
	SecretTypeLoginPassword SecretType = "login_password"
	SecretTypeText          SecretType = "text"
	SecretTypeBinary        SecretType = "binary"
	SecretTypeCard          SecretType = "card"
)

// Secret представляет секретные данные пользователя
type Secret struct {
	ID        string
	UserID    string
	Type      SecretType
	Data      []byte // Зашифрованные данные
	Meta      string // Метаинформация
	Version   int
	CreatedAt time.Time
	UpdatedAt time.Time
}

// SecretData представляет данные для разных типов секретов
type SecretData interface{}

type LoginPasswordData struct {
	Login    string `json:"login"`
	Password string `json:"password"`
	URL      string `json:"url,omitempty"`
}

type TextData struct {
	Text string `json:"text"`
}

type BinaryData struct {
	Data []byte `json:"data"`
	Name string `json:"name,omitempty"`
}

type CardData struct {
	Number     string `json:"number"`
	Holder     string `json:"holder"`
	ExpiryDate string `json:"expiry_date"`
	CVV        string `json:"cvv,omitempty"`
}
