package storage

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/oegegr/gophkeeper/client/internal/domain"
	"github.com/oegegr/gophkeeper/client/internal/usecases"
	"github.com/pkg/errors"
	"github.com/samber/do/v2"
)


var _ usecases.SecretStorage = (*JsonStorage)(nil)
var _ usecases.SyncStorage = (*JsonStorage)(nil)
var _ usecases.SessionStorage = (*JsonStorage)(nil)

// StorageData - все данные хранилища
type StorageData struct {
	Session  *domain.Session          `json:"session,omitempty"`
	Secrets  map[string]domain.Secret `json:"secrets"` // ID -> Secret
	Metadata map[string]string        `json:"metadata"`
}

// JsonStorage - основное хранилище с сериализацией JSON
type JsonStorage struct {
	provider DataProvider
	data     *StorageData
}

func ResolveJsonStorage(i do.Injector) *JsonStorage {
	provider := do.MustInvokeAs[DataProvider](i)
	storage, err := NewJsonStorage(provider)
	if err != nil {
		panic(errors.Wrap(err, "failed to create storage"))
	}
	return storage
}

// NewJsonStorage создает новое хранилище
func NewJsonStorage(provider DataProvider) (*JsonStorage, error) {
	storage := &JsonStorage{
		provider: provider,
	}

	// Загружаем и десериализуем данные
	if err := storage.load(); err != nil {
		return nil, errors.Wrap(err, "failed to load data")
	}

	return storage, nil
}

func (s *JsonStorage) SaveSession(ctx context.Context, session domain.Session) error {
    s.data.Session = &session 
    return s.save()
}

// GetSession возвращает сессию
func (s *JsonStorage) GetSession(ctx context.Context) (*domain.Session, error) {

    if s.data.Session == nil {
        return nil, errors.New("failed to get session")
    }
	s.save()
    return s.data.Session, nil
}

// ClearSession очищает сессию
func (s *JsonStorage) ClearSession(ctx context.Context) error {
    s.data.Session = nil
    return s.save()
}

func (s *JsonStorage) SaveSecret(ctx context.Context, secret domain.Secret) error {
	if secret.CreatedAt.IsZero() {
		secret.CreatedAt = time.Now()
	}
	secret.UpdatedAt = time.Now()

	// Шифрование данных перед сохранением
	secret.Data = encodeBase64(secret.Data)

	s.data.Secrets[secret.ID] = secret
	return s.save()
}

func (s *JsonStorage) SaveSecrets(ctx context.Context, secrets []domain.Secret) error {
	for _, secret := range secrets {
		if secret.CreatedAt.IsZero() {
			secret.CreatedAt = time.Now()
		}
		secret.UpdatedAt = time.Now()

		// Шифрование данных перед сохранением
		secret.Data = encodeBase64(secret.Data)

		s.data.Secrets[secret.ID] = secret
	}
	return s.save()
}

func (s *JsonStorage) GetSecret(ctx context.Context, id string) (domain.Secret, error) {
	secret, exists := s.data.Secrets[id]
	if !exists {
		return domain.Secret{}, domain.ErrNotFound
	}

	// Расшифровка данных после загрузки
	decodedData := decodeBase64(secret.Data)
	secret.Data = decodedData

	return secret, nil
}

func (s *JsonStorage) GetUserSecrets(ctx context.Context, userID string) ([]*domain.Secret, error) {
	var secrets []*domain.Secret
	for _, secret := range s.data.Secrets {
		if secret.UserID == userID {
			s := secret
			secrets = append(secrets, &s)
		}
	}
	return secrets, nil
}

func (s *JsonStorage) GetSecrets(ctx context.Context) ([]domain.Secret, error) {
	var secrets []domain.Secret
	for _, secret := range s.data.Secrets {
		decodedData := decodeBase64(secret.Data)
		secret.Data = decodedData
		s := secret
		secrets = append(secrets, s)
	}
	return secrets, nil
}

func (s *JsonStorage) DeleteSecrets(ctx context.Context, secrets []domain.Secret) error {
	for _, secret := range secrets {
		secret, exists := s.data.Secrets[secret.ID]
		if !exists {
			return domain.ErrNotFound
		}

		secret.UpdatedAt = time.Now()
		s.data.Secrets[secret.ID] = secret
	}

	return s.save()
}

func (s *JsonStorage) DeleteSecret(ctx context.Context, secret domain.Secret) error {
	secret, exists := s.data.Secrets[secret.ID]
	if !exists {
		return domain.ErrNotFound
	}

	secret.UpdatedAt = time.Now()
	s.data.Secrets[secret.ID] = secret

	return s.save()
}

func (s *JsonStorage) UpdateSecret(ctx context.Context, secret *domain.Secret) error {
	if _, exists := s.data.Secrets[secret.ID]; !exists {
		return domain.ErrNotFound
	}

	// Шифрование данных перед сохранением
	secret.Data = encodeBase64(secret.Data)

	secret.UpdatedAt = time.Now()
	s.data.Secrets[secret.ID] = *secret

	return s.save()
}

func (s *JsonStorage) Close() error {
	return s.save()
}

// --- Внутренние методы для работы с данными ---

func (s *JsonStorage) load() error {
	// Загружаем сырые данные
	rawData, err := s.provider.Load()
	if err != nil {
		return err
	}

	// Десериализуем JSON
	var storageData StorageData
	if err := json.Unmarshal(rawData, &storageData); err != nil {
		// Если JSON невалидный, создаем пустые данные
		storageData = StorageData{
			Secrets:  make(map[string]domain.Secret),
			Metadata: make(map[string]string),
		}
	}

	// Инициализируем если нужно
	if storageData.Secrets == nil {
		storageData.Secrets = make(map[string]domain.Secret)
	}
	if storageData.Metadata == nil {
		storageData.Metadata = make(map[string]string)
	}

	s.data = &storageData

	return nil
}

func (s *JsonStorage) save() error {
	// Обновляем метаданные
	s.updateMetadata("last_save", time.Now().Format(time.RFC3339))

	// Сериализуем в JSON
	jsonData, err := json.MarshalIndent(s.data, "", "  ")
	if err != nil {
		return fmt.Errorf("ошибка сериализации: %v", err)
	}

	// Сохраняем через провайдер
	return s.provider.Save(jsonData)
}

func (s *JsonStorage) updateMetadata(key, value string) {
	if s.data.Metadata == nil {
		s.data.Metadata = make(map[string]string)
	}
	s.data.Metadata[key] = value
}

func encodeBase64(data []byte) []byte {
	encoded := make([]byte, base64.StdEncoding.EncodedLen(len(data)))
	base64.StdEncoding.Encode(encoded, data)
	return encoded
}

func decodeBase64(data []byte) []byte {
	decoded := make([]byte, base64.StdEncoding.DecodedLen(len(data)))
	_, err := base64.StdEncoding.Decode(decoded, data)
	if err != nil {
		// Обработка ошибки
	}
	return decoded
}
