// Модели
package models

import "time"

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
