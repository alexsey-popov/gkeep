package models

import (
	"testing"
	"time"

	"github.com/alexsey-popov/gkeep/internal/crypto"
	"github.com/alexsey-popov/gkeep/internal/validator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCard_Encrypt_Decrypt(t *testing.T) {
	cipher, err := crypto.NewWithTempKey()
	require.NoError(t, err)

	t.Run("success encrypt and decrypt roundtrip", func(t *testing.T) {
		card := &Card{
			Id:        "card-1",
			UserId:    "user-1",
			Name:      "Visa Classic",
			Number:    "4532015112830366",
			Expired:   "1228",
			Code:      "777",
			CreatedAt: time.Now(),
		}

		origNumber := card.Number
		origExpired := card.Expired
		origCode := card.Code

		err := card.Encrypt(cipher)
		require.NoError(t, err)

		assert.NotEqual(t, origNumber, card.Number)
		assert.NotEqual(t, origExpired, card.Expired)
		assert.NotEqual(t, origCode, card.Code)

		err = card.Decrypt(cipher)
		require.NoError(t, err)

		assert.Equal(t, origNumber, card.Number)
		assert.Equal(t, origExpired, card.Expired)
		assert.Equal(t, origCode, card.Code)
	})

	t.Run("decrypt error on invalid number", func(t *testing.T) {
		card := &Card{
			Number:  "invalid-hex",
			Expired: "1228",
			Code:    "777",
		}

		err := card.Decrypt(cipher)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Ошибка при декодировании поля number")
	})

	t.Run("decrypt error on invalid expired", func(t *testing.T) {
		encNumber, err := cipher.EncryptStr("4532015112830366")
		require.NoError(t, err)

		card := &Card{
			Number:  encNumber,
			Expired: "invalid-hex",
			Code:    "777",
		}

		err = card.Decrypt(cipher)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Ошибка при декодировании поля expired")
	})

	t.Run("decrypt error on invalid code", func(t *testing.T) {
		encNumber, err := cipher.EncryptStr("4532015112830366")
		require.NoError(t, err)
		encExpired, err := cipher.EncryptStr("1228")
		require.NoError(t, err)

		card := &Card{
			Number:  encNumber,
			Expired: encExpired,
			Code:    "invalid-hex",
		}

		err = card.Decrypt(cipher)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Ошибка при декодировании поля code")
	})
}

func TestCard_Validation(t *testing.T) {
	v, err := validator.New()
	require.NoError(t, err)

	validCard := func() Card {
		return Card{
			Id:      "card-1",
			UserId:  "user-1",
			Name:    "Visa Classic",
			Number:  "4532015112830366", // Valid Luhn checksum
			Expired: "1228",
			Code:    "777",
		}
	}

	t.Run("valid card", func(t *testing.T) {
		card := validCard()
		err := v.Validate(card)
		assert.NoError(t, err)
	})

	t.Run("empty name", func(t *testing.T) {
		card := validCard()
		card.Name = ""
		err := v.Validate(card)
		require.Error(t, err)
		isErr, errCheck := v.ErrorIs(err, "Название", "required")
		require.NoError(t, errCheck)
		assert.True(t, isErr)
	})

	t.Run("invalid card number - bad luhn", func(t *testing.T) {
		card := validCard()
		card.Number = "4532015112830367" // Invalid Luhn checksum
		err := v.Validate(card)
		require.Error(t, err)
		isErr, errCheck := v.ErrorIs(err, "Номер", "luhn_checksum")
		require.NoError(t, errCheck)
		assert.True(t, isErr)
	})

	t.Run("invalid card number - not a number", func(t *testing.T) {
		card := validCard()
		card.Number = "453201511283036a"
		err := v.Validate(card)
		require.Error(t, err)
	})

	t.Run("empty card number", func(t *testing.T) {
		card := validCard()
		card.Number = ""
		err := v.Validate(card)
		require.Error(t, err)
		isErr, errCheck := v.ErrorIs(err, "Номер", "required")
		require.NoError(t, errCheck)
		assert.True(t, isErr)
	})

	t.Run("invalid expired length", func(t *testing.T) {
		card := validCard()
		card.Expired = "122"
		err := v.Validate(card)
		require.Error(t, err)
		isErr, errCheck := v.ErrorIs(err, "Срок", "len")
		require.NoError(t, errCheck)
		assert.True(t, isErr)
	})

	t.Run("invalid expired - non number", func(t *testing.T) {
		card := validCard()
		card.Expired = "12ab"
		err := v.Validate(card)
		require.Error(t, err)
		isErr, errCheck := v.ErrorIs(err, "Срок", "number")
		require.NoError(t, errCheck)
		assert.True(t, isErr)
	})

	t.Run("invalid code length - too short", func(t *testing.T) {
		card := validCard()
		card.Code = "12"
		err := v.Validate(card)
		require.Error(t, err)
		isErr, errCheck := v.ErrorIs(err, "Код", "min")
		require.NoError(t, errCheck)
		assert.True(t, isErr)
	})

	t.Run("invalid code length - too long", func(t *testing.T) {
		card := validCard()
		card.Code = "12345"
		err := v.Validate(card)
		require.Error(t, err)
		isErr, errCheck := v.ErrorIs(err, "Код", "max")
		require.NoError(t, errCheck)
		assert.True(t, isErr)
	})

	t.Run("invalid code - non number", func(t *testing.T) {
		card := validCard()
		card.Code = "abc"
		err := v.Validate(card)
		require.Error(t, err)
		isErr, errCheck := v.ErrorIs(err, "Код", "number")
		require.NoError(t, errCheck)
		assert.True(t, isErr)
	})
}
