package grpc

import (
	"encoding/json"
	"fmt"
	"time"

	pb "github.com/oegegr/gophkeeper/proto"
	"github.com/oegegr/gophkeeper/server/internal/domain"
)

// Mapper преобразует между доменными моделями и protobuf
type Mapper struct{}

// NewMapper создает новый маппер
func NewMapper() *Mapper {
	return &Mapper{}
}

// ToProtoUser преобразует доменного User в protobuf User (если нужно)
func (m *Mapper) ToProtoUser(user *domain.User) *pb.User {
	if user == nil {
		return nil
	}

	return &pb.User{
		Id:        user.ID,
		Login:     user.Login,
		CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

// ToProtoSecret преобразует доменный Secret в protobuf Secret
func (m *Mapper) ToProtoSecret(secret domain.Secret) (*pb.Secret, error) {
	// Определяем тип protobuf
	var dataType pb.DataType
	switch secret.Type {
	case domain.SecretTypeLoginPassword:
		dataType = pb.DataType_LOGIN_PASSWORD
	case domain.SecretTypeText:
		dataType = pb.DataType_TEXT
	case domain.SecretTypeBinary:
		dataType = pb.DataType_BINARY
	case domain.SecretTypeCard:
		dataType = pb.DataType_CARD
	default:
		return nil, fmt.Errorf("unknown secret type: %s", secret.Type)
	}

	return &pb.Secret{
		Id:        secret.ID,
		Type:      dataType,
		Data:      secret.Data,
		Meta:      secret.Meta,
		Version:   int32(secret.Version),
		UpdatedAt: secret.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}

// ToProtoSecrets преобразует список доменных секретов
func (m *Mapper) ToProtoSecrets(secrets []domain.Secret) ([]*pb.Secret, error) {
	protoSecrets := make([]*pb.Secret, 0, len(secrets))

	for _, secret := range secrets {
		protoSecret, err := m.ToProtoSecret(secret)
		if err != nil {
			return nil, err
		}
		protoSecrets = append(protoSecrets, protoSecret)
	}

	return protoSecrets, nil
}

// FromProtoSecret преобразует protobuf Secret в доменный Secret
func (m *Mapper) FromProtoSecret(protoSecret *pb.Secret, userID string) (domain.Secret, error) {
	if protoSecret == nil {
		return domain.Secret{}, nil
	}

	// Определяем доменный тип
	var secretType domain.SecretType
	switch protoSecret.GetType() {
	case pb.DataType_LOGIN_PASSWORD:
		secretType = domain.SecretTypeLoginPassword
	case pb.DataType_TEXT:
		secretType = domain.SecretTypeText
	case pb.DataType_BINARY:
		secretType = domain.SecretTypeBinary
	case pb.DataType_CARD:
		secretType = domain.SecretTypeCard
	default:
		return domain.Secret{}, fmt.Errorf("unknown proto secret type: %v", protoSecret.GetType())
	}

	// Парсим время
	updatedAt, err := parseTime(protoSecret.GetUpdatedAt())
	if err != nil {
		return domain.Secret{}, err
	}

	return domain.Secret{
		ID:        protoSecret.GetId(),
		UserID:    userID,
		Type:      secretType,
		Data:      protoSecret.GetData(),
		Meta:      protoSecret.GetMeta(),
		Version:   int(protoSecret.GetVersion()),
		UpdatedAt: updatedAt,
	}, nil
}

// FromProtoSecrets преобразует список protobuf секретов
func (m *Mapper) FromProtoSecrets(protoSecrets []*pb.Secret, userID string) ([]domain.Secret, error) {
	secrets := make([]domain.Secret, 0, len(protoSecrets))

	for _, protoSecret := range protoSecrets {
		secret, err := m.FromProtoSecret(protoSecret, userID)
		if err != nil {
			return nil, err
		}
		secrets = append(secrets, secret)
	}

	return secrets, nil
}

// ToProtoRegisterResponse создает ответ регистрации
func (m *Mapper) ToProtoRegisterResponse(user *domain.User, token string) *pb.RegisterResponse {
	return &pb.RegisterResponse{
		UserId: user.ID,
		Token:  token,
	}
}

// ToProtoLoginResponse создает ответ аутентификации
func (m *Mapper) ToProtoLoginResponse(user *domain.User, token string) *pb.LoginResponse {
	return &pb.LoginResponse{
		UserId: user.ID,
		Token:  token,
	}
}

// DecodeLoginPasswordData декодирует данные логин/пароль
func (m *Mapper) DecodeLoginPasswordData(data []byte) (*domain.LoginPasswordData, error) {
	var loginData domain.LoginPasswordData
	if err := json.Unmarshal(data, &loginData); err != nil {
		return nil, err
	}
	return &loginData, nil
}

// EncodeLoginPasswordData кодирует данные логин/пароль
func (m *Mapper) EncodeLoginPasswordData(data *domain.LoginPasswordData) ([]byte, error) {
	return json.Marshal(data)
}

// parseTime парсит время из строки
func parseTime(timeStr string) (time.Time, error) {
	if timeStr == "" {
		return time.Time{}, nil
	}
	return time.Parse("2006-01-02T15:04:05Z07:00", timeStr)
}
