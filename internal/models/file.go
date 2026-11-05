// Модели
package models

import "time"

// File Файл
type File struct {
	Id          string    `db:"id"`
	UserId      string    `db:"user_id"`
	Name        string    `db:"name"`
	Filename    string    `db:"filename"`
	StoragePath string    `db:"storage_path"`
	Size        int64     `db:"size"`
	Checksum    string    `db:"checksum"`
	CreatedAt   time.Time `db:"created_at"`
}
