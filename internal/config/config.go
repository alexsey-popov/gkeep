// Конфигурация сервисов.
package config

import (
	"flag"
	"fmt"
)

const (
	// Адрес сервера (gRPC)
	FlagAddress    = "a"
	EnvAddress     = "ADDRESS"
	DefaultAddress = "localhost:50051"

	// Адрес подключения к базе данных
	FlagDatabaseURI    = "d"
	EnvDatabaseURI     = "DATABASE_URI"
	DefaultDatabaseURI = ""

	// Токен проверки аутентификации
	FlagJwtToken    = "j"
	EnvJwtToken     = "JWT_TOKEN"
	DefaultJwtToken = "secret"
)

// Server - Структура конфигурации для сервера
type Server struct {
	Address     string
	DatabaseURI string
	JwtToken    string
}

// Client - Структура конфигурации для клиента
type Client struct {
	Address string
}

// EnvLookupFunc заглушка для `os.LookupEnv`
type EnvLookupFunc func(string) (string, bool)

// getEnvString Получение строковых данных из env или выставление значения по умолчанию
func getEnvString(lookupFunc EnvLookupFunc, name string, defaultValue string) string {
	if value, ok := lookupFunc(name); ok {
		return value
	}

	return defaultValue
}

// NewServer парсинг конфига сервера (приоритет по убыванию: флаги, env, default)
func NewServer(args []string, lookupFunc EnvLookupFunc) (*Server, error) {
	cfg := &Server{}

	fs := flag.NewFlagSet("app", flag.ContinueOnError)

	// Получаем данные из env или их дефолтные значения
	userAddress := getEnvString(lookupFunc, EnvAddress, DefaultAddress)
	databaseURI := getEnvString(lookupFunc, EnvDatabaseURI, DefaultDatabaseURI)
	jwtToken := getEnvString(lookupFunc, EnvJwtToken, DefaultJwtToken)

	// Получаем данные из флагов
	fs.StringVar(&cfg.Address, FlagAddress, userAddress, "Адрес системы программы лояльности")
	fs.StringVar(&cfg.DatabaseURI, FlagDatabaseURI, databaseURI, "Адрес подключения к базе данных")
	fs.StringVar(&cfg.JwtToken, FlagJwtToken, jwtToken, "Токен проверки аутентификации пользователей")

	if err := fs.Parse(args); err != nil {
		return nil, fmt.Errorf("ошибка при парсинге флагов: %w", err)
	}

	return cfg, nil
}

// NewClient парсинг конфига сервера (приоритет по убыванию: флаги, env, default)
func NewClient(args []string, lookupFunc EnvLookupFunc) (*Client, error) {
	cfg := &Client{}

	fs := flag.NewFlagSet("app", flag.ContinueOnError)

	// Получаем данные из env или их дефолтные значения
	userAddress := getEnvString(lookupFunc, EnvAddress, DefaultAddress)

	// Получаем данные из флагов
	fs.StringVar(&cfg.Address, FlagAddress, userAddress, "Адрес системы программы лояльности")

	if err := fs.Parse(args); err != nil {
		return nil, fmt.Errorf("ошибка при парсинге флагов: %w", err)
	}

	return cfg, nil
}
