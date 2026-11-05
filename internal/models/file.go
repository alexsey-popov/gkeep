// Модели
package models

import "time"

// File текстовая информация
type File struct {
	Id        string    `db:"id"`
	UserId    string    `db:"user_id"`
	Name      string    `db:"name"`
	File      string    `db:"file"`
	CreatedAt time.Time `db:"created_at"`
}
