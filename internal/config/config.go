// Конфигурация сервисов.
package config

import (
	"flag"
	"fmt"
)

const (
	// Адрес сервера (gRPC)
	FlagAddress    = "addr"
	EnvAddress     = "ADDRESS"
	DefaultAddress = "localhost:50051"

	// Адрес подключения к базе данных
	FlagDatabaseURI    = "db"
	EnvDatabaseURI     = "DATABASE_URI"
	DefaultDatabaseURI = ""

	// Токен проверки аутентификации
	FlagJwtToken    = "jwt"
	EnvJwtToken     = "JWT_TOKEN"
	DefaultJwtToken = ""

	// TLS сертификат сервера
	FlagServerTLSCert    = "srv-tls-cert"
	EnvServerTLSCert     = "SERVER_TLS_CERT"
	DefaultServerTLSCert = "server.crt"

	// TLS ключ сервера
	FlagServerTLSKey    = "srv-tls-key"
	EnvServerTLSKey     = "SERVER_TLS_KEY"
	DefaultServerTLSKey = "server.key"

	// TLS сертификат сервера
	FlagClientTLSCert    = "client-tls-cert"
	EnvClientTLSCert     = "CLIENT_TLS_CERT"
	DefaultClientTLSCert = "ca.crt"
)

// Server - Структура конфигурации для сервера
type Server struct {
	Address     string
	DatabaseURI string
	JwtToken    string
	TLSCert     string
	TLSKey      string
}

// Client - Структура конфигурации для клиента
type Client struct {
	Address string
	TLSCert string
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
	certTLS := getEnvString(lookupFunc, EnvServerTLSCert, DefaultServerTLSCert)
	keyTLS := getEnvString(lookupFunc, EnvServerTLSKey, DefaultServerTLSKey)

	// Получаем данные из флагов
	fs.StringVar(&cfg.Address, FlagAddress, userAddress, "Адрес системы программы лояльности")
	fs.StringVar(&cfg.DatabaseURI, FlagDatabaseURI, databaseURI, "Адрес подключения к базе данных")
	fs.StringVar(&cfg.JwtToken, FlagJwtToken, jwtToken, "Токен проверки аутентификации пользователей")
	fs.StringVar(&cfg.TLSCert, FlagServerTLSCert, certTLS, "TLS сертификат для сервера")
	fs.StringVar(&cfg.TLSKey, FlagServerTLSKey, keyTLS, "TLS ключ для сервера")

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
	certTLS := getEnvString(lookupFunc, EnvClientTLSCert, DefaultClientTLSCert)

	// Получаем данные из флагов
	fs.StringVar(&cfg.Address, FlagAddress, userAddress, "Адрес системы программы лояльности")
	fs.StringVar(&cfg.TLSCert, FlagClientTLSCert, certTLS, "TLS сертификат для клиента")

	if err := fs.Parse(args); err != nil {
		return nil, fmt.Errorf("ошибка при парсинге флагов: %w", err)
	}

	return cfg, nil
}
