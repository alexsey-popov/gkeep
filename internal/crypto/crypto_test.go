package crypto

import (
	"bytes"
	"crypto/aes"
	"errors"
	"fmt"
	"io"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type failWriter struct{}

func (f *failWriter) Write(p []byte) (n int, err error) {
	return 0, errors.New("write failed")
}

func TestGenerateRandom(t *testing.T) {
	t.Run("negative - random длиной в 0 символов", func(t *testing.T) {
		b, err := generateRandom(0)
		require.NoError(t, err)
		assert.Empty(t, b)
	})

	t.Run("positive - generateRandom", func(t *testing.T) {
		b1, err := generateRandom(32)
		require.NoError(t, err)
		assert.Len(t, b1, 32)

		b2, err := generateRandom(32)
		require.NoError(t, err)
		assert.Len(t, b2, 32)

		assert.NotEqual(t, b1, b2, "последовательные генерации должны возвращать уникальные случайные байты")
	})
}

func TestNew_Keys(t *testing.T) {
	t.Run("positive - проверка длины символов", func(t *testing.T) {
		for _, size := range []int{16, 24, 32} {
			key, err := generateRandom(size)
			require.NoError(t, err)

			service, err := New(key)
			require.NoError(t, err)
			require.NotNil(t, service)
			assert.NotNil(t, service.block)
			assert.NotNil(t, service.cipher)
		}
	})

	t.Run("negative - некорректная длина ключа", func(t *testing.T) {
		invalidSizes := []int{0, 5, 10, 15, 31, 33, 64}
		for _, size := range invalidSizes {
			key := make([]byte, size)
			service, err := New(key)
			assert.Error(t, err, "ключ размером %d байт должен вызывать ошибку", size)
			assert.Nil(t, service)
		}
	})

	t.Run("positive - создание объекта для шифрования", func(t *testing.T) {
		service, err := NewWithTempKey()
		require.NoError(t, err)
		require.NotNil(t, service)
		assert.NotNil(t, service.block)
		assert.NotNil(t, service.cipher)
	})
}

func TestService_EncryptDecrypt(t *testing.T) {
	service, err := NewWithTempKey()
	require.NoError(t, err)
	require.NotNil(t, service)

	testCases := []struct {
		name string
		data []byte
	}{
		{name: "negative - empty bytes", data: []byte{}},
		{name: "positive - short string", data: []byte("my-secret-password-12345")},
		{name: "positive - unicode string", data: []byte("Секретные данные 🔒 Привет, мир!")},
		{name: "positive - large binary", data: bytes.Repeat([]byte{0xDE, 0xAD, 0xBE, 0xEF}, 1024)},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			encrypted, err := service.Encrypt(tc.data)
			require.NoError(t, err)
			require.NotEmpty(t, encrypted)

			if len(tc.data) > 0 {
				assert.NotEqual(t, tc.data, encrypted)
			}

			// Проверяем, что два вызова шифрования одних и тех же данных дают разный шифротекст (за счет случайного nonce)
			encrypted2, err := service.Encrypt(tc.data)
			require.NoError(t, err)
			assert.NotEqual(t, encrypted, encrypted2)

			decrypted, err := service.Decrypt(encrypted)
			require.NoError(t, err)
			assert.True(t, bytes.Equal(tc.data, decrypted), "расшифрованные байты должны совпадать с исходными")

			decrypted2, err := service.Decrypt(encrypted2)
			require.NoError(t, err)
			assert.True(t, bytes.Equal(tc.data, decrypted2), "расшифрованные байты должны совпадать с исходными")
		})
	}
}

func TestService_DecryptInvalid(t *testing.T) {
	service, err := NewWithTempKey()
	require.NoError(t, err)

	t.Run("negative - слишком короткий текст", func(t *testing.T) {
		_, err = service.Decrypt([]byte("short"))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "слишком короткие")
	})

	t.Run("negative - некорретный зашифрованный текст", func(t *testing.T) {
		validCiphertext, err := service.Encrypt([]byte("test data"))
		require.NoError(t, err)

		// Портим байт в шифротексте
		corrupted := make([]byte, len(validCiphertext))
		copy(corrupted, validCiphertext)
		corrupted[len(corrupted)-1] ^= 0xFF

		_, err = service.Decrypt(corrupted)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ошибка при декодировании")
	})

	t.Run("negative - ошибка при расшифровании", func(t *testing.T) {
		randomBytes, err := generateRandom(50)
		require.NoError(t, err)

		_, err = service.Decrypt(randomBytes)
		assert.Error(t, err)
	})
}

