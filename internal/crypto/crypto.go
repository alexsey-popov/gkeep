// Шифрование/Расшифровывание данных секретных данных в БД
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"sync"
)

// Service сервис для шифрования/расшифровывания данных
type Service struct {
	block  cipher.Block
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
		block:  aesblock,
		cipher: aesgcm,
	}, nil
}

// NewEncryptWriter создает io.Writer, который шифрует записываемые данные потоковым алгоритмом AES-CTR.
// В начало целевого writer записывается 16-байтный вектор инициализации (IV).
func (s *Service) NewEncryptWriter(w io.Writer) (io.Writer, error) {
	iv, err := generateRandom(aes.BlockSize)
	if err != nil {
		return nil, fmt.Errorf("ошибка при создании вектора инициализации: %w", err)
	}

	if _, err := w.Write(iv); err != nil {
		return nil, fmt.Errorf("ошибка при записи вектора инициализации: %w", err)
	}

	stream := cipher.NewCTR(s.block, iv)
	return &cipher.StreamWriter{
		S: stream,
		W: w,
	}, nil
}

// NewDecryptReader создает io.Reader, который расшифровывает данные потоковым алгоритмом AES-CTR.
// Из начала целевого reader считывается 16-байтный вектор инициализации (IV).
func (s *Service) NewDecryptReader(r io.Reader) (io.Reader, error) {
	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(r, iv); err != nil {
		return nil, fmt.Errorf("ошибка при чтении вектора инициализации: %w", err)
	}

	stream := cipher.NewCTR(s.block, iv)
	return &cipher.StreamReader{
		S: stream,
		R: r,
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

// EncryptStr шифрование данных c с выводом в строку
func (s *Service) EncryptStr(src string) (string, error) {
	bytes, err := s.Encrypt([]byte(src))
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", bytes), nil
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

// Decrypt расшифровывание данных
func (s *Service) DecryptStr(data string) (string, error) {
	// Декодируем hex строку
	dataBytes, err := hex.DecodeString(data)
	if err != nil {
		return "", fmt.Errorf("Ошибка при декодировании hex поля: %w", err)
	}

	nonceSize := s.cipher.NonceSize()
	if len(dataBytes) < nonceSize {
		return "", fmt.Errorf("зашифрованные данные слишком короткие")
	}

	nonce := dataBytes[:nonceSize]
	ciphertext := dataBytes[nonceSize:]

	// Защищаем aesgcm т.к. он не является потокобезопасным
	s.mu.Lock()
	defer s.mu.Unlock()

	src, err := s.cipher.Open(nil, nonce, ciphertext, nil) // расшифровываем
	if err != nil {
		return "", fmt.Errorf("ошибка при декодировании: %w", err)
	}

	return string(src), nil
}
