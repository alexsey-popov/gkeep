// Модели
package models

import "time"

// Text текстовая информация
type Text struct {
	Id        string    `db:"id"`
	UserId    string    `db:"user_id"`
	Name      string    `db:"name" validate:"required,min=1,max=255" label:"Название"`
	Text      string    `db:"text" validate:"required,min=1,max=10000" label:"Текст"`
	CreatedAt time.Time `db:"created_at"`
}
