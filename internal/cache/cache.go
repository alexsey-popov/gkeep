package cache

import (
	"crypto/sha256"
	"encoding/json/v2"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"sync"

	"github.com/alexsey-popov/gkeep/internal/crypto"
	"github.com/alexsey-popov/gkeep/internal/models"
)

var (
	ErrOpenFile  = errors.New("ошибка при открытии файла кеша")
	ErrDecrypt   = errors.New("ошибка при расшифровывании данных")
	ErrUnmarshal = errors.New("ошибка при десериализации данных")
)

// Cache структура для управления кешем
type Cache struct {
	filepath string
	key      []byte
	cipher   *crypto.Service
	mu       sync.Mutex
	data     *CacheData
	log      *slog.Logger
}

// CacheData структура для хранения всех закешированных данных пользователя
type CacheData struct {
	Passwords map[int][]models.Password `json:"passwords"`
	Cards     map[int][]models.Card     `json:"cards"`
	Texts     map[int][]models.Text     `json:"texts"`
	Files     map[int][]models.File     `json:"files"`
}

// GetKey генерирует ключ шифрования
func GetKey(args ...string) []byte {
	key := sha256.Sum256([]byte("gkeep-client-cache-key:" + strings.Join(args, ":")))
	return key[:]
}

// New создание нового объекта для работы с кешем
func New(filepath string, key []byte) (*Cache, error) {
	// Если передана
	if strings.TrimSpace(filepath) == "" {
		filepath = ".gkeep_cache"
	}

	// создаём объект шифрования
	cipher, err := crypto.New(key)
	if err != nil {
		return nil, fmt.Errorf("ошибка при создании объекта шифрования: %w", err)
	}

	return &Cache{
		filepath: filepath,
		key:      key,
		cipher:   cipher,
		data:     newCacheData(),
	}, nil
}

// newCacheData создание нового объекта для хранения кеша
func newCacheData() *CacheData {
	return &CacheData{
		Passwords: make(map[int][]models.Password),
		Cards:     make(map[int][]models.Card),
		Texts:     make(map[int][]models.Text),
		Files:     make(map[int][]models.File),
	}
}

// Save шифрует и сохраняет данные кэша в файл
func (c *Cache) Save() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	jsonData, err := json.Marshal(c.data)
	if err != nil {
		return fmt.Errorf("ошибка при сериализации данных: %w", err)
	}

	encrypted, err := c.cipher.Encrypt(jsonData)
	if err != nil {
		return fmt.Errorf("ошибка при шифровании данных: %w", err)
	}

	if err := os.WriteFile(c.filepath, encrypted, 0600); err != nil {
		return fmt.Errorf("ошибка при записи файла: %w", err)
	}

	return nil
}

// Load загружает и расшифровывает данные кэша из файла
func (c *Cache) Load() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Читаем данные из файла
	data, err := os.ReadFile(c.filepath)
	if err != nil {
		// Если файла с кешем нет - значит расшифровывать нечего, ошибкой это тоже не является
		if os.IsNotExist(err) {
			return nil
		}

		// Это клиент, логировать некуда, поэтому просто отдаём общее сообщение без логирования
		return ErrOpenFile
	}

	// Если файл пустой, то значит расшифровывать там нечего
	if len(data) == 0 {
		return nil
	}

	// Расшифровываем данные файла
	decrypted, err := c.cipher.Decrypt(data)
	if err != nil {
		return ErrDecrypt
	}

	// Десериализуем данные
	var cache CacheData
	if err = json.Unmarshal(decrypted, &cache); err != nil {
		return ErrUnmarshal
	}

	return nil
}