func TestService_EncryptStrDecryptStr(t *testing.T) {
	service, err := NewWithTempKey()
	require.NoError(t, err)

	testCases := []struct {
		name string
		data string
	}{
		{name: "negative - пустая строка", data: ""},
		{name: "positive - simple text", data: "login:password123"},
		{name: "positive - unicode and special characters", data: "Текстовая заметка с эмодзи 🔑💳 и спецсимволами !@#$%^&*()_+"},
		{name: "positive - multi-line text", data: "line 1\nline 2\nline 3\n\tindented"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			encHex, err := service.EncryptStr(tc.data)
			require.NoError(t, err)
			require.NotEmpty(t, encHex)

			// Результат должен быть валидной hex-строкой
			assert.Regexp(t, "^[0-9a-f]+$", encHex)

			decrypted, err := service.DecryptStr(encHex)
			require.NoError(t, err)
			assert.Equal(t, tc.data, decrypted)
		})
	}
}

func TestService_DecryptStr_Invalid(t *testing.T) {
	service, err := NewWithTempKey()
	require.NoError(t, err)

	t.Run("negative - некорректный hex", func(t *testing.T) {
		_, err := service.DecryptStr("not-a-valid-hex-string!@#")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Ошибка при декодировании hex поля")
	})

	t.Run("negative - слишком короткий hex", func(t *testing.T) {
		// 6 байт в hex (12 hex символов), при NonceSize = 12 байт (24 hex символа)
		_, err := service.DecryptStr("010203040506")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "слишком короткие")
	})

	t.Run("negative - некорректный hex", func(t *testing.T) {
		encHex, err := service.EncryptStr("secret message")
		require.NoError(t, err)

		// Портим последний символ hex-строки
		corruptedHex := encHex[:len(encHex)-1]
		if encHex[len(encHex)-1] == '0' {
			corruptedHex += "1"
		} else {
			corruptedHex += "0"
		}

		_, err = service.DecryptStr(corruptedHex)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ошибка при декодировании")
	})
}

func TestService_StreamEncryptDecrypt(t *testing.T) {
	service, err := NewWithTempKey()
	require.NoError(t, err)

	testCases := []struct {
		name string
		data []byte
	}{
		{
			name: "negative - пустые данные",
			data: []byte(""),
		},
		{
			name: "negative - короткие данные",
			data: []byte("hello stream cipher"),
		},
		{
			name: "positive - большой объем данных",
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
				if errors.Is(rErr, io.EOF) {
					break
				}
				require.NoError(t, rErr)
			}

			assert.True(t, bytes.Equal(tc.data, decryptedBuf.Bytes()), "расшифрованные данные должны совпадать с исходными")
		})
	}
}

func TestService_StreamEncrypt_WriteError(t *testing.T) {
	service, err := NewWithTempKey()
	require.NoError(t, err)

	_, err = service.NewEncryptWriter(&failWriter{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ошибка при записи вектора инициализации")
}

func TestService_StreamDecryptInvalid(t *testing.T) {
	service, err := NewWithTempKey()
	require.NoError(t, err)

	// Передаем буфер короче 16 байт
	shortReader := bytes.NewReader([]byte("too-short"))
	_, err = service.NewDecryptReader(shortReader)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "ошибка при чтении вектора инициализации")
}

func TestService_Concurrency(t *testing.T) {
	service, err := NewWithTempKey()
	require.NoError(t, err)

	var wg sync.WaitGroup
	workers := 50
	iterations := 20

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				// 1. Байтовое шифрование / расшифрование
				msgBytes := []byte(fmt.Sprintf("worker-%d-iter-%d-data", workerID, j))
				encBytes, eErr := service.Encrypt(msgBytes)
				assert.NoError(t, eErr)
				decBytes, dErr := service.Decrypt(encBytes)
				assert.NoError(t, dErr)
				assert.Equal(t, msgBytes, decBytes)

				// 2. Строковое шифрование / расшифрование
				msgStr := fmt.Sprintf("str-worker-%d-iter-%d-data-🔒", workerID, j)
				encStr, sErr := service.EncryptStr(msgStr)
				assert.NoError(t, sErr)
				decStr, sdErr := service.DecryptStr(encStr)
				assert.NoError(t, sdErr)
				assert.Equal(t, msgStr, decStr)
			}
		}(i)
	}

	wg.Wait()
}
