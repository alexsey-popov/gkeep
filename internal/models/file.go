// Модели
package models

import (
	"fmt"
	"time"

	"github.com/alexsey-popov/gkeep/internal/crypto"
)

// File Файл
type File struct {
	Id          string    `json:"id" db:"id"`
	UserId      string    `json:"user_id" db:"user_id"`
	Name        string    `json:"name" db:"name" validate:"required,min=1,max=255" label:"Название"`
	Filename    string    `json:"filename" db:"filename" validate:"required,min=1,max=255" label:"Название файла"`
	StoragePath string    `json:"storage_path" db:"storage_path" label:"Путь до файла"`
	Size        int64     `json:"size" db:"size" label:"Размер файла"`
	Checksum    string    `json:"checksum" db:"checksum"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// Encrypt Шифрование чувствительных данных
func (f *File) Encrypt(cipher *crypto.Service) error {
	filename, err := cipher.EncryptStr(f.Filename)
	if err != nil {
		return err
	}

	path, err := cipher.EncryptStr(f.StoragePath)
	if err != nil {
		return err
	}

	f.Filename = filename
	f.StoragePath = path

	return nil
}

// Decrypt Расшифровывание чувствительных данных
func (f *File) Decrypt(cipher *crypto.Service) error {
	filename, err := cipher.DecryptStr(f.Filename)
	if err != nil {
		return fmt.Errorf("ошибка при декодировании поля: %w", err)
	}
	f.Filename = filename

	path, err := cipher.DecryptStr(f.StoragePath)
	if err != nil {
		return fmt.Errorf("ошибка при декодировании поля: %w", err)
	}
	f.StoragePath = path

	return nil
}
