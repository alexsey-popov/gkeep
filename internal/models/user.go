// Модели
package models

// User пользователь
type User struct {
	Id       string `db:"id"`
	Login    string `db:"login"`
	Password string `db:"password"`
}
