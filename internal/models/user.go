// Модели
package models

// User пользователь
type User struct {
	Id       string `db:"id"`
	Login    string `db:"login" validate:"required,min=4,max=255" label:"Логин"`
	Password string `db:"password" validate:"required,min=4,max=255" label:"Пароль"`
}
