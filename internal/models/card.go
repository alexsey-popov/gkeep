// Модели
package models

import "time"

// Card банковская карта
type Card struct {
	Id        string    `db:"id"`
	UserId    string    `db:"user_id"`
	Name      string    `db:"name"`
	Number    string    `db:"number"`
	Expired   string    `db:"expired"`
	Code      string    `db:"code"`
	CreatedAt time.Time `db:"created_at"`
}
