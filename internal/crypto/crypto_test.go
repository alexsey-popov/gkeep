package crypto

import (
	"bytes"
	"crypto/aes"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestService_EncryptDecrypt(t *testing.T) {
	service, err := NewWithTempKey()
	require.NoError(t, err)
	require.NotNil(t, service)

	plainText := []byte("my-secret-password-12345")

	encrypted, err := service.Encrypt(plainText)
	require.NoError(t, err)
	require.NotEmpty(t, encrypted)
	assert.NotEqual(t, plainText, encrypted)

	decrypted, err := service.Decrypt(encrypted)
	require.NoError(t, err)
	assert.Equal(t, plainText, decrypted)
}

func TestService_DecryptInvalid(t *testing.T) {
	service, err := NewWithTempKey()
	require.NoError(t, err)

	_, err = service.Decrypt([]byte("short"))
	assert.Error(t, err)

	_, err = service.Decrypt(make([]byte, 50))
	assert.Error(t, err)
}

func TestService_StreamEncryptDecrypt(t *testing.T) {
	service, err := NewWithTempKey()
	require.NoError(t, err)

	testCases := []struct {
		name string
		data []byte
	}{
		{
			name: "empty data",
			data: []byte(""),
		},
		{
			name: "small data",
			data: []byte("hello stream cipher"),
		},
		{
			name: "medium binary data",
			data: bytes.Repeat([]byte{0x01, 0x02, 0x03, 0x04, 0x05, 0xAA, 0xBB, 0xCC}, 5000), // 40 KB
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer

			// 1. Шифрование потока
			writer, err := service.NewEncryptWriter(&buf)
			require.NoError(t, err)

			// Пишем чанками
			chunkSize := 7
			for i := 0; i < len(tc.data); i += chunkSize {
				end := i + chunkSize
				if end > len(tc.data) {
					end = len(tc.data)
				}
				n, wErr := writer.Write(tc.data[i:end])
				require.NoError(t, wErr)
				assert.Equal(t, end-i, n)
			}

			rawEncrypted := buf.Bytes()
			assert.Equal(t, aes.BlockSize+len(tc.data), len(rawEncrypted), "Размер зашифрованных данных должен быть IV (16 байт) + размер данных")

			if len(tc.data) > 0 {
				assert.NotEqual(t, tc.data, rawEncrypted[aes.BlockSize:], "Зашифрованное тело не должно совпадать с исходными данными")
			}

			// 2. Расшифровывание потока
			reader, err := service.NewDecryptReader(bytes.NewReader(rawEncrypted))
			require.NoError(t, err)

			var decryptedBuf bytes.Buffer
			readBuf := make([]byte, 13)
			for {
				n, rErr := reader.Read(readBuf)
				if n > 0 {
					decryptedBuf.Write(readBuf[:n])
				}
				if rErr == io.EOF {
					break
				}
				require.NoError(t, rErr)
			}

			assert.True(t, bytes.Equal(tc.data, decryptedBuf.Bytes()), "расшифрованные данные должны совпадать с исходными")
		})
	}
}

func TestService_StreamDecryptInvalid(t *testing.T) {
	service, err := NewWithTempKey()
	require.NoError(t, err)

	// Передаем буфер короче 16 байт
	shortReader := bytes.NewReader([]byte("too-short"))
	_, err = service.NewDecryptReader(shortReader)
	require.Error(t, err)
}
