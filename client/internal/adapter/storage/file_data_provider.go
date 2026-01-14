package storage

import (
    "os"
    "path/filepath"
)

// FileDataProvider - чтение/запись в файл
type FileDataProvider struct {
    path string
}

func NewFileDataProvider(path string) *FileDataProvider {
    return &FileDataProvider{path: path}
}

func (p *FileDataProvider) Load() ([]byte, error) {
    data, err := os.ReadFile(p.path)
    if err != nil {
        if os.IsNotExist(err) {
            // Новый файл - пустой JSON
            return []byte("{}"), nil
        }
        return nil, err
    }
    
    if len(data) == 0 {
        return []byte("{}"), nil
    }
    
    return data, nil
}

func (p *FileDataProvider) Save(data []byte) error {
    // Создаем директорию
    dir := filepath.Dir(p.path)
    if err := os.MkdirAll(dir, 0700); err != nil {
        return err
    }
    
    // Атомарная запись через временный файл
    tmpPath := p.path + ".tmp"
    if err := os.WriteFile(tmpPath, data, 0600); err != nil {
        return err
    }
    
    return os.Rename(tmpPath, p.path)
}