// Модели
package models

import (
	"fmt"
	"time"

	"github.com/alexsey-popov/gkeep/internal/crypto"
)

// Text текстовая информация
type Text struct {
	Id        string    `db:"id"`
	UserId    string    `db:"user_id"`
	Name      string    `db:"name" validate:"required,min=1,max=255" label:"Название"`
	Text      string    `db:"text" validate:"required,min=1,max=10000" label:"Текст"`
	CreatedAt time.Time `db:"created_at"`
}

// Encrypt Шифрование чувствительных данных
func (t *Text) Encrypt(cipher *crypto.Service) error {
	text, err := cipher.EncryptStr(t.Text)
	if err != nil {
		return err
	}

	t.Text = text

	return nil
}

// Decrypt Расшифровывание чувствительных данных
func (t *Text) Decrypt(cipher *crypto.Service) error {
	text, err := cipher.DecryptStr(t.Text)
	if err != nil {
		return fmt.Errorf("Ошибка при декодировании поля: %w", err)
	}
	t.Text = text

	return nil
}
