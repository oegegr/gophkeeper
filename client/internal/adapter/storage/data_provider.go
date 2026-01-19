package storage

// DataProvider интерфейс для работы с данными
type DataProvider interface {
    Load() ([]byte, error)
    Save(data []byte) error
}