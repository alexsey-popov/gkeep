// Модели
package models

import (
	"fmt"
	"time"

	"github.com/alexsey-popov/gkeep/internal/crypto"
)

// Password связка логин/пароль
type Password struct {
	Id        string    `db:"id"`
	UserId    string    `db:"user_id" validate:"required,uuid" label:"ID пользователя"`
	Name      string    `db:"name" validate:"required,min=2,max=255" label:"Название"`
	Login     string    `db:"login" validate:"required,min=2,max=255" label:"Логин"`
	Password  string    `db:"password" validate:"required,min=2,max=255" label:"Пароль"`
	CreatedAt time.Time `db:"created_at"`
}

// Encrypt Шифрование чувствительных данных
func (p *Password) Encrypt(cipher *crypto.Service) error {
	loginEncrypted, err := cipher.EncryptStr(p.Login)
	if err != nil {
		return err
	}

	passwordEncrypted, err := cipher.EncryptStr(p.Password)
	if err != nil {
		return err
	}

	p.Login = loginEncrypted
	p.Password = passwordEncrypted

	return nil
}

// Decrypt Расшифровывание чувствительных данных
func (p *Password) Decrypt(cipher *crypto.Service) error {
	login, err := cipher.DecryptStr(p.Login)
	if err != nil {
		return fmt.Errorf("Ошибка при декодировании поля login: %w", err)
	}
	p.Login = login

	pass, err := cipher.DecryptStr(p.Password)
	if err != nil {
		return fmt.Errorf("Ошибка при декодировании поля password: %w", err)
	}
	p.Password = pass

	return nil
}
