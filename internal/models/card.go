// Модели
package models

import (
	"fmt"
	"time"

	"github.com/alexsey-popov/gkeep/internal/crypto"
)

// Card банковская карта
type Card struct {
	Id        string    `db:"id"`
	UserId    string    `db:"user_id"`
	Name      string    `db:"name" validate:"required,min=1,max=255" label:"Название"`
	Number    string    `db:"number" validate:"required,number,luhn_checksum" label:"Номер"`
	Expired   string    `db:"expired" validate:"required,number,len=4" label:"Срок"`
	Code      string    `db:"code" validate:"required,number,min=3,max=4" label:"Код"`
	CreatedAt time.Time `db:"created_at"`
}

// Encrypt Шифрование чувствительных данных
func (c *Card) Encrypt(cipher *crypto.Service) error {
	number, err := cipher.EncryptStr(c.Number)
	if err != nil {
		return err
	}

	expired, err := cipher.EncryptStr(c.Expired)
	if err != nil {
		return err
	}

	code, err := cipher.EncryptStr(c.Code)
	if err != nil {
		return err
	}

	c.Number = number
	c.Expired = expired
	c.Code = code

	return nil
}

// Decrypt Расшифровывание чувствительных данных
func (c *Card) Decrypt(cipher *crypto.Service) error {
	number, err := cipher.DecryptStr(c.Number)
	if err != nil {
		return fmt.Errorf("Ошибка при декодировании поля number: %w", err)
	}
	c.Number = number

	expired, err := cipher.DecryptStr(c.Expired)
	if err != nil {
		return fmt.Errorf("Ошибка при декодировании поля expired: %w", err)
	}
	c.Expired = expired

	code, err := cipher.DecryptStr(c.Code)
	if err != nil {
		return fmt.Errorf("Ошибка при декодировании поля code: %w", err)
	}
	c.Code = code

	return nil
}
