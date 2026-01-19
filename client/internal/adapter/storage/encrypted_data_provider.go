package storage

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"io"

	"github.com/pkg/errors"
)

// EncryptedDataProvider - добавляет шифрование к другому провайдеру
type EncryptedDataProvider struct {
    inner    DataProvider
    password string
}

func NewEncryptedDataProvider(inner DataProvider, password string) *EncryptedDataProvider {
    return &EncryptedDataProvider{
        inner:    inner,
        password: password,
    }
}

func (p *EncryptedDataProvider) Load() ([]byte, error) {
    // Загружаем данные
    data, err := p.inner.Load()
    if err != nil {
        return nil, err
    }
    
    // Если пусто или это новый файл - не шифруем
    if len(data) == 0 || string(data) == "{}" {
        return data, nil
    }
    
    // Пробуем дешифровать
    decrypted, err := decrypt(data, p.password)
    if err != nil {
        return nil, errors.Wrap(err, "failed to decrypt data")
    }
    
    return decrypted, nil
}

func (p *EncryptedDataProvider) Save(data []byte) error {
    // Шифруем
    encrypted, err := encrypt(data, p.password)
    if err != nil {
        return err
    }
    
    // Сохраняем через внутренний провайдер
    return p.inner.Save(encrypted)
}

// --- Шифрование ---

func deriveKey(password string) []byte {
    hash := sha256.Sum256([]byte(password))
    return hash[:]
}

func encrypt(data []byte, password string) ([]byte, error) {
    key := deriveKey(password)
    
    block, err := aes.NewCipher(key)
    if err != nil {
        return nil, err
    }
    
    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return nil, err
    }
    
    nonce := make([]byte, gcm.NonceSize())
    if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
        return nil, err
    }
    
    return gcm.Seal(nonce, nonce, data, nil), nil
}

func decrypt(data []byte, password string) ([]byte, error) {
    key := deriveKey(password)
    
    block, err := aes.NewCipher(key)
    if err != nil {
        return nil, err
    }
    
    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return nil, err
    }
    
    nonceSize := gcm.NonceSize()
    if len(data) < nonceSize {
        return nil, err
    }
    
    nonce, ciphertext := data[:nonceSize], data[nonceSize:]
    return gcm.Open(nil, nonce, ciphertext, nil)
}