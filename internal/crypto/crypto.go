// Шифрование/Расшифровывание данных секретных данных в БД
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"sync"
)

// Service сервис для шифрования/расшифровывания данных
type Service struct {
	cipher cipher.AEAD
	mu     sync.Mutex
}

// generateRandom набор случайных символов указанной длины
func generateRandom(size int) ([]byte, error) {
	// генерируем криптостойкие случайные байты в b
	b := make([]byte, size)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}

	return b, nil
}

// NewWithTempKey Создание сервиса с временным ключом шифрования
func NewWithTempKey() (*Service, error) {
	// будем использовать AES-256, создав ключ длиной 32 байта
	key, err := generateRandom(2 * aes.BlockSize) // ключ шифрования
	if err != nil {
		return nil, fmt.Errorf("ошибка при создании ключа: %w", err)
	}

	return New(key)
}

// New создание сервиса
func New(key []byte) (*Service, error) {
	// NewCipher создает блок шифрования
	aesblock, err := aes.NewCipher(key)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return nil, fmt.Errorf("ошибка при создании cipher.Block: %w", err)
	}

	// NewGCM создаёт блочный шифр
	aesgcm, err := cipher.NewGCM(aesblock)
	if err != nil {
		return nil, fmt.Errorf("ошибка при создании GCM: %w", err)
	}

	return &Service{
		cipher: aesgcm,
	}, nil
}

// Encrypt шифрование данных
func (s *Service) Encrypt(src []byte) ([]byte, error) {
	// создаём вектор инициализации
	nonce, err := generateRandom(s.cipher.NonceSize())
	if err != nil {
		return nil, fmt.Errorf("ошибка при создании вектора инициализации: %w", err)
	}

	// Защищаем aesgcm т.к. он не является потокобезопасным
	s.mu.Lock()
	defer s.mu.Unlock()

	dst := s.cipher.Seal(nonce, nonce, src, nil)

	return dst, nil
}

// Decrypt расшифровывание данных
func (s *Service) Decrypt(dst []byte) ([]byte, error) {
	nonceSize := s.cipher.NonceSize()
	if len(dst) < nonceSize {
		return nil, fmt.Errorf("зашифрованные данные слишком короткие")
	}

	nonce := dst[:nonceSize]
	ciphertext := dst[nonceSize:]

	// Защищаем aesgcm т.к. он не является потокобезопасным
	s.mu.Lock()
	defer s.mu.Unlock()

	src, err := s.cipher.Open(nil, nonce, ciphertext, nil) // расшифровываем
	if err != nil {
		return nil, fmt.Errorf("ошибка при декодировании: %w", err)
	}

	return src, nil
}
