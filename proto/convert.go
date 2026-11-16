package proto

import (
	"github.com/alexsey-popov/gkeep/internal/models"
	"github.com/alexsey-popov/gkeep/proto/card"
	"github.com/alexsey-popov/gkeep/proto/file"
	"github.com/alexsey-popov/gkeep/proto/password"
	"github.com/alexsey-popov/gkeep/proto/text"
)

// ConvertFile преобразует модель File в protobuf сообщение File
func ConvertFile(item *models.File) *file.File {
	if item == nil {
		return nil
	}

	return file.File_builder{
		Id:       &item.Id,
		Name:     &item.Name,
		Filename: &item.Filename,
		Size:     &item.Size,
		Checksum: &item.Checksum,
	}.Build()
}

// ConvertText преобразует модель Text в protobuf сообщение Text
func ConvertText(item *models.Text) *text.Text {
	if item == nil {
		return nil
	}

	return text.Text_builder{
		Id:   &item.Id,
		Name: &item.Name,
		Text: &item.Text,
	}.Build()
}

// ConvertCard преобразует модель Card в protobuf сообщение Card
func ConvertCard(item *models.Card) *card.Card {
	if item == nil {
		return nil
	}

	return card.Card_builder{
		Id:      &item.Id,
		Name:    &item.Name,
		Number:  &item.Number,
		Expired: &item.Expired,
		Code:    &item.Code,
	}.Build()
}

// ConvertPassword преобразует модель Password в protobuf сообщение Password
func ConvertPassword(item *models.Password) *password.Password {
	if item == nil {
		return nil
	}

	return password.Password_builder{
		Id:       &item.Id,
		Name:     &item.Name,
		Login:    &item.Login,
		Password: &item.Password,
	}.Build()
}
