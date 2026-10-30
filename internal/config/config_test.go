package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestNewServer Тестирование парсинга конфига сервера
func TestNewServer(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		env     map[string]string
		success bool
		want    *Server
	}{
		{
			name:    "positive - Пустой конфиг",
			args:    make([]string, 0),
			env:     make(map[string]string),
			success: true,
			want: &Server{
				Address:     DefaultAddress,
				DatabaseURI: DefaultDatabaseURI,
				JwtToken:    DefaultJwtToken,
				TLSCert:     DefaultServerTLSCert,
				TLSKey:      DefaultServerTLSKey,
			},
		},
		{
			name: "positive - Только env",
			args: make([]string, 0),
			env: map[string]string{
				EnvAddress:     "localhost:1111",
				EnvDatabaseURI: "database",
				EnvJwtToken:    "env token",
			},
			success: true,
			want: &Server{
				Address:     "localhost:1111",
				DatabaseURI: "database",
				JwtToken:    "env token",
				TLSCert:     DefaultServerTLSCert,
				TLSKey:      DefaultServerTLSKey,
			},
		},
		{
			name: "positive - Только флаги",
			args: []string{
				"-" + FlagAddress, "localhost:3333",
				"-" + FlagDatabaseURI, "different_database",
				"-" + FlagJwtToken, "flag token",
			},
			env:     make(map[string]string),
			success: true,
			want: &Server{
				Address:     "localhost:3333",
				DatabaseURI: "different_database",
				JwtToken:    "flag token",
				TLSCert:     DefaultServerTLSCert,
				TLSKey:      DefaultServerTLSKey,
			},
		},
		{
			name: "positive - флаги поверх env",
			args: []string{
				"-" + FlagAddress, "localhost:3333",
				"-" + FlagDatabaseURI, "different_database",
				"-" + FlagJwtToken, "flag token",
			},
			env: map[string]string{
				EnvAddress:     "localhost:1111",
				EnvDatabaseURI: "database",
				EnvJwtToken:    "env token",
			},
			success: true,
			want: &Server{
				Address:     "localhost:3333",
				DatabaseURI: "different_database",
				JwtToken:    "flag token",
				TLSCert:     DefaultServerTLSCert,
				TLSKey:      DefaultServerTLSKey,
			},
		},
		{
			name: "negative - Ошибка при парсинге флагов",
			args: []string{
				"-" + FlagAddress,
			},
			env:     make(map[string]string),
			success: false,
			want:    nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			lookupFunc := func(name string) (string, bool) {
				value, ok := test.env[name]
				return value, ok
			}

			cfg, err := NewServer(test.args, lookupFunc)
			if test.success {
				assert.NoError(t, err)
				assert.Equal(t, test.want, cfg)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

// TestNewClient Тестирование парсинга конфига клиента
func TestNewClient(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		env     map[string]string
		success bool
		want    *Client
	}{
		{
			name:    "positive - По умолчанию",
			args:    make([]string, 0),
			env:     make(map[string]string),
			success: true,
			want: &Client{
				Address: DefaultAddress,
				TLSCert: DefaultClientTLSCert,
			},
		},
		{
			name:    "positive - Флаги",
			args:    []string{"-" + FlagAddress, "localhost:9090"},
			env:     make(map[string]string),
			success: true,
			want: &Client{
				Address: "localhost:9090",
				TLSCert: DefaultClientTLSCert,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			lookupFunc := func(name string) (string, bool) {
				value, ok := test.env[name]
				return value, ok
			}

			cfg, err := NewClient(test.args, lookupFunc)
			if test.success {
				assert.NoError(t, err)
				assert.Equal(t, test.want, cfg)
			} else {
				assert.Error(t, err)
			}
		})
	}
}
