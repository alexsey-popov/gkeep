// Модели
package models

import (
	"time"

	"github.com/alexsey-popov/gkeep/internal/crypto"
)

// File Файл
type File struct {
	Id          string    `db:"id"`
	UserId      string    `db:"user_id"`
	Name        string    `db:"name" validate:"required,min=1,max=255" label:"Название"`
	Filename    string    `db:"filename" validate:"required,min=1,max=255" label:"Название файла"`
	StoragePath string    `db:"storage_path" validate:"required,min=1,max=255" label:"Путь до файла"`
	Size        int64     `db:"size" validate:"required,min=1,max=255" label:"Размер файла"`
	Checksum    string    `db:"checksum"`
	CreatedAt   time.Time `db:"created_at"`
}

// Encrypt Шифрование чувствительных данных
func (f *File) Encrypt(cipher *crypto.Service) error {
	return nil
}

// Decrypt Расшифровывание чувствительных данных
func (p *File) Decrypt(cipher *crypto.Service) error {
	return nil
}
