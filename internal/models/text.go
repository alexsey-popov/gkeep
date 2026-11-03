// Модели
package models

import "time"

// Text текстовая информация
type Text struct {
	Id        string    `db:"id"`
	UserId    string    `db:"user_id"`
	Name      string    `db:"name"`
	Text      string    `db:"text"`
	CreatedAt time.Time `db:"created_at"`
}
